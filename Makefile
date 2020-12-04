ballotstudio:	static/demoelection.json .PHONY
	go build ./cmd/ballotstudio

static/demoelection.json:	draw/demorace.py
	python3 draw/demorace.py > static/demoelection.json

.PHONY:

