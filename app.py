import base64
import io
import time

# pip install Flask python-memcached
from flask import Flask, render_template, request, g
import memcache

from . import draw
ElectionPrinter = draw.ElectionPrinter

app = Flask(__name__)

def mc():
    mc = getattr(g, '_memcache', None)
    if mc is None:
        mc = g._memcache = memcache.Client(['127.0.0.1:11211'], debug=0)
    return mc

@app.route('/')
def home():
    return render_template('index.html')

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
