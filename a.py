#!/usr/bin/env python3
# -*- mode: Python; coding: utf-8 -*-
#

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
mainfontname = 'Liberation-Sans'
fontpath = '/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf'

ftt = fontTools.ttLib.TTFont(fontpath)
# do some font metrics about the size of the capital letters
caps = [chr(x) for x in range(ord('A'), ord('Z')+1)]
caps.remove('Q') # outlier descender
glyfminmax = [(ftt['glyf'][glyc].yMax, ftt['glyf'][glyc].yMin, glyc) for glyc in caps]
gmaxes = [x[0] for x in glyfminmax]
gmins = [x[1] for x in glyfminmax]
capmin = statistics.median(gmins)
capmax = statistics.median(gmaxes)
capHeightPerPt = (capmax - capmin) / 2048

pointsize = 26
capHeight = capHeightPerPt * pointsize
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

lfont = TTFont(mainfontname, fontpath)
pdfmetrics.registerFont(lfont)


c = canvas.Canvas('/tmp/a.pdf', pagesize=letter) # pageCompression=1
widthpt, heightpt = letter

nowstrFontSize = 12
dtw = pdfmetrics.stringWidth(nowstr, mainfontname, nowstrFontSize)
c.setFont(mainfontname, nowstrFontSize)
c.drawString(widthpt - 0.5*inch - dtw, heightpt - 0.5*inch - nowstrFontSize*1.2, nowstr)

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
    c.setFont(mainfontname, pointsize)
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
