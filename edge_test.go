package main

import (
	"image"
	"os"
	"testing"
)

func TestMesurments(t *testing.T) {
	infile, err := os.Open("256.png")
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer infile.Close()
	
	src, _, err := image.Decode(infile)
	if err != nil {
		t.Fatalf("Failed to decode image: %v", err)
	}
	
	w, h := getMesurments(src)
	if w != 256 || h != 256 {
		t.Errorf("Measurements are incorrect, got: %d x %d, want: %d x %d.", w, h, 256, 256)
	}
}

func TestEdgeDetection(t *testing.T) {
	tests := []struct {
		filename string
	}{
		{"256.png"},
		{"start.png"},
		{"large.png"},
		{"3840.png"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			infile, err := os.Open(tt.filename)
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer infile.Close()
			
			src, _, err := image.Decode(infile)
			if err != nil {
				t.Fatalf("Failed to decode image: %v", err)
			}
			
			edge := edgeDetection(src)
			if edge == nil {
				t.Errorf("Edge detection failed for file: %s", tt.filename)
			}
		})
	}
}
