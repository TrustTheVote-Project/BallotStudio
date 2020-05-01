#!/usr/bin/env python3

'''demo NIST 1500-100 v2 race in JSON style'''

import json
import sys
import time

class IdSource:
    def __init__(self, root, start=1):
        self.root = root
        self.count = start
    def __call__(self):
        out = self.root + str(self.count)
        self.count += 1
        return out

_party_id = IdSource('party')

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
        "Name": "Kryptonian",
        "Slogan": "Nice yellow sun you got there",
    },
]

def partyIdByName(they, name):
    for p in they:
        if p['Name'] == name:
            return p['@id']
    raise KeyError(name)

_person_id = IdSource('person')

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
        "PartyId": partyIdByName(parties, 'Kryptonian'),
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
]

def personIdByFullName(they, name):
    for p in they:
        if p['FullName'] == name:
            return p['@id']
    raise KeyError(name)

_candidate_id = IdSource('candidate')
_csel_id = IdSource('csel')

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

_office_id = IdSource('office')

offices = [
    {
        "@id": _office_id(),
        "@type": "ElectionResults.Office",
        "Name": "Head Dwarf",
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
]

def officeIdByName(they, name):
    for x in they:
        if x['Name'] == name:
            return x['@id']
    raise KeyError(name)

_gpunit_id = IdSource('gpunit')

gpunits = [
    {
        "@id": _gpunit_id(),
        "@type": "ElectionResults.ReportingUnit",
        "Type": "city",
        "Name": "Springfield",
    },
]

gpunitIdByName = officeIdByName

_contest_id = IdSource('contest')

def candidateSelectionsFromNames(candidates, *names):
    out = []
    for name in names:
        out.append({
            "@id": _csel_id(),
            "@type": "ElectionResults.CandidateSelection",
            "CandidateIds": [candidateIdByName(candidates, name)],
        })
    return out

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
        "Name": "Bottom",
        "ElectionDistrictId": gpunitIdByName(gpunits, 'Springfield'),
        "VoteVariation": "plurality",
        "VotesAllowed": 1,
        # other
        "BallotTitle": "The Race To The Bottom",
        "BallotSubTitle": "Vote for one",
        "ContestSelection": candidateSelectionsFromNames(candidates, "Zaphod Beeblebrox", "Zod", "Zardoz"), # TODO: write-in
        "NumberElected": 1,
        "OfficeIds": [officeIdByName(offices, 'Head Dwarf')],
    },
]

contestIdByName = officeIdByName

_header_id = IdSource('header')

headers = [
    {
        "@id": _header_id(),
        "@type": "ElectionResults.Header",
        "Name": "Instructions",
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
                        # {
                        #     "@type": "ElectionResults.OrderedHeader",
                        #     "HeaderId": headerIdByName(headers, 'Instructions'),
                        # },
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
                            "ContestId": contestIdByName(contests, 'Bottom'),
                        },
                    ],
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
    json.dump(ElectionReport, sys.stdout, indent=2)
    sys.stdout.write('\n')
