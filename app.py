import base64
import io
import json
import os
import sqlite3
import time

# pip install Flask python-memcached
from flask import Flask, render_template, request, g
#import memcache
memcache = None

from . import cache
from . import demorace
from . import draw
ElectionPrinter = draw.ElectionPrinter

app = Flask(__name__)

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
        _putelection(demorace.ElectionReport, 1, conn)
        g._database = conn
    return conn

def putelection(ob, itemid=None):
    conn = db()
    _putelection(ob, itemid, conn)

def _putelection(ob, itemid, conn):
    c = conn.cursor()
    if itemid:
        itemid = int(itemid)
        # TODO: why isn't sqlite "ON CONFLICT ..." syntax working? sqlite3.sqlite_version === '3.22.0'
        #c.execute("INSERT INTO elections (ROWID, data) VALUES (?, ?) ON CONFLICT (ROWID) DO UPDATE SET data = EXCLUDED.data", (itemid, json.dumps(ob)))
        c.execute("INSERT OR REPLACE INTO elections (ROWID, data) VALUES (?, ?)", (itemid, json.dumps(ob)))
        c.close()
        conn.commit()
        return itemid
    else:
        c.execute("INSERT INTO elections (data) VALUES (?)", (json.dumps(ob),))
        itemid = c.lastrowid
        c.close()
        conn.commit()
        return itemid


def getelection(itemid):
    conn = db()
    c = conn.cursor()
    c.execute("SELECT data FROM elections WHERE ROWID = ?", (int(itemid),))
    row = c.fetchone()
    c.close()
    return json.loads(row[0])


@app.route('/')
def home():
    return render_template('index.html', electionid="")

@app.route('/edit/<int:electionid>')
def edit(electionid):
    return render_template('index.html', electionid=electionid)

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

@app.route("/election", methods=['POST'])
def putNewElection():
    er = request.get_json()
    # todo: validate content
    itemid = putelection(er)
    return {"itemid":itemid}, 200

@app.route("/election/<int:itemid>", methods=['GET', 'POST'])
def elections(itemid):
    if request.method == 'POST':
        er = request.get_json()
        # todo: validate content
        itemid = putelection(er, itemid)
        return {"itemid":itemid}, 200
    elif request.method == 'GET':
        er = getelection(itemid)
        if er is None:
            return {'error': 'no item {}'.format(itemid)}, 404
        return er, 200
    return 'nope', 400
