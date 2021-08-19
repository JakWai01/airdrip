package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

// Examples represents the examples loaded from examples.json
type Examples []*Example

var (
	errListExamples  = errors.New("failed to list examples (please run in the exampels folder)")
	errParseExamples = errors.New("faild to parse examples")
)

// Example represents an example loaded from examples.json.
type Example struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	Type        string `json:"type"`
	IsJS        bool
	IsWASM      bool
}

func main() {
	addr := flag.String("address", ":8080", "Address to host the HTTP server on.")
	flag.Parse()

	log.Println("Listening on", *addr)
	err := serve(*addr)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func serve(addr string) error {
	// Load the examples
	examples, err := getExamples()
	if err != nil {
		return err
	}

	// Load the templates
	homeTemplate := template.Must(template.ParseFiles("index.html"))

	// Serve the required pages
	// DIY 'mux' to avoid additional dependencies
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Path

		// Split up the URL. Expected parts:
		// 1: Base url
		// 2: "example"
		// 3: Example type: js or wasm
		// 4: Example folder, e.g.: data-channels
		// 5: Static file as part of the example
		parts := strings.Split(url, "/")

		for _, example := range *examples {
			if len(parts[4]) != 0 {
				http.StripPrefix("/example/js/data-channels/", http.FileServer(http.Dir("data-channels/jsfiddle"))).ServeHTTP(w, r)
				return
			}

			temp := template.Must(template.ParseFiles("example.html"))

			data := struct {
				*Example
				JS bool
			}{
				example,
				true,
			}

			err = temp.Execute(w, data)
			if err != nil {
				panic(err)
			}
			return
		}

		// Serve the main page
		err = homeTemplate.Execute(w, examples)
		if err != nil {
			panic(err)
		}
	})

	// Start the server
	return http.ListenAndServe(addr, nil)
}

// getExamples loads the examples from the examples.json file.
func getExamples() (*Examples, error) {
	file, err := os.Open("./examples.json")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errListExamples, err)
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			panic(closeErr)
		}
	}()

	var examples Examples
	err = json.NewDecoder(file).Decode(&examples)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errParseExamples, err)
	}

	examples[0].IsJS = true

	return &examples, nil
}
