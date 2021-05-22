#!/usr/bin/env python3

'''demo NIST 1500-100 v2 race in JSON style'''

import json
import logging
import os
import sys
import time

logger = logging.getLogger(__name__)

idSourcesByType = {}

class _seqSource:
    def __init__(self, s, sel):
        self.s = s
        self.sel = sel
    def __call__(self):
        return self.s.next(self.sel)

class Sequences:
    def __init__(self):
        self.sequences = {}
        self.attype = {}
        self.unki = 0
    def next(self, sel):
        v = self.sequences.get(sel, 0) + 1
        self.sequences[sel] = v
        return str(sel) + str(v)
    def source(self, sel):
        return _seqSource(self, sel)
    def setTypeMap(self, ob):
        self.attype = ob
    def sourceForType(self, attype):
        sel = self.attype.get(attype)
        if sel is None:
            self.unki += 1
            sel = 'unk%d_'.format(self.unki)
            logger.warning('unknown attype %r, setting up seq %s', attype, sel)
            self.attype[attype] = sel
        return _seqSource(self, sel)

typeSequences = Sequences()

_thisdir = os.path.dirname(os.path.abspath(__file__))
_path = (_thisdir, os.path.join(os.path.dirname(_thisdir), 'data'))
for xd in _path:
    fp = os.path.join(xd, 'type_seq.json')
    if os.path.exists(fp):
        with open(fp) as fin:
            ob = json.load(fin)
        typeSequences.setTypeMap(ob)
        break

_party_id = typeSequences.sourceForType("ElectionResults.Party")

parties = [
    {
        "@id": _party_id(),
        "@type": "ElectionResults.Party",
        "Name": "Woot",
        "Slogan": "Come party with the party",
    },
    {
        "@id": _party_id(),
        "@type": "ElectionResults.Party",
        "Name": "Anklebiter Assembly",
        "Slogan": "grrr, narf, bite",
    },
    {
        "@id": _party_id(),
        "@type": "ElectionResults.Party",
        "Name": "Boring Board",
        "Slogan": "zzzzz",
    },
    {
        "@id": _party_id(),
        "@type": "ElectionResults.Party",
        "Name": "Cowardly Coalition",
        "Slogan": "zzzzz",
    },
    {
        "@id": _party_id(),
        "@type": "ElectionResults.Party",
        "Name": "Dumb Department",
        "Slogan": "duh, er, wut?",
    },
    {
        "@id": _party_id(),
        "@type": "ElectionResults.Party",
        "Name": "Electable Entertainers",
        "Slogan": "Let us entertain you",
    },
    {
        "@id": _party_id(),
        "@type": "ElectionResults.Party",
        "Name": "Galaxy",
        "Slogan": "Don't Panic",
    },
    {
        "@id": _party_id(),
        "@type": "ElectionResults.Party",
        "Name": "Krypton",
        "Slogan": "Nice yellow sun you got there",
    },
]

def partyIdByName(they, name):
    for p in they:
        if p['Name'] == name:
            return p['@id']
    raise KeyError(name)

_person_id = typeSequences.sourceForType("ElectionResults.Person")

persons = [
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Zaphod Beeblebrox",
        "PartyId": partyIdByName(parties, 'Galaxy'),
        "Profession": "He's just this guy, you know?",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Zod",
        "PartyId": partyIdByName(parties, 'Krypton'),
        "Title": "General",
        "Profession": "Kneel",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Zardoz",
        "PartyId": partyIdByName(parties, 'Woot'),
        "Profession": "There can be only one",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Alice Argyle",
        "PartyId": partyIdByName(parties, 'Anklebiter Assembly'),
        "Profession": "Axolotol",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Bob Brocade",
        "PartyId": partyIdByName(parties, 'Boring Board'),
        "Profession": "Bollard",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Çandidate Ñame 亀",
        "PartyId": partyIdByName(parties, 'Cowardly Coalition'),
        "Profession": "Camp Councilor",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Dorian Duck",
        "PartyId": partyIdByName(parties, 'Dumb Department'),
        "Profession": "Doorstop",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Elaine Entwhistle",
        "PartyId": partyIdByName(parties, 'Electable Entertainers'),
        "Profession": "Electrician",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Sleepy",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Happy",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Dopey",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Grumpy",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Sneezy",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Bashful",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Doc",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Larry",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Moe",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Curly",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Kevin",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Tamika Flynn",
    },
    {
        "@id": _person_id(),
        "@type": "ElectionResults.Person",
        "FullName": "Hiram McDaniels",
    },
]

