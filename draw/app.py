# pip install Flask
# pdf to png requires ghostscript `pdftoppm` and ImageMagick `convert`
import base64
import io
import json
import logging
import os
import sqlite3
import subprocess
import time

from flask import Flask, render_template, request, g, url_for
# pip install python-memcached
#import memcache
memcache = None

from . import cache
from . import demorace
from . import draw
ElectionPrinter = draw.ElectionPrinter

app = Flask(__name__)
if os.getenv("BFLASK_CONF"):
    app.config.from_envvar("BFLASK_CONF")

if (os.getenv('SERVER_SOFTWARE') or '').startswith('gunicorn') and (__name__ != '__main__'):
    gunicorn_logger = logging.getLogger('gunicorn.error')
    app.logger.handlers = gunicorn_logger.handlers
    app.logger.setLevel(gunicorn_logger.level)

draw.logger = app.logger

_cache = None

def mc():
    # use memcached if installed?
    if memcache is not None:
        mc = getattr(g, '_memcache', None)
        if mc is None:
            mc = g._memcache = memcache.Client(['127.0.0.1:11211'], debug=0)
        return mc
    # use local built in cache
    global _cache
    if _cache is None:
        _cache = cache.Cache()
    return _cache

# TODO: ownership, ACLs, any kind of security at all
current_schema = [
    "CREATE TABLE IF NOT EXISTS elections (data TEXT, meta TEXT)", # use builtin ROWID
    "CREATE TABLE IF NOT EXISTS migrations (mid INT PRIMARY KEY) WITHOUT ROWID",
]

# never delete a migration or change its int key
migrations = [
    (1, ["CREATE TABLE IF NOT EXISTS migrations (mid INT PRIMARY KEY) WITHOUT ROWID","ALTER TABLE elections ADD COLUMN meta TEXT"])
]

def db():
    conn = getattr(g, '_database', None)
    if conn is None:
        sqlite3path = os.getenv('BALLOTSTUDIO_SQLITE') or 'ballotstudio.sqlite'
        conn = sqlite3.connect(sqlite3path)
        c = conn.cursor()
        try:
            c.execute("SELECT mid FROM migrations")
            migrations_done = set([row[0] for row in c.fetchall()])
        except:
            migrations_done = set()
        try:
            c.execute("SELECT COUNT(*) FROM elections")
            row = c.fetchone()
            num_elections = row and row[0]
        except:
            num_elections = 0
        if not num_elections:
            # new db
            for stmt in current_schema:
                c.execute(stmt)
            # mark all migrations as applied
            c.executemany("INSERT INTO migrations (mid) VALUES (?)", [(mig[0],) for mig in migrations])
        else:
            migs_applied = []
            for mig in migrations:
                mid = mig[0]
                if mid not in migrations_done:
                    for stmt in mig[1]:
                        c.execute(stmt)
                    migs_applied.append( (mid,) )
            c.executemany("INSERT INTO migrations (mid) VALUES (?)", migs_applied)
        conn.commit()
        demo = _getelection(1, conn)
        if not demo:
            _putelection(demorace.ElectionReport, 1, conn)
        g._database = conn
    return conn

def putelection(ob, itemid=None):
    conn = db()
    return _putelection(ob, itemid, conn)

def _putelection(ob, itemid, conn):
    c = conn.cursor()
    if itemid:
        itemid = int(itemid)
        # TODO: why isn't sqlite "ON CONFLICT ..." syntax working? sqlite3.sqlite_version === '3.22.0'
        #c.execute("INSERT INTO elections (ROWID, data) VALUES (?, ?) ON CONFLICT (ROWID) DO UPDATE SET data = EXCLUDED.data", (itemid, json.dumps(ob)))
        c.execute("INSERT OR REPLACE INTO elections (ROWID, data) VALUES (?, ?)", (itemid, json.dumps(ob)))
        conn.commit()
        c.close()
        return itemid
    else:
        c.execute("INSERT INTO elections (data) VALUES (?)", (json.dumps(ob),))
        conn.commit()
        itemid = c.lastrowid
        app.logger.info('new election %s', itemid)
        c.close()
        return itemid


def getelection(itemid):
    conn = db()
    return _getelection(itemid, conn)

def _getelection(itemid, conn):
    c = conn.cursor()
    c.execute("SELECT data FROM elections WHERE ROWID = ?", (int(itemid),))
    row = c.fetchone()
    c.close()
    if not row:
        return None
    return json.loads(row[0])


# pdf bytes in, png bytes out
# bounces of two subprocess calls (all pipes, no disk) 'pdftoppm' and 'convert'
def pdfToPng(pdfbytes):
    result = subprocess.run(['pdftoppm'], input=pdfbytes, stdout=subprocess.PIPE)
    result.check_returncode()
    ppmbytes = result.stdout
    result = subprocess.run(['convert', 'ppm:-', 'png:-'], input=ppmbytes, stdout=subprocess.PIPE)
    result.check_returncode()
    return result.stdout # png bytes


@app.route('/')
def home():
    return render_template('index.html', electionid="", urls=_election_urls(), prefix=request.environ.get('SCRIPT_NAME','').rstrip('/'))

