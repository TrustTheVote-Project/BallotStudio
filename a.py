#!/usr/bin/env python3
# -*- mode: Python; coding: utf-8 -*-
#

import glob
import json
import time
import statistics
import sys

import fontTools.ttLib
from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import letter, A4
from reportlab.lib.units import inch, mm, cm
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
#from reportlab.platypus import Paragraph
#from reportlab.lib.units import ParagraphStyle

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

#print('fonts: ' + ', '.join([repr(n) for n in fonts.keys()]))


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
        self.columnMargin = 0.1 * inch


gs = Settings()


class Choice:
    def __init__(self, name, subtext=None):
        self.name = name
        self.subtext = subtext
        # TODO: measure text for box width & wrap. see reportlab.platypus.Paragraph
        # TODO: wrap with optional max-5% squish instead of wrap
        self._bubbleCoords = None
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
        self._bubbleCoords = (x + gs.bubbleLeftPad, bubbleBottom, gs.bubbleWidth, bubbleHeight)
        c.roundRect(*self._bubbleCoords, radius=bubbleHeight/2)
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
        height = self.height()

        pos = y - 1.5
        # title
        c.setStrokeColorRGB(*gs.titleBGColor)
        c.setFillColorRGB(*gs.titleBGColor)
        c.rect(x, pos - gs.titleLeading, width, gs.titleLeading, fill=1, stroke=0)
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
        c.rect(x, pos - gs.subtitleLeading, width, gs.subtitleLeading, fill=1, stroke=0)
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

        # top border
        c.setStrokeColorRGB(0,0,0)
        c.setLineWidth(3)
        c.line(x-0.5, y, x + width, y) # -0.5 caps left border 1.0pt line
        # left border and bottom border
        c.setLineWidth(1)
        path = c.beginPath()
        path.moveTo(x, y)
        path.lineTo(x, y-height)
        path.lineTo(x+width, y-height)
        c.drawPath(path, stroke=1)
        return
    def getBubbles(self):
        choices = self.choices or []
        return {ch.name:ch._bubbleCoords for ch in choices}


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

raceZ = Contest(
    'Nothing', 'The Race To The Bottom', 'Vote for one',
    [
        Choice('Zaphod', "He's just this guy, you know?"),
        Choice('Zardoz', 'There can be only one'),
        Choice('Zod', 'Kneel'),
    ],
)

headDwarfRace = Contest(
    'Head Dwarf',
    'Head Dwarf',
    'Vote for one',
    [
        Choice('Sleepy'),
        Choice('Happy'),
        Choice('Dopey'),
        Choice('Grumpy'),
        Choice('Sneezy'),
        Choice('Bashful'),
        Choice('Doc'),
    ],
)

# maxChoiceHeight = max([x.height() for x in choices])
# pos = heightpt - 0.5*inch # - pointsize * 1.2
# for ch in choices:
#     ch.draw(c, 0.5*inch, pos)
#     pos -= maxChoiceHeight

# case-insensitive dict.pop()
def cp(key, d, exc=True):
    if key in d:
        return d.pop(key)
    kl = key.lower()
    for dk in d.keys():
        dkl = dk.lower()
        if dkl == kl:
            return d.pop(dk)
    if exc:
        raise KeyError(key)
    return None

# multi-option case-insensitive dict.pop()
def ocp(d, *keys, default=None, exc=True):
    for k in keys:
        try:
            return cp(k, d)
        except:
            pass
    if default is not None:
        return default
    if exc:
        raise KeyError(key)
    return default

def maybeToDict(o):
    if isinstance(o, list):
        return [maybeToDict(ox) for ox in o]
    elif isinstance(o, dict):
        return {k:maybeToDict(v) for k,v in o.items()}
    elif hasattr(o, 'toDict'):
        return o.toDict()
    return o

class Builder:
    def toDict(self):
        d = {}
        if hasattr(self, '_type'):
            d['@type'] = self._type
        if hasattr(self, '_id'):
            d['@id'] = self._id
        for k,v in self.__dict__.items():
            if k[0] != '_' and not hasattr(v, '__call__'):
                d[k] = maybeToDict(v)

class Person(Builder):
    _fields = ()
    _type = "ElectionResults.Person"
    def __init__(self):
        pass
class ElectionReportBuilder:
    _type = "ElectionResults.ElectionReport"
    def __init__(self, **kwargs):
        # required
        self.Format = ocp(kwargs, "Format", default="summary-contest")
        self.GeneratedDate = ocp(kwargs, "GeneratedDate", "date", default=time.strftime("%Y-%m-%d %H:%M:%S %z", time.localtime()))
        self.Issuer = ocp(kwargs, "Issuer", default="bolson")
        self.IssuerAbbreviation = ocp(kwargs, "IssuerAbbreviation", default=self.Issuer)
        self.SequenceStart = int(ocp(kwargs, "SequenceStart", default=1))
        self.SequenceEnd = int(ocp(kwargs, "SequenceEnd", default=1))
        self.Status = ocp(kwargs, "Status", default="pre-election")
        self.VendorApplicationId = ocp(kwargs, "VendorApplicationId", default="BallotGen 0.0.1")
        # etc
        self.Election = []
        self.ExternalIdentifier = []
        self.Header = []
        self.IsTest = True
        self.Notes = ""
        self.Office = []
        self.OfficeGroup = []
        self.Party = []
        self.Person = []
        self.TestType = ""
    def Person(self, **kwargs):
        pass


