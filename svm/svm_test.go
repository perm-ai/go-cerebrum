package svm

import (
	"fmt"
	"math"
	"testing"
)

func TestPackSvmTrainingData(t *testing.T) {
	testCaseX := make([][]float64, 10)
	testCaseY := []float64{1, -1, 1, -1, 1, -1, 1, -1, 1, -1}

	for i := range testCaseX {
		testCaseX[i] = []float64{1, 2, 3, 4, 5}
	}

	x, y := PackSvmTrainingData(testCaseX, testCaseY, int(math.Pow(2, 15)))

	fmt.Println(x[0: 18])
	fmt.Println(y[0: 18])
}
