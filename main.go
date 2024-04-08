package main

import (
	"fmt"
	"image"
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
	histogram   []int
	// README: don't we normalize the histograms?
	// so the histogram should be an array of floats right ?
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

	h := Histogram{0.0, imagePath, make([]int, depth)}
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

func main() {

	os.Open("queryImages/q01.jpg")

	args := os.Args

	queryFilePath := args[1]
	dataset, err := os.ReadDir(args[2])

	if err != nil {
		panic(err)
	}

	k_vals := []int{1, 2, 4, 16, 64, 156, 1048}
	d := 5
	histChan := make(chan Histogram)

	var filenames []string
	for _, file := range dataset {
		if strings.HasSuffix(file.Name(), ".jpg") {
			filenames = append(filenames, filepath.Join(args[2], file.Name()))
		}
	}

	for k := range k_vals {

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

		tests := make([]int64, 10)
		for i := 0; i < 10; i++ {
			start := time.Now()
			var wg sync.WaitGroup
			for _, s := range slices {
				wg.Add(1)
				go func(filenames []string) {
					defer wg.Done()
					computeHistograms(filenames, d, histChan)
				}(s)
			}

			queryHist, err := computeHistogram(queryFilePath, d)
			if err != nil {
				panic(err)
			}

			similarImages := make([]Histogram, 0, 5)
			for hist := range histChan {
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

			wg.Wait()

			runTime := time.Since(start)
			fmt.Printf("Test #%d Runtime: %d\n", (i + 1), runTime.Milliseconds())
			tests[i] = runTime.Milliseconds()
		}
		var sum, max, min int64
		var median float64
		sum = 0
		for n := 0; n < 10; n++ {
			sum += tests[n]
		}

		sort.Slice(tests, func(i, j int) bool {
			return tests[i] < tests[j]
		})
		max = tests[9]
		min = tests[0]
		median = float64((tests[4] + tests[5]) / 2)

		fmt.Println("TEST STATISTICS:")
		fmt.Printf("k=%d\n  max runtime=%d\n  min runtime=%d\n  median runtime:%f\n  average runtime: %f", k, max, min, median, float64(sum/10))
	}
}
