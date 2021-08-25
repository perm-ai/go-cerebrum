package layers

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

type AveragePooling2D struct{
	utils 		utility.Utils
	InputSize	[]int
	Size		[]int
	Strides 	[]int
}

func NewPoolingLayer(utils utility.Utils, inputSize []int, poolingSize []int, strides []int) AveragePooling2D {

	return AveragePooling2D{utils, inputSize, poolingSize, strides}

}

func (p AveragePooling2D) GetOutputSize() []int{

	row := int(float64(p.InputSize[0] - p.Size[0]) / float64(p.Strides[0])) + 1
	column := int(float64(p.InputSize[1] - p.Size[1]) / float64(p.Strides[1])) + 1

	return []int{row, column, p.InputSize[2]}

}

func (p AveragePooling2D) Forward(input [][][]*ckks.Ciphertext) [][][]*ckks.Ciphertext {

	currentOutRow := 0
	outputSize := p.GetOutputSize()
	output := make([][][]*ckks.Ciphertext, outputSize[0])

	// Loop through each input datapoint that corresponds to the first row of the pooling filter
	for row := 0; row <= p.InputSize[0] - p.Size[0]; row += p.Strides[0]{

		currentOutCol := 0
		output[currentOutRow] = make([][]*ckks.Ciphertext, outputSize[1])

		// Loop through each input datapoint that corresponds to the first column of the pooling filter
		for column := 0; column <= p.InputSize[1] - p.Size[1]; column += p.Strides[1]{

			output[currentOutRow][currentOutCol] = make([]*ckks.Ciphertext, outputSize[2])

			// Loop through each depth of input
			for depth := 0; depth < p.InputSize[2]; depth++{

				// Compute pooling sum
				for poolRow := 0; poolRow < p.Size[0]; poolRow++{
					for poolCol := 0; poolCol < p.Size[1]; poolCol++{

						if output[currentOutRow][currentOutCol][depth] == nil{
							output[currentOutRow][currentOutCol][depth] = input[row + poolRow][column + poolCol][depth]
						} else {
							p.utils.Add(*output[currentOutRow][currentOutCol][depth], *input[row + poolRow][column + poolCol][depth], output[currentOutRow][currentOutCol][depth])
						}

					}
				}

				// Compute pooling average
				averager := p.utils.EncodePlaintextFromArray(p.utils.GenerateFilledArray(1.0 / float64(p.Size[0] * p.Size[1])))
				p.utils.MultiplyPlain(output[currentOutRow][currentOutCol][depth], &averager, output[currentOutRow][currentOutCol][depth], true, false)

			}

			// Increment current output column
			currentOutCol++

		}

		// Increment current output row
		currentOutRow++

	}

	return output

}

func (p AveragePooling2D) Backward(gradient [][][]*ckks.Ciphertext) [][][]*ckks.Ciphertext {

	gradientSize := p.GetOutputSize()

	// Divide each gradient by output size
	divider := p.utils.EncodePlaintextFromArray(p.utils.GenerateFilledArray(1.0 / float64(gradientSize[0] * gradientSize[1])))
	for row := range gradient{
		for col := range gradient[row]{
			for depth := range gradient[row][col]{
				p.utils.MultiplyPlain(gradient[row][col][depth], &divider, gradient[row][col][depth], true, false)
			}
		}
	}

	currentGradRow := 0
	upSampledGradient := make([][][]*ckks.Ciphertext, p.InputSize[0])

	// Loop through each input datapoint that corresponds to the first row of the pooling filter
	for row := 0; row <= p.InputSize[0] - p.Size[0]; row += p.Strides[0]{

		currentGradCol := 0

		// Loop through each input datapoint that corresponds to the first column of the pooling filter
		for column := 0; column <= p.InputSize[1] - p.Size[1]; column += p.Strides[1]{

			for poolRow := 0; poolRow < p.Size[0]; poolRow++{

				// Make row slice if undeclared
				if upSampledGradient[row + poolRow] == nil{
					upSampledGradient[row + poolRow] = make([][]*ckks.Ciphertext, p.InputSize[1])
				}
				
				for poolCol := 0; poolCol < p.Size[1]; poolCol++{

					// Make column slice if undeclared
					if upSampledGradient[row + poolRow][column + poolCol] == nil{
						upSampledGradient[row + poolRow][column + poolCol] = make([]*ckks.Ciphertext, p.InputSize[2])
					}

					// Loop through each depth of input
					for depth := 0; depth < p.InputSize[2]; depth++{

						if upSampledGradient[row+poolRow][column+poolCol][depth] == nil{
							upSampledGradient[row+poolRow][column+poolCol][depth] = gradient[currentGradRow][currentGradCol][depth]
						} else {
							p.utils.Add(*upSampledGradient[row+poolRow][column+poolCol][depth], *gradient[currentGradRow][currentGradCol][depth], upSampledGradient[row+poolRow][column+poolCol][depth])
						}

					}

				}
			}

			// Increment current gradient column
			currentGradCol++

		}

		// Increment current gradient row
		currentGradRow++

	}

	return upSampledGradient

}