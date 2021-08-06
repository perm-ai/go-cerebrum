package array

import (
	"math"
)

//add 2 arrays together, input arrays must have the same length
func AddArraysNew(a1 []float64, a2 []float64) []float64 {
	result := make([]float64, len(a1))
	if len(a1) != len(a2) {
		panic("Error, the arrays have unequal sizes.")
	} else {
		for i, p := range a1 {
			result[i] = a2[i] + p
		}
	}
	return result
}

//add 2 arrays together, input arrays must have the same length, the result will be stored in destination
func AddArrays(a1 []float64, a2 []float64, destination []float64) {
	if len(a1) != len(a2) {
		panic("Error, the arrays have unequal sizes.")
	} else {
		for i, p := range a1 {
			destination[i] = a2[i] + p
		}
	}
}

//add constant to every element in the array
func AddConstantNew(constant float64, a []float64) []float64 {
	result := make([]float64, len(a))
	for i, p := range a {
		result[i] = p + constant
	}
	return result
}

func AddConstant(constant float64, a []float64, destination []float64) {
	for i, p := range a {
		destination[i] = p + constant
	}
}

//subtract 2 arrays together(a1-a2), input arrays must have the same length
func SubtArraysNew(a1 []float64, a2 []float64) []float64 {
	result := make([]float64, len(a1))
	if len(a1) != len(a2) {
		panic("Error, the arrays have unequal sizes.")
	} else {
		for i, p := range a1 {
			result[i] = p - a2[i]
		}
	}
	return result
}

//subtract 2 arrays together(a1-a2), input arrays must have the same length, the result will be stored in destination
func SubtArrays(a1 []float64, a2 []float64, destination []float64) {
	if len(a1) != len(a2) {
		panic("Error, the arrays have unequal sizes.")
	} else {
		for i, p := range a1 {
			destination[i] = p - a2[i]
		}
	}
}

//multiply 2 arrays, input arrays must have the same length
func MulArraysNew(a1 []float64, a2 []float64) []float64 {
	result := make([]float64, len(a1))
	if len(a1) != len(a2) {
		panic("Error, the arrays have unequal sizes.")
	} else {
		for i, p := range a1 {
			result[i] = a2[i] * p
		}
	}
	return result
}

//multiply 2 arrays, input arrays must have the same length ,the result will be stored in destination
func MulArrays(a1 []float64, a2 []float64, destination []float64) {
	if len(a1) != len(a2) {
		panic("Error, the arrays have unequal sizes.")
	} else {
		for i, p := range a1 {
			destination[i] = a2[i] * p
		}
	}
}

//multiply array with a constant
func MulConstantArrayNew(constant float64, input []float64) []float64 {
	result := make([]float64, len(input))
	for i, p := range input {
		result[i] = p * constant
	}
	return result
}

//multiply array with a constant, the result will be stored in the input array
func MulConstantArray(constant float64, input []float64) {
	for i, p := range input {
		input[i] = p * constant
	}
}

//Output the total sum of the value in arrays
func SumElementArrays(input []float64) float64 {
	var result float64
	for _, p := range input {
		result += p
	}
	return result
}

//output a1 dot a2
func DotArrays(a1 []float64, a2 []float64) float64 {
	if len(a1) != len(a2) {
		panic("Error, the arrays have unequal sizes.")
	} else {
		return SumElementArrays(MulArraysNew(a1, a2))
	}
}

//input an array, output the sigmoid'd array
func SigmoidArray(input []float64) []float64 {
	output := make([]float64, len(input))
	for i, p := range input {
		output[i] = Sigmoidfloat(p)
	}
	return output
}

//input a float, output the sigmoid'd float
func Sigmoidfloat(input float64) float64 {
	return 1 / (1 + math.Pow(math.E, -1*input))
}

//Create plain array with "value" as every value inside the array with specified length
func GeneratePlainArray(value float64, length int) []float64 {
	array := make([]float64, length)
	for i := 0; i < length; i++ {
		array[i] = value
	}
	return array
}

func GenFilledArraysofArrays(value float64, size1 int, size2 int) [][]float64 {
	output := make([][]float64, size1)
	for i := 0; i < size1; i++ {
		for j := 0; j < size2; j++ {
			output[i] = append(output[i], 0.0)
		}
	}
	return output
}
