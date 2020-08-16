ballotstudio:	static/demoelection.json .PHONY
	go build ./cmd/ballotstudio

static/demoelection.json:	demorace.py
	python3 demorace.py > static/demoelection.json

.PHONY:

