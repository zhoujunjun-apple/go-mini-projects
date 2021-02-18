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
