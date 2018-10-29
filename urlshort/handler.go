package urlshort

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format)
// and responds with a 301 redirect.
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	if pathsToUrls == nil {
		pathsToUrls = make(map[string]string)
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var handler http.Handler
		if url, found := pathsToUrls[request.URL.Path]; found && len(url) > 0 {
			handler = http.RedirectHandler(url, http.StatusMovedPermanently)
		} else if fallback != nil {
			handler = fallback
		} else {
			handler = http.NotFoundHandler()
		}
		handler.ServeHTTP(writer, request)
	})
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//     - path: /some-path
//       url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yaml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	parsedYaml, err := parseYAML(yaml)
	if err != nil {
		return nil, err
	}
	pathMap := buildMap(parsedYaml)
	return MapHandler(pathMap, fallback), nil
}

func parseYAML(data []byte) ([]pathExpansion, error) {
	var expansions []pathExpansion
	err := yaml.UnmarshalStrict(data, &expansions)
	if err != nil {
		return nil, err
	}
	return expansions, nil
}

// JSONHandler will parse the provided JSON and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// JSON is expected to be in the format:
//
//  [
//    {
//      path: "/some-path",
//      url:  "https://www.some-url.com/demo"
//    },
//  ]
//
// The only errors that can be returned all related to having
// invalid JSON data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func JSONHandler(json []byte, fallback http.Handler) (http.HandlerFunc, error) {
	parsedJSON, err := parseJSON(json)
	if err != nil {
		return nil, err
	}
	pathMap := buildMap(parsedJSON)
	return MapHandler(pathMap, fallback), nil
}

func parseJSON(data []byte) ([]pathExpansion, error) {
	var expansions []pathExpansion
	err := json.Unmarshal(data, &expansions)
	if err != nil {
		return nil, err
	}
	return expansions, nil
}

func buildMap(expansions []pathExpansion) map[string]string {
	pathsToUrls := make(map[string]string)
	for _, expansion := range expansions {
		if expansion.valid() {
			pathsToUrls[expansion.Path] = expansion.URL
		}
	}
	return pathsToUrls
}

type pathExpansion struct {
	Path string `yaml:"path" json:"path"`
	URL  string `yaml:"url"  json:"url"`
}

func (expansion *pathExpansion) valid() bool {
	return expansion != nil && expansion.Path != "" && expansion.URL != ""
}

// LoadHandlerFromFile reads the YAML or JSON file
// and will return the results of calling YAMLHandler or JSONHandler.
// Should the file does not have a .yml, .yaml, or .json extention,
// both file formats will be tried.
// Returns an error when the file cannot be read or the
// JSON or YAML cannot be parsed.
func LoadHandlerFromFile(filename string, fallback http.Handler) (http.HandlerFunc, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(filepath.Ext(filename)) {
	case ".yaml", ".yml":
		handler, err := YAMLHandler(data, fallback)
		if err != nil {
			return nil, parsingError(filename, "YAML", err)
		}
		return handler, nil
	case ".json":
		handler, err := JSONHandler(data, fallback)
		if err != nil {
			return nil, parsingError(filename, "JSON", err)
		}
		return handler, nil
	default:
		if handler, err := JSONHandler(data, fallback); err == nil {
			return handler, nil
		}
		if handler, err := YAMLHandler(data, fallback); err == nil {
			return handler, nil
		}
		return nil, parsingError(filename, "UNKNOWN", nil)
	}
}

func parsingError(filename string, format string, err error) error {
	if err == nil {
		return fmt.Errorf("unable to parse %s formatted file: %s", format, filename)
	}
	return fmt.Errorf("unable to parse %s formatted file: %s: %v", format, filename, err)
}
