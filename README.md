# Ballot Studio

System for designing, printing, and scanning paper ballots.

Ballot Studio is designed to be a core piece of an open source voting system.

## Getting Started

Ballot Studio is in two parts, a Python back end that draws PDF ballots, and a go middle layer that runs a web app with login and database and editing ballot data and scanning ballot images.

### Python Setup

* `python3 -m venv bsvenv`
* `bsvenv/bin/pip install fonttools Flask mercurial`
* `bsvenv/bin/hg clone https://hg.reportlab.com/hg-public/reportlab`
* `(cd reportlab && ../bsvenv/bin/pip install -e .)`
* get the resources blob (images and fonts):
  * `curl -O https://bolson.org/ballotstudio/resources.tar.gz`
  * `tar zxvf resources.tar.gz`
* Liberation fonts may have newer versions available at https://github.com/liberationfonts/liberation-fonts/releases

### Go Setup

* Install Go from https://golang.org/dl
* `make`
* install poppler-utils to rasterize pdf, one of:
  * `apt-get install -y poppler-utils`
  * `yum install -y poppler-utils`
  * `git clone https://github.com/brianolson/poppler.git`
     * See Development dependencies below
* `./ballotstudio -flask bsvenv/bin/flask -sqlite bss -debug`
  * **open the login link shown in initial status log lines**

## Development

Dependencies:

* https://github.com/brianolson/login
  * Local user/pass and oauth2 login
* https://github.com/brianolson/httpcache
  * Simple local http cached used by login
* https://github.com/brianolson/poppler
  * Fork of poppler with pdftoppm that does multiple pages in one operation
  * `git clone https://github.com/brianolson/poppler.git`
  * `(cd poppler && git checkout png-multi-block && mkdir build)``
     * redhatish:
       * `sudo yum install -y cmake3 freetype-devel fontconfig-devel libjpeg-turbo-devel openjpeg2-devel libtiff-devel`
       * `(cd poppler/build && cmake3 .. && make pdftoppm)`
     * debianish:
       * `sudo apt-get install -y libjpeg-dev libopenjp2-7-dev`
       * `(cd poppler/build && cmake .. && make pdftoppm)`
  * copy `poppler/build/utils/pdftoppm` to somewhere at head of PATH to be found by `cmd/ballotstudio/drawsub.go`

### Developing the Python Draw Server

`ballotstudio` will normally automatically start and stop the draw server.
It can be run on its own with:

`FLASK_ENV=development FLASK_APP=draw/app.py bsvenv/bin/flask run`

`ballotstudio` can be given the option `-draw-backend http://127.0.0.1:5000/` to point to that at Flask's default port 5000.

## Production Notes

The draw server should can be run by gunicorn for a production environment. `ballotstudio` would be given a `-draw-backend http://localhost:port/` option to point at the gunicorn server.

## NIST 1500-100 extensions

NIST 1500-100 (version 2) is a specification on election results *reporting*, but is used here because it has all the structural information about candidates and contests and the election as a whole.
We extend it with a few additional fields about ballot layout and rendering.

### "ElectionResults.BallotStyle"

Optional field "PageHeader" is a string that would be rendered at the top of each page. For example:

```
Header Election Name, YYYY-MM-DD
Precinct 1234, Some Town, Statename, page {PAGE} of {PAGES}
```

`\n` newlines within the text will be rendered.
Templated fields within the text are:

* `{PAGE}` the current page number
* `{PAGES}` the total number of pages
