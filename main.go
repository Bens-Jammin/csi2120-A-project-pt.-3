// Ben Miller - 300297574
// Arin Barak - 300280812
// CSI2120 Course Project - Pt. 3

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



func compareHistograms(h1 Histogram, h2 Histogram) float64 {

	var sum float64 = 0

	for i := 0; i < len(h1.histogram); i++ {

		curr_h1_val := float64(h1.histogram[i])
		curr_h2_val := float64(h2.histogram[i])

		sum += math.Min(curr_h1_val, curr_h2_val)
	}
	return sum
}



//###########################################################\\
//				    HISTOGRAM COMPUTATION
//###########################################################\\



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



func computeHistograms(imagePaths []string, depth int, hChan chan<- Histogram, n int) {
	for _, path := range imagePaths {
		histogram, err := computeHistogram(path, depth)

		if err != nil {
			fmt.Printf("Error concurrently computing histograms for file %s: %p", path, err)
			return
		}

		hChan <- histogram
	}
}


//###########################################################\\
//						MAIN FUNCTION
//###########################################################\\



func main() {
	
	os.Open("queryImages/q01.jpg")
	
	args := os.Args
	
	queryFilePath := args[1]
	dataset, err := os.ReadDir(args[2])
	
	if err != nil {
		panic(err)
	}
	
	// Step 1: create histogram channel
	k := 1048
	d := 3
	
	for t:=0;t<10;t++{
		histChan := make(chan Histogram, k)
		var wg sync.WaitGroup
		
		
		// Step 2: get the list of all image filenames in the dir
		var filenames []string
		for _, file := range dataset {
		
			if strings.HasSuffix(file.Name(), ".jpg") {
				filenames = append(filenames, filepath.Join(args[2], file.Name()))
			}
		}
		
		// Step 3: split the list into k slices
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
		
		start := time.Now()
		// Step 3½: send the slices to concurrently compute the histograms
		for i, s := range slices {
			wg.Add(1)
			go func(filenames []string) {
				defer wg.Done()
				computeHistograms(filenames, d, histChan, i)
				}(s)
		}

		// Step 4: compute the query histogram
		queryHist, err := computeHistogram(queryFilePath, d)
		if err != nil {
			panic(err)
		}
		
		go func() {
			wg.Wait()
			close(histChan)
		}()
		
		// Step 5: read the channel of histograms
		similarImages := make([]Histogram, 0, 5)
		for hist := range histChan {
			var r float64
			// don't count the query histogram
			if hist.fileName == queryHist.fileName {
				continue
			}
			// Step 5a: compare the histograms to the query when it's recieved
			r = compareHistograms(queryHist, hist)
			hist.relatedness = r
			similarImages = append(similarImages, hist)
		}

		// Step 5b (kinda): get the 5 most similar images
		sort.Slice(similarImages, func(i, j int) bool {
			return similarImages[i].relatedness > similarImages[j].relatedness
		})
		runtime := time.Since(start)

		// Step 6: print the list of the 5 most similar images
		fmt.Printf("Top 5 most related images to %s: (test # %d , k=%d)\n", filepath.Base(queryHist.fileName),t+1, k)
		for i := 0; i < 5; i++ {
			image := similarImages[i]
			fmt.Printf("%d: %s (%.2f)\n", i+1, filepath.Base(image.fileName), image.relatedness)
		}
		fmt.Printf("total runtime: %d ms", runtime.Milliseconds())
		}
}