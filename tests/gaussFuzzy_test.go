package tests

import (
	"github.com/1000Delta/imageserver/utils"
	"image/jpeg"
	"log"
	"os"
	"runtime"
	"testing"
)

const gaussSigma = 10
const coreRadius = 10

func TestGaussFuzzyMultiThreads(t *testing.T) {
	img, err := readImage("./test.jpg")
	if err != nil {
		t.Fatal(err.Error())
	}
	cpus := runtime.GOMAXPROCS(0)
	t.Logf("cpus: %d\n", cpus)
	nImg := utils.GaussFuzzy(img, coreRadius, cpus, gaussSigma)
	//nImg := GaussFuzzy(img, coreRadius, 1)
	nf, err := os.Create("./test_fuzzy_out_multi_threads.jpg")
	if err != nil {
		log.Fatal(err)
	}
	err = jpeg.Encode(nf, nImg, &jpeg.Options{Quality: 100})
	if err != nil {
		log.Fatal(err)
	}
}

func TestGaussFuzzy(t *testing.T) {
	img, err := readImage("./test.jpg")
	if err != nil {
		t.Fatal(err.Error())
	}
	cpus := 1
	t.Logf("cpus: %d", cpus)
	nImg := utils.GaussFuzzy(img, coreRadius, cpus, gaussSigma)
	nf, err := os.Create("./test_fuzzy_out.jpg")
	if err != nil {
		log.Fatal(err)
	}
	err = jpeg.Encode(nf, nImg, &jpeg.Options{Quality: 100})
	if err != nil {
		log.Fatal(err)
	}
}
