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


fonts = {}

for fpath in glob.glob('/usr/share/fonts/truetype/liberation/*.ttf'):
    xf = Bfont(fpath)
    fonts[xf.name] = xf

print('fonts: ' + ', '.join([repr(n) for n in fonts.keys()]))


class Settings:
    def __init__(self):
        self.titleFontName = 'Liberation Sans Bold'
        self.titleFontSize = 12
        self.titleBGColor = (.85, .85, .85)
        self.titleLeading = self.titleFontSize * 1.4
        self.subtitleFontName = 'Liberation Sans Bold'
        self.subtitleFontSize = 12
        self.subtitleBGColor = ()
        self.subtitleLeading = self.subtitleFontSize * 1.4
        self.candidateFontName = 'Liberation Sans Bold'
        self.candidateFontSize = 12
        self.candidateLeading = 13
        self.candsubFontName = 'Liberation Sans'
        self.candsubFontSize = 12
        self.candsubLeading = 13
        self.bubbleLeftPad = 0.1 * inch
        self.bubbleRightPad = 0.1 * inch
        self.bubbleWidth = 8 * mm
        self.bubbleMaxHeight = 3 * mm


gs = Settings()


class Choice:
    def __init__(self, name, subtext=None):
        self.name = name
        self.subtext = subtext
        # TODO: measure text for box width & wrap
        # TODO: wrap with optional max-5% squish instead of wrap
    def height(self):
        # TODO: multiline for name and subtext
        y = 0
        ypos = y - gs.candidateLeading
        if self.subtext:
            ypos -= gs.candsubLeading
        return -1*(ypos - (0.1 * inch))
    def draw(self, c, x, y, width=(7.5/2)*inch - 1):
        # x,y is a top,left of a box to draw bubble and text into
        capHeight = fonts[gs.candidateFontName].capHeightPerPt * gs.candidateFontSize
        bubbleHeight = min(3*mm, capHeight)
        bubbleYShim = (capHeight - bubbleHeight) / 2.0
        bubbleBottom = y - gs.candidateFontSize + bubbleYShim
        c.setStrokeColorRGB(0,0,0)
        c.setLineWidth(1)
        c.setFillColorRGB(1,1,1)
        c.roundRect(x + gs.bubbleLeftPad, bubbleBottom, gs.bubbleWidth, bubbleHeight, radius=bubbleHeight/2)
        textx = x + gs.bubbleLeftPad + gs.bubbleWidth + gs.bubbleRightPad
        # TODO: assumes one line
        c.setFillColorRGB(0,0,0)
        txto = c.beginText(textx, y - gs.candidateFontSize)
        txto.setFont(gs.candidateFontName, gs.candidateFontSize, gs.candidateLeading)
        txto.textLines(self.name)
        c.drawText(txto)
        ypos = y - gs.candidateLeading
        if self.subtext:
            txto = c.beginText(textx, ypos - gs.candsubFontSize)
            txto.setFont(gs.candsubFontName, gs.candsubFontSize, leading=gs.candsubLeading)
            txto.textLines(self.subtext)
            c.drawText(txto)
            ypos -= gs.candsubLeading
        # separator line
        c.setStrokeColorRGB(0,0,0)
        c.setLineWidth(0.25)
        sepy = ypos - (0.1 * inch)
        c.line(textx, sepy, x+width, sepy)
        return
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


class Contest:
    def __init__(self, name, title=None, subtitle=None, choices=None):
        self.name = name
        self.title = title or name
        self.subtitle = subtitle
        self.choices = choices
        self._choices_height = None
        self._height = None
        self._maxChoiceHeight = None
    def height(self):
        if self._height is None:
            choices = self.choices or []
            self._maxChoiceHeight = max([x.height() for x in choices])
            ch = self._maxChoiceHeight * len(choices)
            ch += 4 # top and bottom border
            ch += gs.titleLeading + gs.subtitleLeading
            ch += 0.1 * inch # header-choice gap
            ch += 0.1 * inch # bottom padding
            self._height = ch
        return self._height
    def draw(self, c, x, y, width=(7.5/2)*inch - 1):
        # x,y is a top,left
        # top border
        c.setStrokeColorRGB(0,0,0)
        c.setLineWidth(3)
        c.line(x, y, x + width, y)
        # left border
        c.setLineWidth(1)
        c.line(x, y, x, y-self.height())
        # bottom border
        c.line(x, y-self.height(), x+width, y-self.height())

        pos = y - 1.5
        # title
        c.setStrokeColorRGB(*gs.titleBGColor)
        c.setFillColorRGB(*gs.titleBGColor)
        c.rect(x + 1, pos - gs.titleLeading, width - 1, gs.titleLeading, fill=1)
        c.setFillColorRGB(0,0,0)
        c.setStrokeColorRGB(0,0,0)
        txto = c.beginText(x + 1 + (0.1 * inch), pos - gs.titleFontSize)
        txto.setFont(gs.titleFontName, gs.titleFontSize)
        txto.textLines(self.title)
        c.drawText(txto)
        pos -= gs.titleLeading
        # subtitle
        c.setStrokeColorCMYK(.1,0,0,0)
        c.setFillColorCMYK(.1,0,0,0)
        c.rect(x + 1, pos - gs.subtitleLeading, width - 1, gs.subtitleLeading, fill=1)
        c.setFillColorRGB(0,0,0)
        c.setStrokeColorRGB(0,0,0)
        txto = c.beginText(x + 1 + (0.1 * inch), pos - gs.subtitleFontSize)
        txto.setFont(gs.subtitleFontName, gs.subtitleFontSize)
        txto.textLines(self.subtitle)
        c.drawText(txto)
        pos -= gs.subtitleLeading
        pos -= 0.1 * inch # header-choice gap
        c.setFillColorRGB(0,0,0)
        c.setStrokeColorRGB(0,0,0)
        choices = self.choices or []
        for ch in choices:
            ch.draw(c, x + 1, pos, width=width - 1)
            pos -= self._maxChoiceHeight
        return


c = canvas.Canvas('/tmp/a.pdf', pagesize=letter) # pageCompression=1
widthpt, heightpt = letter

nowstrFontSize = 12
mainfontname = 'Liberation Sans'
dtw = pdfmetrics.stringWidth(nowstr, mainfontname, nowstrFontSize)
c.setFont(mainfontname, nowstrFontSize)
c.drawString(widthpt - 0.5*inch - dtw, heightpt - 0.5*inch - nowstrFontSize*1.2, nowstr)
c.setTitle('ballot test ' + nowstr)


candidateNames = [
    'Alice Argyle',
    'Bob Brocade',
    'Çandidate Ñame',
    'Dorian Duck',
    'Elaine Entwhistle',
    #'ÇÅÌy º Ċ úü  Ã ° ~', # TODO: more i18n testing
]

#choices = [Choice(nameText) for nameText in candidateNames]
choices = [
    Choice('Alice Argyle', 'Anklebiter Assembly'),
    Choice('Bob Brocade', 'Boring Board'),
    Choice('Çandidate Ñame', 'Cowardly Coalition'),
    Choice('Dorian Duck', 'Dumb Department'),
    Choice('Elaine Entwhistle', 'Electable Entertainers'),
]

therace = Contest('Everything', 'The Race for Everything', 'Choose as many as you like', choices)

# maxChoiceHeight = max([x.height() for x in choices])
# pos = heightpt - 0.5*inch # - pointsize * 1.2
# for ch in choices:
#     ch.draw(c, 0.5*inch, pos)
#     pos -= maxChoiceHeight

therace.draw(c, 0.5 * inch, heightpt - 0.5*inch)

c.showPage()
c.save()
