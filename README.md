# Ballot Studio

System for designing, printing, and scanning paper ballots.

## Getting Started

Ballot Studio is in two parts, a Python back end that draws PDF ballots, and a go middle layer that runs a web app with login and database and editing ballot data and scanning ballot images.

### Python Setup

* `python3 -m venv bsvenv`
* `bsvenv/bin/pip install fonttools Flask mercurial`
* `bsvenv/bin/hg clone https://hg.reportlab.com/hg-public/reportlab`
* `(cd reportlab && ../bsvenv/bin/pip install -e .)`
* `FLASK_ENV=development FLASK_APP=app.py bsvenv/bin/flask run`

### Go Setup

* Install Go from https://golang.org/dl
* `make`
* install poppler-utils to rasterize pdf, one of:
  * `apt-get install -y poppler-utils`
  * `yum install -y poppler-utils`
  * `git clone https://github.com/brianolson/poppler.git`
     * See Development dependencies below
* `./ballotstudio -draw-backend http://127.0.0.1:5000/ -sqlite bss`
  * open link shown in initial status log lines

## Development

Dependencies:

* https://github.com/brianolson/ballotscan
  * Compare scanned image to PNG generated cleanly from source PDF, find bubbles, report which bubbles are marked
* https://github.com/brianolson/login
  * Local user/pass and oauth2 login
* https://github.com/brianolson/httpcache
  * Simple local http cached used by login
* https://github.com/brianolson/poppler
  * Fork of poppler with pdftoppm that does multiple pages in one operation
  * `git clone https://github.com/brianolson/poppler.git`
  * `(cd poppler && git checkout png-multi-block && mkdir build)``
     * redhatish: `sudo yum install -y cmake3 freetype-devel fontconfig-devel libjpeg-turbo-devel openjpeg2-devel libtiff-devel && (cd poppler/build && cmake3 .. && make pdftoppm)`
     * debianish: `(cd poppler/build && cmake .. && make pdftoppm)`
  * copy `poppler/build/utils/pdftoppm` to somewhere at head of PATH to be found by `cmd/ballotstudio/drawsub.go`
