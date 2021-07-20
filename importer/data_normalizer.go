package importer

import (
	"math"
)

func FindMinMax(input []float64) [2]float64 {

	// find minmax in format {min, max}

	max := math.Pow(2, -15)
	min := math.Pow(2, 15)
	for _, value := range input {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	minmax := [2]float64{min, max}
	return minmax
}

func NormalizeData(input []float64) {

	// normalize dataset into (0,1)

	MinMax := FindMinMax(input)
	for i := 0; i < len(input); i++ {
		input[i] = (input[i] - MinMax[0]) / (MinMax[1] - MinMax[0])
	}
}
