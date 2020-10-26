// edge detection using sobel algoritham
package main

import ( // for console I/O

	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"time"
)

// image support
// png loading

var sobelH = [9]int{-1, 0, 1, -2, 0, 2, -1, 0, 1} // both are applied to the pixels of the image
var sobelV = [9]int{1, 2, 1, 0, 0, 0, -1, -2, -1}

var movementx = [9]int{-1, 0, 1, -1, 0, 1, -1, 0, 1}

var movementy = [9]int{-1, -1, -1, 0, 0, 0, 1, 1, 1}

var images = [4]string{"256.png", "start.png", "large.png", "3840.png"}

const threads = 7
const runs = 3

var startTimer = time.Now()

type Pixel struct { // store pixel data
	X, Y int         // location
	C    color.Color // pixel value
}

func sobelWorker(get <-chan [9]Pixel, set chan<- Pixel) {
	var done Pixel
	done.C = color.Black
	set <- done
	for {
		var image [9]Pixel
		//for i := 0; i < 9; i++ { // get pixels out of array
		//var input = <-get
		image = <-get
		//}
		var gh = 0
		var gv = 0
		//fmt.Printf("creating new pixel %d %d C: %d \n", image[4].X, image[4].Y, image[4].C)
		for i := 0; i < 9; i++ {
			var pixelvalue, _, _, _ = image[i].C.RGBA() // get pixel value from image
			realpixelvalue := uint8(pixelvalue)         // lower value to reduce noise
			var numberh = int(realpixelvalue) * int(sobelH[i])
			var numberv = int(realpixelvalue) * int(sobelV[i])
			gh = gh + numberh
			gv = gv + numberv
		}
		done.X = image[4].X
		done.Y = image[4].Y

		var newpixel = uint32(math.Ceil(math.Sqrt(float64((gh * gh) + (gv * gv))))) // calculate new pixel value
		done.C = color.Gray{uint8(newpixel)}
		//fmt.Printf("made new pixel %d %d C: %d old C: %d \n", done.X, done.Y, done.C, image[4].C)
		set <- done // send new pixel to image
	}
}

func getMesurments(src image.Image) (w, h int) {
	bounds := src.Bounds()
	w, h = bounds.Max.X, bounds.Max.Y
	fmt.Printf("image is %d by %d \n", w, h)
	return
}

func getpixels(src image.Image, currentx, currenty, w, h int) (subimage [9]Pixel) {
	for i := 0; i < 9; i++ {
		var pixels Pixel
		var locationx = currentx + movementx[i] // -1 1 +1
		var locationy = currenty + movementy[i] // -1 1 +1
		if locationx < 0 || locationy < 0 || locationx > w || locationy > h {
			pixels.C = color.Black // clamp edge of image
			subimage[i] = pixels
		} else {
			pixels.X = locationx
			pixels.Y = locationy
			pixels.C = src.At(locationx, locationy)
			subimage[i] = pixels
		}
	}
	return
}

func edgeDetection(src image.Image) (edge image.Image) {
	w, h := getMesurments(src)
	edge = image.NewGray(image.Rectangle{image.Point{0, 0}, image.Point{w, h}})
	inputpixels := make([]chan [9]Pixel, threads)
	updated := make([]chan Pixel, threads)

	for i := 0; i < threads; i++ {
		inputpixels[i] = make(chan [9]Pixel)
		updated[i] = make(chan Pixel)
	}
	for i := 0; i < threads; i++ {
		go sobelWorker(inputpixels[i], updated[i])
	}

	for currentx := 0; currentx < w; currentx++ { //handle sending parts of the image
		for currenty := 0 - threads; currenty < h; currenty++ {
			//fmt.Printf("starting pixel %d  %d \n", currentx, currenty)
			//get next part of image
			var subimage = getpixels(src, currentx, currenty, w, h)
			//fmt.Printf("sending to worker \n")
			// send pixels
			var sent = false
			for !sent {
				for i := 0; i < threads; i++ {
					if !sent {
						select {
						case done := <-updated[i]:
							//fmt.Printf("got pixel %d  %d c: %d from worker %d \n", done.X, done.Y, done.C, i)
							if done.X < 0 || done.Y < 0 { // ignore
								inputpixels[i] <- subimage
								sent = true
							} else {
								edge.(*image.Gray).Set(done.X, done.Y, done.C)
								inputpixels[i] <- subimage
								sent = true
								//fmt.Printf("%d \n", done)
							}
						default:
						}
					}
				}
			}
		}
	}
	return
}

func handleerror(err error) {
	if err != nil {
		// replace this with real error handling
		panic(err.Error())
	}
}

func edgerun(filename, outfilename string, a, b int) {
	infile, err := os.Open(filename)
	handleerror(err)
	Results, err := os.OpenFile("Results.txt", os.O_APPEND|os.O_CREATE, 0666)
	handleerror(err)
	defer infile.Close()
	defer Results.Close()
	// Decode will figure out what type of image is in the file on its own.
	// We just have to be sure all the image packages we want are imported.
	src, _, err := image.Decode(infile)
	handleerror(err)
	// Create a new grayscale image

	startTimer = time.Now() // start of real work
	// execute algoritham

	edge := edgeDetection(src)

	elapsedTime := time.Since(startTimer)
	var w, h = getMesurments(src)
	var simpleThroughput = ((float64)(w*h*18) / (elapsedTime.Seconds() / 1000.0) / 1000000000.0)
	var stperline = simpleThroughput / 6
	fmt.Println("elapsed time: ", elapsedTime)
	fmt.Println("go throughput: ", simpleThroughput)
	fmt.Println("go throughput per line: ", stperline)
	_, _ = Results.WriteString(fmt.Sprintf("run %d ", (a + 1)))
	_, _ = Results.WriteString(fmt.Sprintf("\nimage is %d by %d ", w, h))
	_, _ = Results.WriteString(fmt.Sprintf("\nelapsed time: %s", elapsedTime))
	_, _ = Results.WriteString(fmt.Sprintf("\ngo throughput: %f ", simpleThroughput))
	_, _ = Results.WriteString(fmt.Sprintf("\ngo throughput per line: %f \n\n", stperline))
	// Encode the grayscale image to the output file
	outfile, err := os.Create(outfilename)
	handleerror(err)
	defer outfile.Close()
	png.Encode(outfile, edge)
}

//-- main process ----------------------------------------------------
func maintest() {
	for b := 0; b < 4; b++ {
		for a := 0; a < runs; a++ {
			edgerun(images[b], "done"+images[b], a, b)
		}
	}
}
