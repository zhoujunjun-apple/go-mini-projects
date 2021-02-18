package main

import (
	"encoding/json"
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

// parseJSON function parse configuraion JSON file into ptus objects
func parseJSON(jsondata *[]byte) (*ptus, error) {
	ptusObj := new(ptus)
	err := json.Unmarshal(*jsondata, &ptusObj.ps)
	if err != nil {
		return nil, err
	}

	return ptusObj, nil
}

// buildMap function convert parsed path-to-urls into a map
func buildMap(parsed *ptus) *map[string]string {
	ret := make(map[string]string, len(parsed.ps))

	for _, p := range parsed.ps {
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

// readByte function read configuration from file 'filepath'
func readByte(filepath string) (*[]byte, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ret, err := ioutil.ReadAll(f)
	return &ret, err
}

// JSONHandler function build request handler according to given JSON configuration
func JSONHandler(jsondata *[]byte, fallback http.Handler) (http.HandlerFunc, error) {
	parsedJSON, err := parseJSON(jsondata)
	if err != nil {
		return nil, err
	}

	pathMap := buildMap(parsedJSON)
	return MapHandler(pathMap, fallback), nil
}

func exit(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	yamlfile = flag.String("yaml", "", "the configuration YAML file path")
	jsonfile = flag.String("json", "", "the configuration JSON file path with higher priority than 'yaml' flag")
)

func main() {
	flag.Parse()

	mux := defaultMux()

	var handler http.HandlerFunc
	if *jsonfile != "" {
		cfg, err := readByte(*jsonfile)
		exit(err)

		jhandler, err := JSONHandler(cfg, mux)
		exit(err)

		handler = jhandler
	} else if *yamlfile != "" {
		cfg, err := readByte(*yamlfile)
		exit(err)

		yhandler, err := YAMLHandler(*cfg, mux)
		exit(err)

		handler = yhandler
	} else {
		panic(fmt.Errorf("one of 'yaml' or 'json' flag must be specified. input -h or --help to check all the flags"))
	}
	
	fmt.Println("starting the server on :8080")
	err := http.ListenAndServe(":8080", handler)
	exit(err)
}
