# NIST 1500-100 V2 JSON document structure

This describes the nested structure of record types, the fields within each record are described in other documents. This doesn't include references, but only nested objects.

ElectionResults.Election is the top level record. Under that:

- BallotCounts
- BallotStyle
   - OrderedContent, items of 2 types:
      - OrderedHeader
         - recursively contains OrderedContent items
	 - refs Header
      - OrderedContest
         - refs {PartyContest, BallotMeasureContest, CandidateContest, RetentionContest}; {PartySelection, BallotMeasureSelection, CandidateSelection}
   - refs {ReportingDevice, ReportingUnit}, {Party, Coalition,}
- Candidate
   - refs Person
- Contest, items of 4 types:
   - BallotMeasureContest
      - BallotMeasureSelection
      - refs ReportingUnit
   - CandidateContest
      - CandidateSelection
         - refs Candidate, {Party, Coalition}
      - refs Office, {Party, Coalition}
   - PartyContest
      - PartySelection
         - refs {Party, Coalition}
      - refs ReportingUnit
   - RetentionContest
      - CandidateSelection
      - refs Candidate, ReportingUnit, Office
- CountStatus
