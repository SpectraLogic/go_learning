package urlshort

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type Outcome int

const (
	ERROR    Outcome = -1
	REDIRECT Outcome = http.StatusMovedPermanently
	FALLBACK Outcome = http.StatusOK
	NOTFOUND Outcome = http.StatusNotFound
)

func (outcome Outcome) statusCode() int {
	return int(outcome)
}

func (outcome Outcome) nameAndCode() string {
	if outcome == ERROR {
		return ERROR.String()
	}
	return fmt.Sprintf("%v(%d)", outcome.String(), outcome.statusCode())
}

func (outcome Outcome) String() string {
	switch outcome {
	case REDIRECT:
		return "Redirect"
	case FALLBACK:
		return "Fallback"
	case NOTFOUND:
		return "Not Found"
	case ERROR:
		return "Error"
	default:
		return "Unknown"
	}
}

type Want struct {
	wantOutcome     Outcome
	wantRedirectURL string
}

func WantRedirect(url string) Want {
	return Want{REDIRECT, url}
}

func WantFallback() Want {
	return Want{FALLBACK, ""}
}

func WantNotFound() Want {
	return Want{NOTFOUND, ""}
}

func WantError() Want {
	return Want{ERROR, ""}
}

func WantErrorPrefix(prefix string) Want {
	return Want{ERROR, prefix}
}

func (want *Want) GotError(err error) bool {
	return want.Got(ERROR, err.Error())
}

func (want *Want) Got(outcome Outcome, value string) bool {
	if want.wantOutcome != outcome {
		return false
	}
	switch outcome {
	case REDIRECT:
		return want.wantRedirectURL == value
	case FALLBACK, NOTFOUND:
		return want.wantRedirectURL == "" || want.wantRedirectURL == value
	case ERROR:
		return want.wantRedirectURL == "" || strings.HasPrefix(value, want.wantRedirectURL)
	default:
		return false
	}
}

func (want *Want) String() string {
	var mid string
	if want.wantRedirectURL != "" {
		if want.wantOutcome == ERROR {
			mid = " starting with: "
		} else {
			mid = " "
		}
	}
	return fmt.Sprintf("%s%s%s", want.wantOutcome.nameAndCode(), mid, want.wantRedirectURL)
}

type Result struct {
	outcome Outcome
	value   string
}

func (result *Result) String() string {
	var mid string
	if result.value != "" {
		if result.outcome == ERROR {
			mid = ": "
		} else {
			mid = " "
		}
	}
	return fmt.Sprintf("%s%s%s", result.outcome.nameAndCode(), mid, result.value)
}

func TestMapHandler(t *testing.T) {
	type args struct {
		url         string
		pathsToUrls map[string]string
		fallback    http.Handler
	}
	tests := []struct {
		name string
		args args
		want Want
	}{
		{"Test shortened url", args{"/urlshort", testPathsToUrls(), testFallback()}, WantRedirect("/long/expansion/of/url")},
		{"Test shortened external url with host", args{"http://localhost:8080/urlshort-godoc", testPathsToUrls(), testFallback()}, WantRedirect("https://godoc.org/github.com/gophercises/urlshort")},
		{"Test shortened external url with params", args{"/urlshort-godoc-map", testPathsToUrls(), testFallback()}, WantRedirect("https://godoc.org/github.com/gophercises/urlshort?field1=value+one&field2=value2#MapHandler")},
		{"Test fallback with map", args{"/not/in/map", testPathsToUrls(), testFallback()}, WantFallback()},
		{"Test empty url expansion", args{"/bad-short", testPathsToUrls(), testFallback()}, WantFallback()},
		{"Test empty path with expansion", args{"", testPathsToUrls(), testFallback()}, WantRedirect("/from/empty/path")},
		{"Test shortened no fallback", args{"/urlshort", testPathsToUrls(), testFallback()}, WantRedirect("/long/expansion/of/url")},
		{"Test no fallback", args{"/not/in/map", testPathsToUrls(), nil}, WantNotFound()},
		{"Test no map", args{"/urlshort", nil, testFallback()}, WantFallback()},
		{"Test no map no fallback", args{"/urlshort", nil, nil}, WantNotFound()},
	}

	for _, tt := range tests {
		t.Run(tt.name+" "+tt.want.wantOutcome.String(), func(t *testing.T) {
			handler := MapHandler(tt.args.pathsToUrls, tt.args.fallback)

			runHandlerTest(t, handler, tt.args.url, tt.want)
		})
	}
}

