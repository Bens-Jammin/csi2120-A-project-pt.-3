package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"math"
	"os"
	"sync"
)

type Histogram struct {
	fileName  string
	histogram []int
	// README: don't we normalize the histograms?
	// so the histogram should be an array of floats right ?
}

// from the 'readImage.go' file on the brightspace
func computeHistogram(imagePath string, depth int) (Histogram, error) {
	// Open the JPEG file
	file, err := os.Open(imagePath)
	if err != nil {
		return Histogram{"", nil}, err
	}
	defer file.Close()

	// Decode the JPEG image
	img, _, err := image.Decode(file)
	if err != nil {
		return Histogram{"", nil}, err
	}

	// Get the dimensions of the image
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Display RGB values for the first 5x5 pixels
	// remove y < 5 and x < 5  to scan the entire image
	for y := 0; y < height && y < 5; y++ {
		for x := 0; x < width && x < 5; x++ {

			// Convert the pixel to RGBA
			red, green, blue, _ := img.At(x, y).RGBA()
			// A color's RGBA method returns values in the range [0, 65535].
			// Shifting by 8 reduces this to the range [0, 255].
			red >>= 8
			blue >>= 8
			green >>= 8

			// Display the RGB values
			fmt.Printf("Pixel at (%d, %d): R=%d, G=%d, B=%d\n", x, y, red, green, blue)
		}
	}

	h := Histogram{imagePath, make([]int, depth)}
	return h, nil
}

func compareHistograms(h1 Histogram, h2 Histogram) float64 {

	var sum float64 = 0

	for i := 0; i < 255; i++ {

		curr_h1_val := float64(h1.histogram[i])
		curr_h2_val := float64(h2.histogram[i])

		sum += math.Min(curr_h1_val, curr_h2_val)
	}
	return sum
}

func computeHistograms(imagePaths []string, depth int, hChan chan<- Histogram) {

	var wg sync.WaitGroup

	for _, path := range imagePaths {
		wg.Add(1)

		go func(path string) {
			defer wg.Done()
			histogram, err := computeHistogram(path, depth)

			if err != nil {
				fmt.Printf("Error concurrently computing histograms for file %s: %p", path, err)
				return
			}
			hChan <- histogram
		}(path)
	}

	go func() {
		wg.Wait()
		close(hChan)
	}()
}
