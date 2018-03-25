package main

import (
	"fmt"
	"html/template"
	"image"
	"log"
	"net/http"
)

func servepage(w http.ResponseWriter) {
	t, _ := template.ParseFiles("main.html")
	t.Execute(w)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {

	} else {
		r.ParseForm()
		file, _, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		image, _, err := image.Decode(file)
		var reply = edgeDetection(image)

	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
