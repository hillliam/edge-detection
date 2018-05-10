package main

import (
	"bytes"
	"fmt"
	"html/template"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
)

func servepage(w http.ResponseWriter) {
	t, _ := template.ParseFiles("main.html")
	data := 0
	t.Execute(w, data)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		servepage(w)
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
		buf := new(bytes.Buffer)
		png.Encode(buf, reply)
		w.Write(buf.Bytes())
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("*", handler)
	var port = os.Getenv("$PORT")
	if port == "" {
		port = "8080"
	}
	log.Print(http.ListenAndServe(":"+port, nil))
}