def personIdByFullName(they, name):
    for p in they:
        if p['FullName'] == name:
            return p['@id']
    raise KeyError(name)

_candidate_id = typeSequences.sourceForType("ElectionResults.Candidate")
_csel_id = typeSequences.sourceForType("ElectionResults.CandidateSelection")
_bmsel_id = typeSequences.sourceForType("ElectionResults.BallotMeasureSelection")

def makeCandidate(fullname, persons):
    return {
        #required
        "@id": _candidate_id(),
        "@type": "ElectionResults.Candidate",
        "BallotName": fullname,
        #etc
        "PersonId": personIdByFullName(persons, fullname),
    }

# make a Candidate wrapper for every Person
candidates = [
    makeCandidate(x['FullName'], persons) for x in persons
]

def candidateIdByName(candidates, name):
    for x in candidates:
        if x['BallotName'] == name:
            return x['@id']
    raise KeyError(name)

def candidateIdsForNames(candidates, *args):
    return [candidateIdByName(candidates, x) for x in args]

_office_id = typeSequences.sourceForType("ElectionResults.Office")

offices = [
    {
        "@id": _office_id(),
        "@type": "ElectionResults.Office",
        "Name": "Head Dwarf",
    },
    {
        "@id": _office_id(),
        "@type": "ElectionResults.Office",
        "Name": "Chief Stooge",
    },
    {
        "@id": _office_id(),
        "@type": "ElectionResults.Office",
        "Name": "Everything",
    },
    {
        "@id": _office_id(),
        "@type": "ElectionResults.Office",
        "Name": "Bottom",
        "Description": "The race for the bottom",
    },
    {
        "@id": _office_id(),
        "@type": "ElectionResults.Office",
        "Name": "Smile",
        "Description": "Best Smile",
    },
]

def officeIdByName(they, name):
    for x in they:
        if x['Name'] == name:
            return x['@id']
    raise KeyError(name)

_gpunit_id = typeSequences.sourceForType("ElectionResults.ReportingUnit")

gpunits = [
    {
        "@id": _gpunit_id(),
        "@type": "ElectionResults.ReportingUnit",
        "Type": "city",
        "Name": "Springfield",
    },
    {
        "@id": _gpunit_id(),
        "@type": "ElectionResults.ReportingUnit",
        "Type": "city",
        "Name": "Desert Bluffs",
    },
    # TODO: heirarchical example, e.g. state-county-city nesting
]

gpunitIdByName = officeIdByName

_contest_id = typeSequences.sourceForType("ElectionResults.CandidateContest")
_bmcont_id = typeSequences.sourceForType("ElectionResults.BallotMeasureContest")

def candidateSelectionsFromNames(candidates, *names):
    out = []
    for name in names:
        out.append({
            "@id": _csel_id(),
            "@type": "ElectionResults.CandidateSelection",
            "CandidateIds": [candidateIdByName(candidates, name)],
        })
    return out

def yesOrNoBallotMeasureSelections():
    return [
        {
            "@id": _bmsel_id(),
            "@type": "ElectionResults.BallotMeasureSelection",
            "Selection": "Yes",
            "SequenceOrder": 1,
        },
        {
            "@id": _bmsel_id(),
            "@type": "ElectionResults.BallotMeasureSelection",
            "Selection": "No",
            "SequenceOrder": 2,
        },
    ]

