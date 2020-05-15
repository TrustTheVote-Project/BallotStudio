import base64
import io
import json
import logging
import os
import sqlite3
import time

# pip install Flask python-memcached
from flask import Flask, render_template, request, g, url_for
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

def db():
    conn = getattr(g, '_database', None)
    if conn is None:
        sqlite3path = os.getenv('BALLOTSTUDIO_SQLITE') or 'ballotstudio.sqlite'
        conn = sqlite3.connect(sqlite3path)
        c = conn.cursor()
        # TODO: ownership, ACLs, any kind of security at all
        c.execute("CREATE TABLE IF NOT EXISTS elections (data TEXT)") # use builtin ROWID
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
    itemid = request.args.get('i')
    if not itemid:
        itemid = '{:08x}'.format(int(time.time()-1588036000))
    bothob = {
        'pdfb64': base64.b64encode(pdfbytes.getvalue()).decode(),
        'bubbles': ep.getBubbles(),
    }
    if request.args.get('both'):
        return bothob, 200
    if request.args.get('bubbles'):
        mc().set(itemid, bothob, time=3600)
        return {'bubbles':ep.getBubbles(),'item':itemid}, 200
    # otherwise just pdf
    return pdfbytes.getvalue(), 200, {"Content-Type":"application/pdf"}

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

@app.route("/election/<int:itemid>_bubbles.json")
def election_bubblejson(itemid):
    bothob = _bothob_core(itemid)
    return bothob['bubbles'], 200 # implicit dict-to-json return
