#!/usr/bin/env python3
# -*- mode: Python; coding: utf-8 -*-
#

import glob
import time
import statistics

import fontTools.ttLib
from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import letter, A4
from reportlab.lib.units import inch, mm, cm
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont

nowstr = 'generated ' + time.strftime('%Y-%m-%d %H:%M:%S UTC', time.gmtime())

def bubble(c, width=12*mm, height=4*mm, r=1*mm):
    pth = c.beginPath()
    pth.roundRect(0, 0, width, height, r)
    return pth

#mainfontname = 'Times-Roman'
#mainfontname = 'Liberation-Serif'
#fontpath = '/usr/share/fonts/truetype/liberation/LiberationSerif-Regular.ttf'
mainfontname = 'Liberation Sans'
#fontpath = '/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf'
#fontpathbold = '/usr/share/fonts/truetype/liberation/LiberationSans-Bold.ttf'
boldfontname = 'Liberation Sans Bold'


class Bfont:
    def __init__(self, path, name=None):
        self.name = name
        self.path = path
        self.capHeightPerPt = None
        self._measureCapheight()
        lfont = TTFont(self.name, self.path)
        pdfmetrics.registerFont(lfont)

    def _measureCapheight(self):
        ftt = fontTools.ttLib.TTFont(self.path)
        caps = [chr(x) for x in range(ord('A'), ord('Z')+1)]
        caps.remove('Q') # outlier descender
        glyfminmax = [(ftt['glyf'][glyc].yMax, ftt['glyf'][glyc].yMin, glyc) for glyc in caps]
        gmaxes = [x[0] for x in glyfminmax]
        gmins = [x[1] for x in glyfminmax]
        capmin = statistics.median(gmins)
        capmax = statistics.median(gmaxes)
        self.capHeightPerPt = (capmax - capmin) / 2048
        if self.name is None:
            for xn in ftt['name'].names:
                if xn.nameID == 4:
                    self.name = xn.toUnicode()
                    break

# ftt = fontTools.ttLib.TTFont(fontpath)
# # do some font metrics about the size of the capital letters
# caps = [chr(x) for x in range(ord('A'), ord('Z')+1)]
# caps.remove('Q') # outlier descender
# glyfminmax = [(ftt['glyf'][glyc].yMax, ftt['glyf'][glyc].yMin, glyc) for glyc in caps]
# gmaxes = [x[0] for x in glyfminmax]
# gmins = [x[1] for x in glyfminmax]
# capmin = statistics.median(gmins)
# capmax = statistics.median(gmaxes)
# capHeightPerPt = (capmax - capmin) / 2048

fonts = {}

for fpath in glob.glob('/usr/share/fonts/truetype/liberation/*.ttf'):
    xf = Bfont(fpath)
    fonts[xf.name] = xf

print('fonts: ' + ', '.join([repr(n) for n in fonts.keys()]))

pointsize = 26
bf = fonts[mainfontname]
capHeight = bf.capHeightPerPt * pointsize
candsize = pointsize * 1.4
bubbleHeight = min(3*mm, capHeight)
bubbleYShim = (capHeight - bubbleHeight) / 2.0

candidateNames = [
    'Alice Argyle',
    'Bob Brocade',
    'Çandidate Ñame',
    'Dorian Duck',
    'Elaine Entwhistle',
    #'ÇÅÌy º Ċ úü  Ã ° ~', # TODO: more i18n testing
]

def Settings:
    def __init__(self):
        self.candidateFontName = 'Liberation Sans Bolt'
        self.candidateFontSize = 12
        self.candidateLeading = 13
        self.candsubFontName = 'Liberation Sans'
        self.candsubFontSize = 12
        self.candsubLeading = 13

class Choice:
    def __init__(self, name, subtext=None):
        self.name = name
        self.subtext = subtext
    def draw(self, c, x, y):
        pass
    def sepLineBelow(self, c, x, y):
        c.setStrokeColorRGB(0,0,0)
        choiceBoxHeight = 30 # TODO: calculate
        choiceBoxWidth = 200 # TODO: calculate/specify
        c.setLineWidth(0.25)
        c.line(x, y-choiceBoxHeight, choiceBoxWidth, 0)
    def _writeInLine(self, c):
        c.setDash([4,4])
        c.setLineWidth(0.5)
        c.setDash(None)

#lfont = TTFont(mainfontname, fontpath)
#pdfmetrics.registerFont(lfont)


c = canvas.Canvas('/tmp/a.pdf', pagesize=letter) # pageCompression=1
widthpt, heightpt = letter

nowstrFontSize = 12
dtw = pdfmetrics.stringWidth(nowstr, mainfontname, nowstrFontSize)
c.setFont(mainfontname, nowstrFontSize)
c.drawString(widthpt - 0.5*inch - dtw, heightpt - 0.5*inch - nowstrFontSize*1.2, nowstr)
c.setTitle('ballot test ' + nowstr)

def drawChoice(topliney, nameText):
    # TODO: candidate subtext (occupation, etc), multi-line capable
    # TODO: boxes, chartjunk, etc
    # TODO: multi-line nameText
    # TODO: RCV bubbles
    # bubble
    c.setStrokeColorRGB(0,0,0)
    c.setFillColorRGB(1,1,1)
    c.roundRect(0.5*inch, topliney + bubbleYShim, 12*mm, bubbleHeight, radius=bubbleHeight/2)

    if False:
        # debugging marks
        c.setStrokeColorRGB(1,1,1)
        c.setFillColorRGB(0,0.3,0)
        c.rect(0.5*inch + 13*mm, topliney, 1*mm, pointsize, stroke=0, fill=1)
        c.rect(0.5*inch + 14.1*mm, topliney, 1*mm, capHeight, stroke=0, fill=1)

    # candidate name
    c.setFont(boldfontname, pointsize)
    c.setFillColorRGB(0,0,0)
    txto = c.beginText(0.5*inch + 15*mm,topliney)
    txto.textLines(nameText)
    c.drawText(txto)
    return

pos = heightpt - 0.5*inch - pointsize * 1.2
for nameText in candidateNames:
    drawChoice(pos, nameText)
    pos -= candsize

c.showPage()
c.save()
