package layers

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

type AveragePooling2D struct {
	utils      utility.Utils
	InputSize  []int
	Size       []int
	Strides    []int
	btspOutput []bool
}

func NewPoolingLayer(utils utility.Utils, inputSize []int, poolingSize []int, strides []int) AveragePooling2D {

	return AveragePooling2D{utils, inputSize, poolingSize, strides, []bool{false, false}}

}

func (p AveragePooling2D) GetOutputSize() []int {

	row := int(float64(p.InputSize[0]-p.Size[0])/float64(p.Strides[0])) + 1
	column := int(float64(p.InputSize[1]-p.Size[1])/float64(p.Strides[1])) + 1

	return []int{row, column, p.InputSize[2]}

}

func (p AveragePooling2D) Forward(input [][][]*ckks.Ciphertext) Output2d {

	currentOutRow := 0
	outputSize := p.GetOutputSize()
	output := make([][][]*ckks.Ciphertext, outputSize[0])
	outputChannels := make([]chan [][]*ckks.Ciphertext, outputSize[0])
	averager := p.utils.EncodePlaintextFromArray(p.utils.GenerateFilledArray(1.0 / float64(p.Size[0]*p.Size[1])))

	// Loop through each input datapoint that corresponds to the first row of the pooling filter
	for row := 0; row <= p.InputSize[0]-p.Size[0]; row += p.Strides[0] {

		outputChannels[row] = make(chan [][]*ckks.Ciphertext)

		go func(rowIndex int, outputChannel chan [][]*ckks.Ciphertext) {

			// Create array of channels for sending array of depth in each column in a concurrent operations
			outputColumnChannels := make([]chan []*ckks.Ciphertext, outputSize[1])

			// Loop through each input datapoint that corresponds to the first column of the pooling filter
			for column := 0; column <= p.InputSize[1]-p.Size[1]; column += p.Strides[1] {

				outputColumnChannels[column] = make(chan []*ckks.Ciphertext)

				go func(rowIndex int, colIndex int, outputColumnChannel chan []*ckks.Ciphertext) {

					outputDepthChannels := make([]chan *ckks.Ciphertext, outputSize[2])

					// Loop through each depth of input
					for depth := 0; depth < p.InputSize[2]; depth++ {

						outputDepthChannels[depth] = make(chan *ckks.Ciphertext)

						go func(rowIndex int, colIndex int, depIndex int, utils utility.Utils, outputDepthChannel chan *ckks.Ciphertext) {

							var poolResult *ckks.Ciphertext

							// Compute pooling sum
							for poolRow := 0; poolRow < p.Size[0]; poolRow++ {
								for poolCol := 0; poolCol < p.Size[1]; poolCol++ {

									if poolResult == nil {
										poolResult = input[rowIndex+poolRow][colIndex+poolCol][depIndex]
									} else {
										utils.Add(*poolResult, *input[rowIndex+poolRow][colIndex+poolCol][depIndex], poolResult)
									}

								}
							}

							// Compute pooling average
							utils.MultiplyPlain(poolResult, &averager, poolResult, true, false)

							// Return avg pooling result from that pool
							outputDepthChannel <- poolResult

						}(rowIndex, colIndex, depth, p.utils.CopyWithClonedEval(), outputDepthChannels[depth])

					}

					outputColumn := make([]*ckks.Ciphertext, outputSize[2])

					for depth := range outputDepthChannels {
						outputColumn[depth] = <-outputDepthChannels[depth]
					}

					if p.btspOutput[0] {
						p.utils.Bootstrap1dInPlace(outputColumn, true)
					}

					outputColumnChannel <- outputColumn

				}(rowIndex, column, outputColumnChannels[column])

			}

			outputRow := make([][]*ckks.Ciphertext, outputSize[1])

			// Capture and store value of columns in this row in to a row array
			for col := range outputColumnChannels {
				outputRow[col] = <-outputColumnChannels[col]
			}

			// Send row array back through channel
			outputChannel <- outputRow

		}(row, outputChannels[row])

		// Increment current output row
		currentOutRow++

	}

	for row := range outputChannels {

		output[row] = <-outputChannels[row]

	}

	return Output2d{Output: output}

}

