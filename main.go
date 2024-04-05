package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

func main() {
	args := os.Args
	files, err := os.ReadDir(args[1])
	k := 1 // temp
	d := 1 // temp
	histChan := make(chan Histogram)
	queryHistogramFilePath := "temp"

	if err != nil {
		panic(err)
	}

	var filenames []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".jpg") {
			filenames = append(filenames, file.Name())
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
	for _, s := range slices {
		wg.Add(1)
		go func(filenames []string) {
			defer wg.Done()
			computeHistograms(filenames, d, histChan)
		}(s)
	}

	queryHist, err := computeHistogram(queryHistogramFilePath, d)

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
}