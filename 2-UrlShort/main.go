package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// ptu struct represent a path-to-url mapping
type ptu struct {
	Path string `yaml:"path"`
	URL  string `yaml:"url"`
}

// ptus struct represents a ptu list
type ptus struct {
	ps []ptu
}

// defaultMux function returns the default request multiplexer
func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

// hello function is the default handler
func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World!")
}

// MapHandler function build request handler according to given map and default handler
func MapHandler(pathsToUrls *map[string]string, fallback http.Handler) http.HandlerFunc {
	ret := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if rdirect, ok := (*pathsToUrls)[path]; ok {
			fmt.Printf("ready to redirect from [%s] to [%s]\n", path, rdirect)
			r.URL.Path = rdirect
			http.Redirect(w, r, rdirect, http.StatusMovedPermanently)
		} else {
			fmt.Println("use the default handler.")
			fallback.ServeHTTP(w, r)
		}
	}

	return ret
}

// parseYAML function parse configuration YAML file into ptus objects
func parseYAML(yml []byte) (*ptus, error) {
	pathUrls := ptus{}
	err := yaml.Unmarshal(yml, &pathUrls.ps)
	if err != nil {
		return nil, err
	}

	return &pathUrls, nil
}

// buildMap function convert parsed path-to-urls into a map
func buildMap(parsedYAML *ptus) *map[string]string {
	ret := make(map[string]string, len(parsedYAML.ps))

	for _, p := range parsedYAML.ps {
		ret[p.Path] = p.URL
	}

	return &ret
}

// YAMLHandler function build request handler according to given YAML configuration
func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	parsedYAML, err := parseYAML(yml)
	if err != nil {
		return nil, err
	}

	pathMap := buildMap(parsedYAML)
	return MapHandler(pathMap, fallback), nil
}

// readYAML function read YAML configuration from file 'filepath'
func readYAML(filepath string) (*[]byte, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ret, err := ioutil.ReadAll(f)
	return &ret, err
}

var (
	yamlfile = flag.String("yaml", "path_to_urls.yaml", "the configuration YAML file path")
)

func main() {
	flag.Parse()

	mux := defaultMux()

	yamlByte, err := readYAML(*yamlfile)
	if err != nil {
		panic(err)
	}

	yamlhandler, err := YAMLHandler(*yamlByte, mux)
	if err != nil {
		panic(err)
	}

	fmt.Println("starting the server on :8080")
	err = http.ListenAndServe(":8080", yamlhandler)
	if err != nil {
		panic(err)
	}
}