contests = [
    {
        # required
        "@id": _contest_id(),
        "@type": "ElectionResults.CandidateContest",
        "Name": "Everything",
        "ElectionDistrictId": gpunitIdByName(gpunits, 'Springfield'),
        "VoteVariation": "approval",
        "VotesAllowed": 5, # TODO: for approval, number of choices
        # other
        "BallotTitle": "The Race For Everything",
        "BallotSubTitle": "Vote for as many as you like",
        "ContestSelection": candidateSelectionsFromNames(candidates, 'Alice Argyle', "Bob Brocade", "Çandidate Ñame 亀", "Dorian Duck", "Elaine Entwhistle"),
        "NumberElected": 1,
        "OfficeIds": [officeIdByName(offices, 'Everything')],
    },
    {
        # required
        "@id": _contest_id(),
        "@type": "ElectionResults.CandidateContest",
        "Name": "Head Dwarf",
        "ElectionDistrictId": gpunitIdByName(gpunits, 'Springfield'),
        "VoteVariation": "plurality",
        "VotesAllowed": 1,
        # other
        "BallotTitle": "Head Dwarf",
        "BallotSubTitle": "Vote for one",
        "ContestSelection": candidateSelectionsFromNames(candidates, "Sleepy", "Happy", "Dopey", "Grumpy", "Sneezy", "Bashful", "Doc"),
        "NumberElected": 1,
        "OfficeIds": [officeIdByName(offices, 'Head Dwarf')],
    },
    {
        # required
        "@id": _contest_id(),
        "@type": "ElectionResults.CandidateContest",
        "Name": "Chief Stooge",
        "ElectionDistrictId": gpunitIdByName(gpunits, 'Springfield'),
        "VoteVariation": "plurality",
        "VotesAllowed": 1,
        # other
        "BallotTitle": "Chief Stooge",
        "BallotSubTitle": "Vote for one",
        "ContestSelection": candidateSelectionsFromNames(candidates, "Larry", "Moe", "Curly"),
        "NumberElected": 1,
        "OfficeIds": [officeIdByName(offices, 'Chief Stooge')],
    },
    {
        # required
        "@id": _bmcont_id(),
        "@type": "ElectionResults.BallotMeasureContest",
        "Name": "Winning",
        "ElectionDistrictId": gpunitIdByName(gpunits, 'Springfield'),
        #"VoteVariation": "plurality",
        #"VotesAllowed": 1,
        # other
        "BallotTitle": "Should We Win",
        "BallotSubTitle": "Vote Yes or No",
        "ConStatement": "Winning is hard work, let's take a nap",
        "ProStatement": "Winning is awesome, let's do it",
        "ContestSelection": yesOrNoBallotMeasureSelections(),
        #"EffectOfAbstain": "Not voting is dumb",
        "FullText": "blah blah blah [insert full text of plan here] fnord fnord fnord",
        "InfoUri": "https://betterpolls.com/",
        "SummaryText": "[insert short text here]",
        "Type": "referendum",
    },
    {
        # required
        "@id": _contest_id(),
        "@type": "ElectionResults.CandidateContest",
        "Name": "Bottom",
        "ElectionDistrictId": gpunitIdByName(gpunits, 'Springfield'),
        "VoteVariation": "plurality",
        "VotesAllowed": 1,
        # other
        "BallotTitle": "The Race To The Bottom",
        "BallotSubTitle": "Vote for one",
        "ContestSelection": candidateSelectionsFromNames(candidates, "Zaphod Beeblebrox", "Zod", "Zardoz"), # TODO: write-in
        "NumberElected": 1,
        "OfficeIds": [officeIdByName(offices, 'Bottom')],
    },
    {
        # required
        "@id": _contest_id(),
        "@type": "ElectionResults.CandidateContest",
        "Name": "Best Smile",
        "ElectionDistrictId": gpunitIdByName(gpunits, 'Desert Bluffs'),
        "VoteVariation": "plurality",
        "VotesAllowed": 1,
        # other
        "BallotTitle": "Best Smile",
        "BallotSubTitle": "Vote for one",
        "ContestSelection": candidateSelectionsFromNames(candidates, "Kevin", "Tamika Flynn", "Hiram McDaniels"), # TODO: write-in
        "NumberElected": 1,
        "OfficeIds": [officeIdByName(offices, 'Smile')],
    },
    # TODO: ElectionResults.RetentionContest
    # TODO: ElectionResults.PartyContest
]