@app.route('/edit/<int:electionid>')
def edit(electionid):
    ctx = {
        "electionid":electionid,
        "urls":_election_urls(electionid),
        "prefix":request.environ.get('SCRIPT_NAME','').rstrip('/'),
    }
    return render_template('index.html', **ctx)

@app.route('/draw', methods=['POST'])
def drawHandler():
    if request.content_type != 'application/json':
        return 'bad content-type', 400
    er = request.get_json()
    elections = er.get('Election', [])
    el = elections[0]
    ep = ElectionPrinter(er, el)
    pdfbytes = io.BytesIO()
    ep.draw(outfile=pdfbytes)
    pdfbytes = pdfbytes.getvalue()
    if len(pdfbytes) == 0:
        app.logger.warning('zero byte pdf /draw')
    bothob = {
        'pdfb64': base64.b64encode(pdfbytes).decode(),
        'bubbles': ep.getBubbles(),
    }
    if request.args.get('both'):
        return bothob, 200
    if request.args.get('bubbles'):
        itemid = request.args.get('i')
        if not itemid:
            itemid = '{:08x}'.format(int(time.time()-1588036000))
        mc().set(itemid, bothob, time=3600)
        return {'bubbles':ep.getBubbles(),'item':itemid}, 200
    # otherwise just pdf
    return pdfbytes, 200, {"Content-Type":"application/pdf"}

@app.route('/item')
def itemHandler():
    itemid = request.args.get('i')
    if not itemid:
        return '', 404
    bothob = mc().get(itemid)
    if not bothob:
        return '', 404
    if request.args.get('both'):
        return bothob, 200
    if request.args.get('bubbles'):
        return {'bubbles':bothob['bubbles'],'item':itemid}, 200
    # otherwise just pdf
    pdfbytes = base64.b64decode(bothob['pdfb64'])
    return pdfbytes, 200, {"Content-Type":"application/pdf"}

def _election_urls(itemid=None):
    if itemid is not None:
        out = {
            "itemid":itemid,
            "url":url_for('elections', itemid=itemid),
            "pdf":url_for('election_pdf', itemid=itemid),
            "bubbles":url_for('election_bubblejson', itemid=itemid),
            "edit":url_for('edit', electionid=itemid),
            "post":url_for('elections', itemid=itemid),
        }
    else:
        out = {
            "post":url_for('putNewElection'),
        }
    out['staticroot'] = request.environ.get('SCRIPT_NAME','').rstrip('/') + '/static'
    return out

@app.route("/election", methods=['POST'])
def putNewElection():
    er = request.get_json()
    bothob = _er_bothob(er)
    itemid = putelection(er)
    mc().set('e{}'.format(itemid), bothob, time=3600)
    return _election_urls(itemid), 200

@app.route("/election/<int:itemid>", methods=['GET', 'POST'])
def elections(itemid):
    if request.method == 'POST':
        er = request.get_json()
        bothob = _er_bothob(er)
        mc().set('e{}'.format(itemid), bothob, time=3600)
        itemid = putelection(er, itemid)
        return _election_urls(itemid), 200
    elif request.method == 'GET':
        er = getelection(itemid)
        if er is None:
            return {'error': 'no election {}'.format(itemid)}, 404
        return er, 200
    return 'nope', 400

def _er_bothob(er):
    elections = er.get('Election', [])
    el = elections[0]
    ep = ElectionPrinter(er, el)
    pdfbytes = io.BytesIO()
    ep.draw(outfile=pdfbytes)
    pdfbytes = pdfbytes.getvalue()
    return {'pdf':pdfbytes, 'bubbles':ep.getBubbles()}

def _bothob_core(itemid):
    er = getelection(itemid)
    if er is None:
        return {'error': 'no election {}'.format(itemid)}, 404

    cachekey = 'e{}'.format(itemid)
    bothob = mc().get(cachekey)
    if not bothob:
        bothob = _er_bothob(er)
        mc().set(cachekey, bothob, time=3600)
    return bothob

@app.route("/election/<int:itemid>.pdf")
def election_pdf(itemid):
    bothob = _bothob_core(itemid)
    pdfbytes = bothob['pdf']
    return pdfbytes, 200, {"Content-Type":"application/pdf"}

@app.route("/election/<int:itemid>.png")
def election_png(itemid):
    bothob = _bothob_core(itemid)
    pngbytes = bothob.get('png')
    if pngbytes is None:
        pngbytes = pdfToPng(bothob['pdf'])
        bothob['png'] = pngbytes
    return pngbytes, 200, {"Content-Type":"image/png"}

@app.route("/election/<int:itemid>_bubbles.json")
def election_bubblejson(itemid):
    bothob = _bothob_core(itemid)
    return bothob['bubbles'], 200 # implicit dict-to-json return

@app.route("/election/<int:electionid>/scan")
def scanform(electionid):
    if request.method == 'POST':
        return 'wat', 500
    elif request.method == 'GET':
        return render_template('scanform.html', electionid=electionid, urls=_election_urls(electionid))
    return 'nope', 400

@app.route("/scan/<int:electionid>", methods=['POST'])
def scanproxy(electionid):
    return 'wat', 500
