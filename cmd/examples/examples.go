package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"strings"
)

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Path

		parts := strings.Split(url, "/")

		if len(parts[4]) != 0 {
			http.StripPrefix("/example/js/data-channels/", http.FileServer(http.Dir("data-channels/jsfiddle"))).ServeHTTP(w, r)
			return
		}

		temp := template.Must(template.ParseFiles("example.html"))

		var i interface{}
		err := temp.Execute(w, i)
		if err != nil {
			panic(err)
		}
		return
	})

	return http.ListenAndServe(addr, nil)
}
