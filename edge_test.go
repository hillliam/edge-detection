package main

import (
	"image"
	"os"
	"testing"
)

func TestMesurments(t *testing.T) {
	infile, err := os.Open("256.png")
	handleerror(err)
	src, _, err := image.Decode(infile)
	handleerror(err)
	var w, h = getMesurments(src)
	if w != 256 && h != 256 {
		t.Errorf("mesurments are incorrect, got: %d X %d, want: %d X %d.", w, h, 256, 256)
	}
}

func Test256(t *testing.T) {
	infile, err := os.Open("256.png")
	handleerror(err)
	src, _, err := image.Decode(infile)
	handleerror(err)
	edgeDetection(src)

}

func Test512(t *testing.T) {
	infile, err := os.Open("start.png")
	handleerror(err)
	src, _, err := image.Decode(infile)
	handleerror(err)
	edgeDetection(src)
}

func Test1080(t *testing.T) {
	infile, err := os.Open("large.png")
	handleerror(err)
	src, _, err := image.Decode(infile)
	handleerror(err)
	edgeDetection(src)
}

func Test4k(t *testing.T) {
	infile, err := os.Open("3840.png")
	handleerror(err)
	src, _, err := image.Decode(infile)
	handleerror(err)
	edgeDetection(src)
}