ElectionReport = {
    # required fields
    "@type": "ElectionReport",
    "Format": "summary-contest",
    "GeneratedDate": time.strftime("%Y-%m-%d %H:%M:%S %z", time.localtime()),
    "Issuer": "bolson",
    "IssuerAbbreviation": "bolson",
    "SequenceStart": 1,
    "SequenceEnd": 1,
    "Status": "pre-election",
    "VendorApplicationId": "BallotGen 0.0.1",

    # data
    "Election": [
        {
            # required
            "@type": "ElectionResults.Election",
            "Name":"Hypothetical Election",
            "Type": "special",
            "ElectionScopeId": "gp1",
            "StartDate": "2022-11-08",
            "EndDate": "2022-11-08",
            # other
            "BallotStyle": [
                {
                    "@type": "ElectionResults.BallotStyle",
                    "GpUnitIds": ["gp1"],
                    "OrderedContent": [
                        {
                            "@type": "ElectionResults.OrderedHeader",
                            "HeaderId": "header1",
                        },
                        {
                            "@type": "ElectionResults.OrderedContest",
                            "ContestId": "contest1",
                        },
                    ],
                },
            ],
            "Candidate": [
                {
                    #required
                    "@id": "candidate1",
                    "@type": "ElectionResults.Candidate",
                    "BallotName": "",
                    #etc
                    "PersonId": "",
                },
            ],
            "Contest": [
                {
                    # required
                    "@id": "contest1",
                    "@type": "ElectionResults.CandidateContest",
                    "Name": "Everything",
                    "ElectionDistrictId": "gp1",
                    "VoteVariation": "approval",
                    "VotesAllowed": 9, # TODO: for approval, number of choices
                    # other
                    "ContestSelection": [
                        {
                            "@id": "csel1",
                            "@type": "ElectionResults.CandidateSelection",
                            "CandidateIds": [],
                        },
                    ],
                    "NumberElectiod": 1,
                    "OfficeIds": ["office2"],
                },
            ],
        },
    ],
    "GpUnit": [
        {
            "@id": "gp1",
            "@type": "ElectionResults.ReportingUnit",
            "Type": "city",
            "Name": "Springfield",
        },
    ],
    "Header": [
        {
            "@id": "header1",
            "@type": "ElectionResults.Header",
            "Name": "Header 1",
        },
    ],
    "Office": [
        {
            "@id": "office1",
            "@type": "ElectionResults.Office",
            "Name": "Head Dwarf",
        },
        {
            "@id": "office2",
            "@type": "ElectionResults.Office",
            "Name": "Everything",
        },
        {
            "@id": "office3",
            "@type": "ElectionResults.Office",
            "Name": "Bottom",
            "Description": "The race for the bottom",
        },
    ],
    #"OfficeGroup": [],
    "Party": [
        {
            "@id": "party1",
            "@type": "ElectionResults.Party",
            "Name": "Woot",
            "Slogan": "Come party with the party",
        },
    ],
    "Person": [
        {
            "@id": "person1",
            "@type": "ElectionResults.Person",
            "FullName": "Zaphod Beeblebrox",
            "PartyId": "party1",
            "Profession": "He's just this guy, you know?",
        },
        {
            "@id": "person2",
            "@type": "ElectionResults.Person",
            "FullName": "Zod",
            "PartyId": "party1",
            "Title": "General",
            "Profession": "Kneel",
        },
        {
            "@id": "person3",
            "@type": "ElectionResults.Person",
            "FullName": "Zardoz",
            "PartyId": "party1",
            "Profession": "There can be only one",
        },
    ],
    "IsTest": True,
    "TestType": "pre-election,design",
}
json.dump(ElectionReport, sys.stdout, indent=2)
sys.stdout.write('\n')
sys.exit(0)

c.rect(0.5 * inch, heightpt - 3.4 * inch, widthpt - 1.0 * inch, 2.9 * inch, stroke=1, fill=0)
c.drawString(0.7*inch, heightpt - 0.8*inch, 'instruction text here, etc.')

races = [therace, headDwarfRace, raceZ]
x = 0.5 * inch
top = heightpt - 0.5*inch - 3*inch
y = top
bottom = 0.5 * inch
colwidth = (7.5/2)*inch - gs.columnMargin
for xr in races:
    height = xr.height()
    if y - height < bottom:
        y = top
        x += colwidth + gs.columnMargin
        # TODO: check for next-page
    xr.draw(c, x, y, colwidth)
    y -= xr.height()
#therace.draw(c,x, y)
#y -= therace.height()
#race2.draw(c, 0.5 * inch, y)

c.showPage()
c.save()

bd = {'bubbles': therace.getBubbles()}
print(json.dumps(bd))
