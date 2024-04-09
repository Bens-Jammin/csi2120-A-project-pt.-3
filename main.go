package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Histogram struct {
	relatedness float64
	fileName    string
	histogram   []float64
}

// from the 'readImage.go' file on the brightspace
func computeHistogram(imagePath string, depth int) (Histogram, error) {
	// Open the JPEG file
	file, err := os.Open(imagePath)
	if err != nil {
		return Histogram{0.0, "", nil}, err
	}
	defer file.Close()

	// Decode the JPEG image
	img, _, err := image.Decode(file)
	if err != nil {
		return Histogram{0.0, "", nil}, err
	}

	// Get the dimensions of the image
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	histo_size := int(math.Pow(2, float64(3*depth)))
	histogram := make([]int, histo_size)

	// Display RGB values for the first 5x5 pixels
	// remove y < 5 and x < 5  to scan the entire image
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {

			// Convert the pixel to RGBA
			red, green, blue, _ := img.At(x, y).RGBA()
			// A color's RGBA method returns values in the range [0, 65535].
			// Shifting by 8 reduces this to the range [0, 255].
			red >>= 8
			blue >>= 8
			green >>= 8

			shifted_red := red >> (8-depth)
			shifted_blue := blue >> (8-depth)
			shifted_green := green >> (8-depth)

			idx := (shifted_red << (2*depth) ) + ( shifted_green << depth ) + shifted_blue 
			histogram[idx] += 1
		}
	}

	normalized_histogram := make([]float64, histo_size)

	for i:=0; i<histo_size;i++ {
		normalized_histogram[i] = float64(histogram[i])/float64(width*height)
	}

	h := Histogram{0.0, imagePath, normalized_histogram}
	return h, nil
}

func compareHistograms(h1 Histogram, h2 Histogram) float64 {

	var sum float64 = 0

	for i := 0; i < len(h1.histogram); i++ {

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

func main() {

	os.Open("queryImages/q01.jpg")

	args := os.Args

	queryFilePath := args[1]
	dataset, err := os.ReadDir(args[2])

	if err != nil {
		panic(err)
	}

	k := 1
	d := 3
	

	var filenames []string
	for _, file := range dataset {
		if strings.HasSuffix(file.Name(), ".jpg") {
			filenames = append(filenames, filepath.Join(args[2], file.Name()))
		}
	}

	sliceSize := len(filenames) / k
	slices := make([][]string, k)
	for i := 0; i < k; i++ {
		start := i * sliceSize
		end := start + sliceSize
		if i == k-1 {
			end = len(filenames)
		}
		slices[i] = filenames[start:end]
	}

	var wg sync.WaitGroup
	start := time.Now()
	histChan := make(chan Histogram)

	for _, s := range slices {
		wg.Add(1)
		go func(filenames []string) {
			defer wg.Done()
			computeHistograms(filenames, d, histChan)
		}(s)

		queryHist, err := computeHistogram(queryFilePath, d)
		if err != nil {
			panic(err)
		}

		similarImages := make([]Histogram, 0, 5)
		for hist := range histChan {
			var r float64
			// don't count the query histogram
			if hist.fileName == queryHist.fileName {
				continue
			}
			r = compareHistograms(queryHist, hist)
			hist.relatedness = r
			similarImages = append(similarImages, hist)
		}

		sort.Slice(similarImages, func(i, j int) bool {
			return similarImages[i].relatedness > similarImages[j].relatedness
		})

		fmt.Printf("Top 5 most related images to %s:\n", queryHist.fileName)
		for i := 0; i < 5; i++ {
			image := similarImages[i]
			fmt.Printf("%d: %s (%.2f)\n", i+1, image.fileName, image.relatedness)
		}
		runtime := time.Since(start)
		fmt.Printf("total runtime: %d ms", runtime.Milliseconds())
	}
}