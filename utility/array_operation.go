package utility

import "math"

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
