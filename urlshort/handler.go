package urlshort

import (
	"log"
	"net/http"

	"gopkg.in/yaml.v2"
)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		redirectUrl, ok := pathsToUrls[path]
		if ok {
			log.Println("Before custom with " + redirectUrl)
			http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
			log.Println("After custom")
		} else {
			log.Println("Before")
			fallback.ServeHTTP(w, r) // call original
			log.Println("After")
		}
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
func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	// Parse yml to a map
	pathsToUrls := make(map[string]string)
	seq := make([]Item, 0)
	err := yaml.Unmarshal(yml, &seq)
	if err != nil {
		return nil, err
	}
	log.Printf("Seq %+v \n", seq)
	for _, item := range seq {
		pathsToUrls[item.Path] = item.Url
	}
	log.Printf("pathsToUrls %v \n", pathsToUrls)
	mapHandler := MapHandler(pathsToUrls, fallback)
	return mapHandler, nil
}

/* Example YAML:
`
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url: https://github.com/gophercises/urlshort/tree/solution
`
*/

type Item struct {
	Path string `yaml:"path"`
	Url  string `yaml:"url"`
}
