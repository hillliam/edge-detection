package main

import (
	"bytes"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"net/http"
	"os"
	"runtime" // Import the runtime package
	"sync"
	"time"
)

// Sobel filter kernels
var sobelH = [9]int{-1, 0, 1, -2, 0, 2, -1, 0, 1}
var sobelV = [9]int{1, 2, 1, 0, 0, 0, -1, -2, -1}

// Pixel movement offsets
var movementx = [9]int{-1, 0, 1, -1, 0, 1, -1, 0, 1}
var movementy = [9]int{-1, -1, -1, 0, 0, 0, 1, 1, 1}

// Image files to process
var images = [4]string{"256.png", "start.png", "large.png", "3840.png"}

var threads = runtime.NumCPU() // Set threads to the number of CPU cores
const runs = 3

// Pixel represents a single pixel in the image
type Pixel struct {
	X, Y int
	C    color.Color
}

// servepage serves the main HTML page
func servepage(w http.ResponseWriter) {
	t, err := template.ParseFiles("main.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}
	data := 0
	t.Execute(w, data)
}

// handler handles HTTP requests for image processing
func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		servepage(w)
	} else {
		r.ParseForm()
		file, _, err := r.FormFile("uploadfile")
		if err != nil {
			log.Println("Error getting form file:", err)
			return
		}
		defer file.Close()
		img, _, err := image.Decode(file)
		if err != nil {
			log.Println("Error decoding image:", err)
			return
		}
		reply := edgeDetection(img)
		buf := new(bytes.Buffer)
		png.Encode(buf, reply)
		w.Write(buf.Bytes())
	}
}

// main starts the HTTP server
func main() {
	log.Print("server started")
	http.HandleFunc("/", handler)
	http.HandleFunc("/*", handler)
	port := os.Getenv("PORT")
	if port == "" {
		log.Print("default port is 8080")
		port = "8080"
	}
	log.Print("server port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// sobelWorker applies the Sobel filter to a set of pixels
func sobelWorker(get <-chan [9]Pixel, set chan<- Pixel, wg *sync.WaitGroup) {
	defer wg.Done()
	for image := range get {
		var gh, gv int
		for i := 0; i < 9; i++ {
			pixelvalue, _, _, _ := image[i].C.RGBA()
			realpixelvalue := uint8(pixelvalue)
			gh += int(realpixelvalue) * sobelH[i]
			gv += int(realpixelvalue) * sobelV[i]
		}
		newpixel := uint32(math.Ceil(math.Sqrt(float64((gh * gh) + (gv * gv)))))
		set <- Pixel{X: image[4].X, Y: image[4].Y, C: color.Gray{uint8(newpixel)}}
	}
}

// getMesurments returns the width and height of the image
func getMesurments(src image.Image) (w, h int) {
	bounds := src.Bounds()
	w, h = bounds.Max.X, bounds.Max.Y
	fmt.Printf("image is %d by %d \n", w, h)
	return
}

// getpixels returns a 3x3 grid of pixels around the current pixel
func getpixels(src image.Image, currentx, currenty, w, h int) [9]Pixel {
	var subimage [9]Pixel
	for i := 0; i < 9; i++ {
		locationx := currentx + movementx[i]
		locationy := currenty + movementy[i]
		if locationx < 0 || locationy < 0 || locationx >= w || locationy >= h {
			subimage[i] = Pixel{C: color.Black}
		} else {
			subimage[i] = Pixel{X: locationx, Y: locationy, C: src.At(locationx, locationy)}
		}
	}
	return subimage
}

func calculateNumberOfChanels(w, h int) int {
	// get total number of pixels
	totalPixels := w * h
	// devide by 9 to get the number of 3x3 grids
	return totalPixels / 9
}

// edgeDetection applies the Sobel filter to the entire image
func edgeDetection(src image.Image) image.Image {
	w, h := getMesurments(src)
	maxChanels := calculateNumberOfChanels(w, h)
	// if maxChanels is less then threads, set threads to maxChanels
	if maxChanels < threads {
		threads = maxChanels
	}
	threads = 1
	edge := image.NewGray(image.Rectangle{image.Point{0, 0}, image.Point{w, h}})
	inputpixels := make([]chan [9]Pixel, threads)
	updated := make([]chan Pixel, threads)
	var wg sync.WaitGroup

	for i := 0; i < threads; i++ {
		inputpixels[i] = make(chan [9]Pixel)
		updated[i] = make(chan Pixel)
		wg.Add(1)
		go sobelWorker(inputpixels[i], updated[i], &wg)
	}

	go func() {
		for currentx := 0; currentx < w; currentx++ {
			for currenty := 0; currenty < h; currenty++ {
				subimage := getpixels(src, currentx, currenty, w, h)
				inputpixels[currentx%threads] <- subimage
			}
		}
		for i := 0; i < threads; i++ {
			close(inputpixels[i])
		}
	}()

	go func() {
		wg.Wait()
		for i := 0; i < threads; i++ {
			close(updated[i])
		}
	}()

	for i := 0; i < threads; i++ {
		for pixel := range updated[i] {
			edge.Set(pixel.X, pixel.Y, pixel.C)
		}
	}

	return edge
}

// handleerror logs an error if it is not nil
func handleerror(err error) {
	if err != nil {
		log.Println("Error:", err)
	}
}

// edgerun processes an image file and writes the result to an output file
func edgerun(filename, outfilename string, a, b int) {
	infile, err := os.Open(filename)
	handleerror(err)
	defer infile.Close()

	Results, err := os.OpenFile("Results.txt", os.O_APPEND|os.O_CREATE, 0666)
	handleerror(err)
	defer Results.Close()

	src, _, err := image.Decode(infile)
	handleerror(err)

	startTimer := time.Now()
	edge := edgeDetection(src)
	elapsedTime := time.Since(startTimer)

	w, h := getMesurments(src)
	simpleThroughput := (float64(w*h*18) / (elapsedTime.Seconds() / 1000.0) / 1000000000.0)
	stperline := simpleThroughput / 6

	fmt.Println("elapsed time:", elapsedTime)
	fmt.Println("go throughput:", simpleThroughput)
	fmt.Println("go throughput per line:", stperline)

	_, _ = Results.WriteString(fmt.Sprintf("run %d ", (a + 1)))
	_, _ = Results.WriteString(fmt.Sprintf("\nimage is %d by %d ", w, h))
	_, _ = Results.WriteString(fmt.Sprintf("\nelapsed time: %s", elapsedTime))
	_, _ = Results.WriteString(fmt.Sprintf("\ngo throughput: %f ", simpleThroughput))
	_, _ = Results.WriteString(fmt.Sprintf("\ngo throughput per line: %f \n\n", stperline))

	outfile, err := os.Create(outfilename)
	handleerror(err)
	defer outfile.Close()
	png.Encode(outfile, edge)
}

// maintest runs the edge detection on a set of images
func maintest() {
	for b := 0; b < 4; b++ {
		for a := 0; a < runs; a++ {
			edgerun(images[b], "done"+images[b], a, b)
		}
	}
}
