package main

import (
	"testing"
)

var yml = `
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url: https://github.com/gophercises/urlshort/tree/solution
  `

func TestParseYAML(t *testing.T) {
	ret, err := parseYAML([]byte(yml))
	if err != nil {
		t.Errorf("parse yaml error: %s\n", err.Error())
	} else {
		t.Log(*ret)
	}
}

func TestBuildMap(t *testing.T) {
	parsed, err := parseYAML([]byte(yml))
	if err != nil {
		t.Errorf("parse yaml error: %s\n", err.Error())
		return
	}

	YAMLMap := buildMap(parsed)

	for k, v := range *YAMLMap {
		t.Logf("key: %s, value: %s\n", k, v)
	}
}

var jsondata = `
[
    {
        "path": "/urlshort",
        "url": "https://github.com/gophercises/urlshort"
    },
    {
        "path": "/urlshort-final",
        "url": "https://github.com/gophercises/urlshort/tree/solution"
    },
    {
        "path": "/urlshort-godoc",
        "url": "https://godoc.org/github.com/gophercises/urlshort"
    },
    {
        "path": "/yaml-godoc",
        "url": "https://godoc.org/gopkg.in/yaml.v2"
    }
]
`

func TestParseJSON(t *testing.T) {
	jsonbyte := []byte(jsondata)
	p, err := parseJSON(&jsonbyte)
	if err != nil {
		t.Errorf(err.Error())
	} else {
		t.Log(p)
	}
}