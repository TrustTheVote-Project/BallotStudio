#!/usr/bin/env python3
# -*- mode: Python; coding: utf-8 -*-
#

from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import letter, A4
from reportlab.lib.units import inch, mm, cm
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont

def bubble(c, width=12*mm, height=4*mm, r=1*mm):
    pth = c.beginPath()
    pth.roundRect(0, 0, width, height, r)
    return pth

import fontTools.ttLib
ftt = fontTools.ttLib.TTFont('/usr/share/fonts/truetype/liberation/LiberationSerif-Regular.ttf')
pointsize = 26
Xheight = (ftt['glyf']['X'].yMax - ftt['glyf']['X'].yMin) * pointsize / 2048


lfont = TTFont('Liberation-Serif', '/usr/share/fonts/truetype/liberation/LiberationSerif-Regular.ttf')
pdfmetrics.registerFont(lfont)

#mainfontname = 'Times-Roman'
mainfontname = 'Liberation-Serif'

c = canvas.Canvas('/tmp/a.pdf', pagesize=letter) # pageCompression=1
widthpt, heightpt = letter
#c.rect(0.5*inch, heightpt - inch, 11 * mm, 4 * mm)
#bpath = bubble(c)
#c.drawPath(bpath)
# bubble:
c.setStrokeColorRGB(0,0,0)
c.setFillColorRGB(1,1,1)
topliney = heightpt - 0.5*inch - pointsize*1.2
bubbleHeight = min(3*mm, Xheight)
bubbleYShim = (Xheight - bubbleHeight) / 2.0
c.roundRect(0.5*inch, topliney + bubbleYShim, 12*mm, bubbleHeight, radius=bubbleHeight/2)

if False:
    c.setStrokeColorRGB(1,1,1)
    c.setFillColorRGB(0,0.3,0)
    c.rect(0.5*inch + 13*mm, topliney, 1*mm, pointsize, stroke=0, fill=1)
    c.rect(0.5*inch + 14.1*mm, topliney, 1*mm, Xheight, stroke=0, fill=1)

c.setFont(mainfontname, pointsize)
c.setFillColorRGB(0,0,0)
line = "ÇÅÌy º Ċ úü  Ã °  yo sup"
text = line + "\nÇandidate Ñame"
#c.drawString(0.5*inch + 15*mm,0.5*inch,line)
txto = c.beginText(0.5*inch + 15*mm,topliney + pointsize*1.2)
txto.textLines(text)
c.drawText(txto)

c.showPage()
c.save()

print(lfont.face.extractInfo(charInfo=0))
#trfont = pdfmetrics.getFont(mainfontname)
#trfont.stringWidth("ÇÅÌy º Ċ úü  Ã °  yo sup")
