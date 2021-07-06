package utility

import (
	"math/rand"
	"time"
)

func GenerateLinearData(NumberOfData int) ([]float64, []float64, []int) {
	rand.Seed(time.Now().UnixNano())
	m := 1.0
	c := 0.2
	Error1 := 0
	Error0 := 0
	var x = make([]float64, NumberOfData)
	var y = make([]float64, NumberOfData)
	var target = make([]int, NumberOfData)
	for i := 0; i < NumberOfData; i++ {
		x[i] = rand.Float64()
		x := x[i]
		y[i] = rand.Float64()
		y := y[i]
		if y >= m*(x)+c || Error1%75 == 0 {
			target[i] = 1
			Error1++
		} else if y < m*(x)+c || Error0%75 == 0 {
			target[i] = 0
			Error0++
		} else {
			target[i] = 1
		}
	}
	return x, y, target
}
