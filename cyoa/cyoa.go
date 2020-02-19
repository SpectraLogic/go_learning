package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	filepath := "gopher.json"
	story, err := ParseStoryFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("%#v\n", story)
	s, _ := json.MarshalIndent(story, "", "  ")
	fmt.Println(string(s))

	t, err := template.ParseFiles("story.html")
	if err != nil {
		log.Fatal(err)
	}
	handler := StoryHandler{
		Template: t,
		Story:    story,
	}
	log.Fatal(http.ListenAndServe(":8080", handler))

}

type StoryHandler struct {
	Template *template.Template
	Story    map[string]StoryArc
}

func (sh StoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// http.Error(w, "Template execution not implemented", http.StatusNotImplemented)
	arc := "intro"
	path := r.URL.Path
	if path != "/" {
		arc = path[1:]
	}
	storyArc, ok := sh.Story[arc]
	if !ok {
		//  http.StatusInternalServerError
		msg := fmt.Sprintf("Internal error: Story does not have an \"%v\" arc", arc)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	log.Printf("Visting arc %v \n", arc)

	err := sh.Template.Execute(w, storyArc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ParseStoryFile(filepath string) (map[string]StoryArc, error) {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var story map[string]StoryArc
	err = json.Unmarshal(bytes, &story)
	if err != nil {
		return nil, err
	}
	return story, nil
}

type Option struct {
	Text string `json:"text"`
	Arc  string `json:"arc"`
}

type StoryArc struct {
	Title   string   `json:"title"`
	Story   []string `json:"story"`
	Options []Option `json:"options"`
}
