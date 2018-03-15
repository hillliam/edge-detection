// edge detection using 
package main

import "fmt" // for console I/O
import (
    "image"
    "image/png"
    "os"
)

func sobel_worker(get <-chan byte, set ->chan image) {

	for {
		letter := <-get
		fmt.Printf("%c", letter) // Report to console  NB could use "fmt.Printf ("%c",<-get)"
	}
}

//-- main process ----------------------------------------------------
func main() {

	    filename := "start.png"
        infile, err := os.Open(filename)
        if err != nil {
            // replace this with real error handling
            panic(err.Error())
        }
        defer infile.Close()

        // Decode will figure out what type of image is in the file on its own.
        // We just have to be sure all the image packages we want are imported.
        src, _, err := image.Decode(infile)
        if err != nil {
            // replace this with real error handling
            panic(err.Error())
        }

        // Create a new grayscale image
        bounds := src.Bounds()
		w, h := bounds.Max.X, bounds.Max.Y
		edge := image.NewGray(image.Rectangle{image.Point{0, 0}, image.Point{w, h}})

		
	c := make(chan Image)

	go sobel(c)

	for { // keep the 'main' process alive!
	}
	        // Encode the grayscale image to the output file
        outfilename := "done.png"
        outfile, err := os.Create(outfilename)
        if err != nil {
            // replace this with real error handling
            panic(err.Error())
        }
        defer outfile.Close()
        png.Encode(outfile, edge)
}