// Calculate loss gradient wrt input of a pooling layer
// input and output params aren't used and can be nil
func (p AveragePooling2D) Backward(input [][][]*ckks.Ciphertext, output [][][]*ckks.Ciphertext, gradient [][][]*ckks.Ciphertext, hasPrevLayer bool) Gradient2d {

	gradientSize := p.GetOutputSize()

	// ===================================================
	// ======= Divide each gradient by output size =======
	// ===================================================
	divider := p.utils.EncodePlaintextFromArray(p.utils.GenerateFilledArray(1.0 / float64(gradientSize[0]*gradientSize[1])))

	// Declare array of channels to recieve value from each Goroutine
	rowChannels := make([]chan [][]*ckks.Ciphertext, len(gradient))

	// loop through each row and start Goroutine
	for row := range gradient {

		rowChannels[row] = make(chan [][]*ckks.Ciphertext)

		go func(rowIndex int, rowChannel chan [][]*ckks.Ciphertext) {

			colChannels := make([]chan []*ckks.Ciphertext, len(gradient[rowIndex]))

			for col := range gradient[rowIndex] {

				colChannels[col] = make(chan []*ckks.Ciphertext)

				go func(rowIndex int, colIndex int, colChannel chan []*ckks.Ciphertext) {

					depChannels := make([]chan *ckks.Ciphertext, len(gradient[rowIndex][colIndex]))
					for depth := range gradient[rowIndex][colIndex] {
						depChannels[depth] = make(chan *ckks.Ciphertext)
						p.utils.MultiplyPlainConcurrent(gradient[rowIndex][colIndex][depth], &divider, true, depChannels[depth])
					}

					colOutput := make([]*ckks.Ciphertext, len(gradient[rowIndex][colIndex]))

					for depth := range depChannels {
						colOutput[depth] = <-depChannels[depth]
					}

					colChannel <- colOutput

				}(rowIndex, col, colChannels[col])

			}

			// Capture value from next and send back to previous
			rowOutput := make([][]*ckks.Ciphertext, len(gradient[rowIndex]))
			for col := range colChannels {
				rowOutput[col] = <-colChannels[col]
			}
			rowChannel <- rowOutput

		}(row, rowChannels[row])

	}

	for row := range rowChannels {
		gradient[row] = <-rowChannels[row]
	}

	// ===================================================
	// ==============  calculate  backward  ==============
	// ===================================================

	currentGradRow := 0
	upSampledGradient := make([][][]*ckks.Ciphertext, p.InputSize[0])

	// Loop through each input datapoint that corresponds to the first row of the pooling filter
	for row := 0; row <= p.InputSize[0]-p.Size[0]; row += p.Strides[0] {

		currentGradCol := 0

		// Loop through each input datapoint that corresponds to the first column of the pooling filter
		for column := 0; column <= p.InputSize[1]-p.Size[1]; column += p.Strides[1] {

			for poolRow := 0; poolRow < p.Size[0]; poolRow++ {

				// Make row slice if undeclared
				if upSampledGradient[row+poolRow] == nil {
					upSampledGradient[row+poolRow] = make([][]*ckks.Ciphertext, p.InputSize[1])
				}

				for poolCol := 0; poolCol < p.Size[1]; poolCol++ {

					// Make column slice if undeclared
					if upSampledGradient[row+poolRow][column+poolCol] == nil {
						upSampledGradient[row+poolRow][column+poolCol] = make([]*ckks.Ciphertext, p.InputSize[2])
					}

					// Loop through each depth of input
					for depth := 0; depth < p.InputSize[2]; depth++ {

						if upSampledGradient[row+poolRow][column+poolCol][depth] == nil {
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

	// Bootstrap output
	if p.btspOutput[1] {
		p.utils.Bootstrap3dInPlace(upSampledGradient)
	}

	return Gradient2d{InputGradient: upSampledGradient}

}

func (p *AveragePooling2D) UpdateGradient(gradient Gradient2d, lr float64) {}

func (p AveragePooling2D) IsTrainable() bool {
	return false
}

func (p AveragePooling2D) HasActivation() bool {
	return false
}

func (p AveragePooling2D) GetForwardLevelConsumption() int {
	return 1
}

func (p AveragePooling2D) GetBackwardLevelConsumption() int {
	return 1
}

func (p AveragePooling2D) GetForwardActivationLevelConsumption() int {
	return 0
}

func (p AveragePooling2D) GetBackwardActivationLevelConsumption() int {
	return 0
}

func (p *AveragePooling2D) SetBootstrapOutput(set bool, direction string) {
	switch direction {
	case "forward":
		p.btspOutput[0] = set
	case "backward":
		p.btspOutput[1] = set
	}
}

func (p *AveragePooling2D) SetBootstrapActivation(set bool, direction string) {

}

func (p *AveragePooling2D) SetWeightLevel(lvl int){

}