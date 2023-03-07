package layers

import "github.com/tuneinsight/lattigo/v4/rlwe"


type Flatten2D struct {
	InputSize 	[]int
	OutputSize 	int
}

func NewFlatten(inputSize []int) Flatten2D{

	outputSize := 0
	
	// Calculate output size
	for i := range inputSize{
		if i == 0{
			outputSize = inputSize[i]
		} else {
			outputSize *= inputSize[i]
		}
	}

	return Flatten2D{inputSize, outputSize}
}

func (f Flatten2D) Forward(input [][][]*rlwe.Ciphertext) Output1d {

	output := make([]*rlwe.Ciphertext, f.OutputSize)

	for r := range input{
		for c := range input[r]{
			for d := range input[r][c]{
				output[(r * f.InputSize[0]) + (c * f.InputSize[1]) + (d * f.InputSize[2])] = input[r][c][d]
			}
		}
	}

	return Output1d{Output: output}

}

func (f Flatten2D) Backward(output []*rlwe.Ciphertext) Gradient2d{

	gradient := make([][][]*rlwe.Ciphertext, f.InputSize[0])

	for r := range gradient{
		gradient[r] = make([][]*rlwe.Ciphertext, f.InputSize[1])
		for c := range gradient[r]{
			gradient[r][c] = make([]*rlwe.Ciphertext, f.InputSize[2])
			for d := range gradient[r][c]{
				gradient[r][c][d] = output[(r * f.InputSize[0]) + (c * f.InputSize[1]) + (d * f.InputSize[2])]
			}
		}
	}

	return Gradient2d{InputGradient: gradient}

}

func (f Flatten2D) GetOutputSize() int {
	return f.OutputSize
}

func (f Flatten2D) IsTrainable() bool {
	return false
}