func TestYAMLHandler(t *testing.T) {
	type args struct {
		url      string
		yaml     []byte
		fallback http.Handler
	}
	tests := []struct {
		name string
		args args
		want Want
	}{
		{"Test shortened url", args{"/urlshort", testYAML(), testFallback()}, WantRedirect("https://github.com/gophercises/urlshort")},
		{"Test shortened url blank in Yaml", args{"/urlshort", testYAML2(), testFallback()}, WantRedirect("https://github.com/gophercises/urlshort")},
		{"Test shortened url", args{"/urlshort-final", testYAML(), testFallback()}, WantRedirect("https://github.com/gophercises/urlshort/tree/solution")},
		{"Test fallback with config", args{"/not/in/config", testYAML(), testFallback()}, WantFallback()},
		{"Test fallback with empty config array", args{"/not/in/config", []byte("[]"), testFallback()}, WantFallback()},
		{"Test fallback with empty config", args{"/not/in/config", []byte(""), testFallback()}, WantFallback()},
		{"Test fallback with nil config", args{"/not/in/config", nil, testFallback()}, WantFallback()},
		{"Test bad YAML", args{"/not/in/config", testBadYAML(), testFallback()}, WantError()},
		{"Test bad single JSON", args{"/not/in/config", []byte("{}"), testFallback()}, WantError()},
		{"Test no fallback", args{"/not/in/config", testYAML(), nil}, WantNotFound()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := YAMLHandler(tt.args.yaml, tt.args.fallback)
			if err != nil {
				if !tt.want.GotError(err) {
					t.Errorf("YAMLHandler() got %v, want %v",
						(&Result{ERROR, err.Error()}).String(), tt.want.String())
				}
				return
			}

			runHandlerTest(t, got, tt.args.url, tt.want)
		})
	}
}

func TestJSONHandler(t *testing.T) {
	type args struct {
		url      string
		json     []byte
		fallback http.Handler
	}
	tests := []struct {
		name string
		args args
		want Want
	}{
		{"Test valid", args{"/urlshort", testJSON(), testFallback()}, WantRedirect("https://github.com/gophercises/urlshort")},
		{"Test valid", args{"/urlshort-final", testJSON(), testFallback()}, WantRedirect("https://github.com/gophercises/urlshort/tree/solution")},
		{"Test fallback with config", args{"/not/in/config", testJSON(), testFallback()}, WantFallback()},
		{"Test fallback with empty config array", args{"/not/in/config", []byte("[]"), testFallback()}, WantFallback()},
		{"Test empty JSON", args{"/not/in/config", []byte(""), testFallback()}, WantError()},
		{"Test nil JSON", args{"/not/in/config", nil, testFallback()}, WantError()},
		{"Test bad JSON", args{"/not/in/config", testBadJSON(), testFallback()}, WantError()},
		{"Test bad empty single JSON", args{"/not/in/config", []byte("{}"), testFallback()}, WantError()},
		{"Test no fallback", args{"/not/in/config", testJSON(), nil}, WantNotFound()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JSONHandler(tt.args.json, tt.args.fallback)
			if err != nil {
				if !tt.want.GotError(err) {
					t.Errorf("JSONHandler() got %v, want %v",
						(&Result{ERROR, err.Error()}).String(), tt.want.String())
				}
				return
			}

			runHandlerTest(t, got, tt.args.url, tt.want)
		})
	}
}