contestIdByName = officeIdByName

_header_id = typeSequences.sourceForType("ElectionResults.Header")

headers = [
    {
        "@id": _header_id(),
        "@type": "ElectionResults.Header",
        "Name": "Instructions",
    },
    {
        "@id": _header_id(),
        "@type": "ElectionResults.Header",
        "Name": "ColumnBreak",
    },
    {
        "@id": _header_id(),
        "@type": "ElectionResults.Header",
        "Name": "PageBreak",
    },
]

headerIdByName = officeIdByName


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
    "VendorApplicationId": "bolson's ballots 0.0.1",

    # data
    "Election": [
        {
            # required
            "@type": "ElectionResults.Election",
            "Name":"Hypothetical Election",
            "Type": "special",
            "ElectionScopeId": gpunitIdByName(gpunits, 'Springfield'),
            "StartDate": "2022-11-08",
            "EndDate": "2022-11-08",
            # other
            "BallotStyle": [
                {
                    "@type": "ElectionResults.BallotStyle",
                    "GpUnitIds": [gpunitIdByName(gpunits, 'Springfield')],
                    "OrderedContent": [
                        {
                            "@type": "ElectionResults.OrderedHeader",
                            "HeaderId": headerIdByName(headers, 'Instructions'),
                        },
                        {
                            "@type": "ElectionResults.OrderedHeader",
                            "HeaderId": headerIdByName(headers, 'ColumnBreak'),
                        },
                        {
                            "@type": "ElectionResults.OrderedContest",
                            "ContestId": contestIdByName(contests, 'Everything'),
                        },
                        {
                            "@type": "ElectionResults.OrderedContest",
                            "ContestId": contestIdByName(contests, 'Head Dwarf'),
                        },
                        {
                            "@type": "ElectionResults.OrderedContest",
                            "ContestId": contestIdByName(contests, 'Chief Stooge'),
                        },
                        {
                            "@type": "ElectionResults.OrderedContest",
                            "ContestId": contestIdByName(contests, 'Winning'),
                        },
                        {
                            "@type": "ElectionResults.OrderedContest",
                            "ContestId": contestIdByName(contests, 'Bottom'),
                        },
                    ],
                    "PageHeader": '''General Election, 2022-11-08
Precinct 1234, Springfield, OR, page {PAGE} of {PAGES}''',
                },
                {
                    "@type": "ElectionResults.BallotStyle",
                    "GpUnitIds": [gpunitIdByName(gpunits, 'Desert Bluffs')],
                    "OrderedContent": [
                        {
                            "@type": "ElectionResults.OrderedHeader",
                            "HeaderId": headerIdByName(headers, 'Instructions'),
                        },
                        {
                            "@type": "ElectionResults.OrderedHeader",
                            "HeaderId": headerIdByName(headers, 'ColumnBreak'),
                        },
                        {
                            "@type": "ElectionResults.OrderedContest",
                            "ContestId": contestIdByName(contests, 'Everything'),
                        },
                        {
                            "@type": "ElectionResults.OrderedContest",
                            "ContestId": contestIdByName(contests, 'Best Smile'),
                        },
                        {
                            "@type": "ElectionResults.OrderedContest",
                            "ContestId": contestIdByName(contests, 'Winning'),
                        },
                        {
                            "@type": "ElectionResults.OrderedContest",
                            "ContestId": contestIdByName(contests, 'Bottom'),
                        },
                    ],
                    "PageHeader": '''General Election, 2022-11-08
Precinct 1234, Desert Bluffs, OR, page {PAGE} of {PAGES}''',
                },
            ],
            "Candidate": candidates,
            "Contest": contests,
        },
    ],
    "GpUnit": gpunits,
    "Header": headers,
    "Office": offices,
    #"OfficeGroup": [],
    "Party": parties,
    "Person": persons,
    "IsTest": True,
    "TestType": "pre-election,design",
}

if __name__ == '__main__':
    json.dump(ElectionReport, sys.stdout, indent=2, sort_keys=True)
    sys.stdout.write('\n')
