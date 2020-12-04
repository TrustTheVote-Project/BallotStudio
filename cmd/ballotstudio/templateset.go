package main

import (
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type tse struct {
	t       *template.Template
	lastmod time.Time
	path    string
}

type TemplateSet struct {
	Reloading bool

	Glob string

	they map[string]tse
}

func HtmlTemplateGlob(pat string) (out TemplateSet, err error) {
	out.Glob = pat
	matches, err := filepath.Glob(pat)
	if err != nil {
		return
	}
	out.they = make(map[string]tse, len(matches))
	for _, fpath := range matches {
		_, fname := filepath.Split(fpath)
		finfo, err := os.Stat(fpath)
		if err != nil {
			return out, err
		}
		nt := template.New(fname)
		b, err := ioutil.ReadFile(fpath)
		if err != nil {
			return out, err
		}
		nt, err = nt.Parse(string(b))
		if err != nil {
			return out, err
		}
		out.they[fname] = tse{nt, finfo.ModTime(), fpath}
	}
	return out, nil
}

func (ts *TemplateSet) Lookup(name string) (t *template.Template, err error) {
	ent, ok := ts.they[name]
	if !ok {
		return
	}
	if ts.Reloading {
		finfo, err := os.Stat(ent.path)
		if err != nil {
			return t, err
		}
		mtime := finfo.ModTime()
		if mtime.After(ent.lastmod) {
			nt := template.New(name)
			b, err := ioutil.ReadFile(ent.path)
			if err != nil {
				return t, err
			}
			nt, err = nt.Parse(string(b))
			if err != nil {
				return t, err
			}
			ent.t = nt
			ent.lastmod = mtime
			ts.they[name] = ent
		}
	}
	return ent.t, nil
}