func TestLoadHandlerFromFile(t *testing.T) {
	type args struct {
		filename string
		fallback http.Handler
	}
	tests := []struct {
		name string
		args args
		want Want
	}{
		{"Invalid File", args{"badPath", testFallback()}, WantError()},
		{"Valid JSON File Correct Extention", args{"valid_json.json", testFallback()}, WantRedirect("/got/json")},
		{"Valid YAML File Correct Extention", args{"valid_yaml.yaml", testFallback()}, WantRedirect("/got/yaml")},
		{"Valid JSON File Unknown Extention", args{"valid_json.txt", testFallback()}, WantRedirect("/got/json")},
		{"Valid YAML File Unknown Extention", args{"valid_yaml.txt", testFallback()}, WantRedirect("/got/yaml")},
		{"Invalid JSON File Correct Extention", args{"invalid_json.json", testFallback()}, WantErrorPrefix("unable to parse JSON formatted file:")},
		{"Invalid YAML File Correct Extention", args{"invalid_yaml.yaml", testFallback()}, WantErrorPrefix("unable to parse YAML formatted file:")},
		{"Invalid Unknown File", args{"unknown.txt", testFallback()}, WantErrorPrefix("unable to parse UNKNOWN formatted file:")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.args.filename) // relative path

			got, err := LoadHandlerFromFile(path, tt.args.fallback)
			if err != nil {
				if !tt.want.GotError(err) {
					t.Errorf("LoadHandlerFromFile() got %v, want %v",
						(&Result{ERROR, err.Error()}).String(), tt.want.String())
				}
				return
			}

			var url string
			switch {
			case strings.HasPrefix(tt.args.filename, "valid_json"):
				url = "/want-json"
			case strings.HasPrefix(tt.args.filename, "valid_yaml"):
				url = "/want-yaml"
			default:
				url = "/SHOULD-NOT-GET-HERE"
			}

			runHandlerTest(t, got, url, tt.want)
		})
	}
}

func Test_buildMap(t *testing.T) {
	type args struct {
		expansions []pathExpansion
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{"Test several", args{[]pathExpansion{{"a", "x"}, {"b", "y"}, {"c", "z"}}}, map[string]string{"a": "x", "b": "y", "c": "z"}},
		{"Test discard invalid", args{[]pathExpansion{{"a", ""}, {"b", "y"}, {"", "z"}, {"", ""}}}, map[string]string{"b": "y"}},
		{"Test only invalid", args{[]pathExpansion{{"a", ""}, {"", "z"}, {"", ""}}}, map[string]string{}},
		{"Test empty", args{[]pathExpansion{}}, map[string]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildMap(tt.args.expansions); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func runHandlerTest(t *testing.T, handler http.Handler, url string, want Want) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)
	outcome, value := getResults(recorder)

	if want.wantOutcome == FALLBACK {
		want.wantRedirectURL = url
	}
	if !want.Got(outcome, value) {
		t.Errorf(".ServeHTTP() got %v, want %v",
			(&Result{outcome, value}).String(), want.String())
	}
}

func getResults(response *httptest.ResponseRecorder) (Outcome, string) {
	switch status := response.Code; status {
	case http.StatusMovedPermanently:
		return REDIRECT, response.Header().Get("Location")
	case http.StatusOK:
		return FALLBACK, strings.TrimSpace(response.Body.String())
	case http.StatusNotFound:
		return NOTFOUND, strings.TrimSpace(response.Body.String())
	default:
		return Outcome(status), ""
	}
}

func testFallback() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, request.URL.String())
	})
}

func testPathsToUrls() map[string]string {
	return map[string]string{
		"/urlshort":           "/long/expansion/of/url",
		"/urlshort-godoc":     "https://godoc.org/github.com/gophercises/urlshort",
		"/bad-short":          "",
		"/urlshort-godoc-map": "https://godoc.org/github.com/gophercises/urlshort?field1=value+one&field2=value2#MapHandler",
		"":                    "/from/empty/path",
	}
}

func testYAML() []byte {
	return []byte(`
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url: https://github.com/gophercises/urlshort/tree/solution
`)
}

func testYAML2() []byte {
	return []byte(`
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url:
- path:
  url: /some/fallback/call
- path: /urlshort-no-url
`)
}

func testBadYAML() []byte {
	return []byte(`
- path: /urlshort
  	url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url: https://github.com/gophercises/urlshort/tree/solution
`)
}

func testJSON() []byte {
	return []byte(`
[
	{
		"path": "/urlshort",
		"url": "https://github.com/gophercises/urlshort"
	},
	{
		"path":"/urlshort-final",
		"url": "https://github.com/gophercises/urlshort/tree/solution"
	},
	{
		"path":"/redirect-local",
		"url": "local-short"
	},
	{
		"path":"/local-short",
		"url": "/some/fallback/call"
	}
]
`)
}

func testBadJSON() []byte {
	return []byte(`
[
	{
		"what": "/urlshort",
		"url": "https://github.com/gophercises/urlshort"
	},
]
`)
}
