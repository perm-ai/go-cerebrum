package utility

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateLinearData(NumberOfData int) ([]float64, []float64, []float64) {
	//This function will create (x,y) and a binary as an output
	//The binary depends on (x,y)
	rand.Seed(time.Now().UnixNano())
	m := 1 - (rand.Float64() * 0.5)
	c := rand.Float64() * 0.3
	Error1 := 0
	Error0 := 0
	var x = make([]float64, NumberOfData)
	var y = make([]float64, NumberOfData)
	var target = make([]float64, NumberOfData)
	for i := 0; i < NumberOfData; i++ {
		x[i] = rand.Float64()
		x := x[i]
		y[i] = rand.Float64()
		y := y[i]
		//Plot the point(x,y) and if its above the line then target = 1
		//The code sometimes add a little noise in the data(eg. once every 75 data)
		if (y >= m*(x)+c || Error1%20 == 0) && Error0%20 != 0 {
			target[i] = 1
			Error1++
		} else {
			target[i] = 0
			Error0++
		}
	}
	fmt.Printf("y = %fx + %f \n", m, c)
	fmt.Printf(" DATA -> 1:%o 0:%o \n", Error1, Error0)
	return x, y, target
}
