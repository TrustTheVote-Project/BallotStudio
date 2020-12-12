ballotstudio:	static/demoelection.json .PHONY data/type_seq_json.go
	go build ./cmd/ballotstudio

data/type_seq_json.go:	data/type_seq.json misc/texttosource/main.go
	cd data && go generate

static/demoelection.json:	draw/demorace.py
	python3 draw/demorace.py > static/demoelection.json

.PHONY:
