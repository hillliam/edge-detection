// edge detection using sobel algoritham
package main

import ( // for console I/O

	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// image support
// png loading

var sobelH = [9]int{-1, 0, 1, -2, 0, 2, -1, 0, 1} // both are applied to the pixels of the image
var sobelV = [9]int{1, 2, 1, 0, 0, 0, -1, -2, -1}

var movementx = [9]int{-1, 0, 1, -1, 0, 1, -1, 0, 1}
var movementy = [9]int{-1, -1, -1, 0, 0, 0, 1, 1, 1}

const threads = 4

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
		fmt.Printf("creating new pixel %d %d C: %d \n", image[4].X, image[4].Y, image[4].C)
		for i := 0; i < 9; i++ {
			var pixelvalue, _, _, _ = image[i].C.RGBA()
			var numberh = int(pixelvalue) * sobelH[i]
			var numberv = int(pixelvalue) * sobelV[i]
			gh = gh + numberh
			gv = gv + numberv
		}
		done.X = image[4].X
		done.Y = image[4].Y
		var newpixel = math.Sqrt(float64((gh * gh) + (gv * gv)))
		//if uint8(newpixel) > 200 {
		done.C = color.Gray{uint8(newpixel)}
		//} else {
		//	done.C = color.Gray{uint8(0)}
		//}
		fmt.Printf("made new pixel %d %d C: %d old C: %d \n", done.X, done.Y, done.C, image[4].C)
		set <- done
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
	fmt.Printf("image is %d by %d \n", w, h)
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
			var subimage [9]Pixel
			for i := 0; i < 9; i++ {
				var pixel Pixel
				var locationx = currentx + movementx[i] // -1 1 +1
				var locationy = currenty + movementy[i] // -1 1 +1
				/*switch i {
				case 0:
					locationx = locationx - 1
					locationy = locationy - 1
				case 1:
					locationy = locationy - 1
				case 2:
					locationx = locationx + 1
					locationy = locationy - 1
				case 3:
					locationx = locationx - 1
				}*/
				if locationx < 0 || locationy < 0 || locationx > w || locationy > h {
					pixel.C = color.Black // clamp edge of image
					subimage[i] = pixel
				} else {
					pixel.X = locationx
					pixel.Y = locationy
					pixel.C = src.At(locationx, locationy)
					subimage[i] = pixel
				}
			}
			//fmt.Printf("sending to worker \n")
			// send pixels
			var sent = false
			for sent != true {
				for i := 0; i < threads; i++ {
					if sent == false {
						select {
						case done := <-updated[i]:
							fmt.Printf("got pixel %d  %d c: %d from worker %d \n", done.X, done.Y, done.C, i)
							if done.X < 0 || done.Y < 0 { // ignore
								inputpixels[i] <- subimage
								sent = true
							} else {
								edge.Set(done.X, done.Y, done.C)
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
