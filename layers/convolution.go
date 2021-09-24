package layers

import (
	"math"
	"sync"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/activations"
	"github.com/perm-ai/go-cerebrum/array"
	"github.com/perm-ai/go-cerebrum/utility"
)

//=================================================
//		   CONVOLUTIONAL KERNEL (FILTER)
//=================================================

type conv2dKernel struct {
	Row    int
	Column int
	Depth  int
	Data   [][][]*ckks.Ciphertext
}

func generateRandomNormal2dKernel(row int, col int, depth int, utils utility.Utils) conv2dKernel {

	weightStdDev := math.Sqrt(2.0 / float64(row*col*depth))
	randomNums := array.GenerateRandomNormalArray(row*col*depth, weightStdDev)
	data := make([][][]*ckks.Ciphertext, row)

	for r := 0; r < row; r++ {

		data[r] = make([][]*ckks.Ciphertext, col)

		for c := 0; c < col; c++ {

			data[r][c] = make([]*ckks.Ciphertext, depth)

			for d := 0; d < depth; d++ {
				data[r][c][d] = utils.EncryptToLevel(utils.GenerateFilledArray(randomNums[(r*col)+c]), 9)
			}
		}
	}

	return conv2dKernel{row, col, depth, data}

}

func generate2dKernelFromArray(data [][][]*ckks.Ciphertext) conv2dKernel {

	row := len(data)
	col := len(data[0])
	depth := len(data[0][0])

	return conv2dKernel{Row: row, Column: col, Depth: depth, Data: data}

}

func (k *conv2dKernel) updateWeight(gradient [][]*ckks.Ciphertext, lr ckks.Plaintext, utils utility.Utils, weightLevel int) {

	// create row weight group
	var rowWg sync.WaitGroup

	for row := range gradient {

		// Add 1 task to wait group
		rowWg.Add(1)
		
		go func(rowIndex int){

			defer rowWg.Done()

			var colWg sync.WaitGroup

			for col := range gradient[rowIndex] {

				colWg.Add(1)

				go func(colIndex int, colUtils utility.Utils){
					defer colWg.Done()
					
					averagedLrGradient := colUtils.MultiplyPlainNew(gradient[rowIndex][colIndex], &lr, true, false)

					if averagedLrGradient.Level() < weightLevel{
						colUtils.BootstrapInPlace(averagedLrGradient)
					}

					var depWg sync.WaitGroup

					for d := range k.Data[rowIndex][colIndex] {
						depWg.Add(1)

						go func(depIndex int, depUtils utility.Utils){
							defer depWg.Done()
							utils.Sub(k.Data[rowIndex][colIndex][depIndex], averagedLrGradient, k.Data[rowIndex][colIndex][depIndex])
						}(d, colUtils.ShallowCopy())
						
					}

					depWg.Done()

				}(col, utils.ShallowCopy())
				
			}

			colWg.Wait()

		}(row)
		
	}

	rowWg.Wait()

}

func (k *conv2dKernel) dilate(dilation []int) {

	// Calculate the number of row for dilated kernel
	newRow := k.Row + ((dilation[0] - 1) * (k.Row - 1))

	// Calculate row modulo for checking if certain row should be empty or not
	rowMod := dilation[0]

	// Calculate the number of column for dilated kernel
	newCol := k.Column + ((dilation[1] - 1) * (k.Column - 1))

	// Calculate column modulo for checking if certain column should be empty or not
	colMod := dilation[1]

	newData := make([][][]*ckks.Ciphertext, newRow)

	oldRow := 0

	for row := range newData {

		oldCol := 0
		newData[row] = make([][]*ckks.Ciphertext, newCol)

		if row%rowMod == 0 {
			for col := range newData[row] {

				newData[row][col] = make([]*ckks.Ciphertext, k.Depth)

				if col%colMod == 0 {

					newData[row][col] = k.Data[oldRow][oldCol]
					oldCol++

				}

			}
			oldRow++
		}

	}

	k.Row = newRow
	k.Column = newCol
	k.Data = newData

}

// Add padding to kernel (used as part of calculation for dL/dA(l-1))
// size is padding you want to add for [row, column]
func (k *conv2dKernel) addPadding(size []int) {

	newRow := k.Row + (2 * size[0])
	newCol := k.Column + (2 * size[1])

	newData := make([][][]*ckks.Ciphertext, newRow)

	oldRowIndex := 0

	// Loop through each new row in new data
	for newRowIndex := 0; newRowIndex < newRow; newRowIndex++ {

		newData[newRowIndex] = make([][]*ckks.Ciphertext, newCol)
		oldColIndex := 0

		// Check if that row isn't in padding
		if newRowIndex >= size[0] && newRowIndex < size[0]+k.Row {

			// Loop through each column in that row
			for newColIndex := 0; newColIndex < newCol; newColIndex++ {

				// Check if column in padding or not
				if newColIndex >= size[1] && newColIndex < size[1]+k.Column {
					newData[newRowIndex][newColIndex] = k.Data[oldRowIndex][oldColIndex]
					oldColIndex++
				} else {
					newData[newRowIndex][newColIndex] = make([]*ckks.Ciphertext, k.Depth)
				}

			}
			oldRowIndex++

		}

	}

	k.Row = newRow
	k.Column = newCol
	k.Data = newData

}

func (k conv2dKernel) rotate180() conv2dKernel {

	rotated := make([][][]*ckks.Ciphertext, k.Row)

	for row := range rotated {

		rotated[row] = make([][]*ckks.Ciphertext, k.Column)

		for col := range rotated[row] {

			rotated[row][col] = k.Data[k.Row-row-1][k.Column-col-1]

		}

	}

	return conv2dKernel{k.Row, k.Column, k.Depth, rotated}

}

//=================================================
//		   		CONVOLUTIONAL LAYER
//=================================================

type Conv2D struct {
	utils          utility.Utils
	Kernels        []conv2dKernel
	Bias           []*ckks.Ciphertext
	Strides        []int
	Padding        bool
	Activation     *activations.Activation
	InputSize      []int
	btspOutput     []bool
	btspActivation []bool
	batchSize      int
	weightLevel	   int
}

// Constructor for Convolutional layer struct
// filters is number of kernel in this layer
// kernelSize is the size of kernel with row in index 0 and column in index 1
// strides specify the stride in convolution along rows and columns
// padding specify the padding when evaluating convolution. Set to true to add padding of size 1 to each side and false for no padding
// activation specify the activation function
// useBias specify whether or not a bias will be used
// inputSize specify the size of input in the following order [row, column, channel] if there's channel
// batchSize, when useBias is true, is the number of training example in a training ciphertexts with lowest number of training data
func NewConv2D(utils utility.Utils, filters int, kernelSize []int, strides []int, padding bool, activation *activations.Activation, useBias bool, inputSize []int, batchSize int) Conv2D {

	kernels := make([]conv2dKernel, filters)

	for i := range kernels {
		kernels[i] = generateRandomNormal2dKernel(kernelSize[0], kernelSize[1], inputSize[2], utils)
	}

	bias := []*ckks.Ciphertext{}
	if useBias {
		randomBias := utils.GenerateFilledArraySize(0, filters)
		bias = make([]*ckks.Ciphertext, filters)
		for i := range bias {
			bias[i] = utils.EncryptToLevel(utils.GenerateFilledArraySize(randomBias[i], batchSize), 9)
		}
	}

	return Conv2D{utils: utils, Kernels: kernels, Bias: bias, Strides: strides, Padding: padding, Activation: activation, InputSize: inputSize, batchSize: batchSize, btspOutput: []bool{false, false}, btspActivation: []bool{false, false}, weightLevel: 9}

}

func (c *Conv2D) LoadKernels(kernels []conv2dKernel) {
	c.Kernels = kernels
}

// Evaluate forward pass of the convolutional 2d layer
// input must be packed according to section 3.1.1 in https://eprint.iacr.org/2018/1056.pdf
func (c Conv2D) Forward(input [][][]*ckks.Ciphertext) Output2d {

	// Calculate the starting coordinate
	start := 0
	if c.Padding {
		start = -1
	}

	outputSize := c.GetOutputSize()

	// Get kernel dimention
	kernelDim := [2]int{c.Kernels[0].Row, c.Kernels[0].Column}

	// Generate array to store output
	output3d := make([][][]*ckks.Ciphertext, outputSize[0])
	activatedOutput3d := make([][][]*ckks.Ciphertext, outputSize[0])

	// Generate array of 2d channels for parallel computation
	output3dChannels := make([]chan [][]*ckks.Ciphertext, outputSize[0])
	activation3dChannels := make([]chan [][]*ckks.Ciphertext, outputSize[0])

	// Store the current row of output
	outputRow := 0

	// Loop through each row for kernel start position
	for row := start; row+kernelDim[0]-1 <= c.InputSize[0]+(-1*start)-1; row += c.Strides[0] {

		output3dChannels[outputRow] = make(chan [][]*ckks.Ciphertext)
		activation3dChannels[outputRow] = make(chan [][]*ckks.Ciphertext)

		go func(rowIndex int, output2dChannel chan [][]*ckks.Ciphertext, activation2dChannel chan [][]*ckks.Ciphertext) {

			// Store the current column of output
			outputCol := 0

			// Generate array to store output
			output2d := make([][]*ckks.Ciphertext, outputSize[1])
			activatedOutput2d := make([][]*ckks.Ciphertext, outputSize[1])

			// Generate array of channels for concurrent computation
			output2dChannels := make([]chan []*ckks.Ciphertext, outputSize[1])
			activation2dChannels := make([]chan []*ckks.Ciphertext, outputSize[1])

			// Loop through each column for kernel start position
			for col := start; col+kernelDim[1]-1 <= c.InputSize[1]+(-1*start)-1; col += c.Strides[1] {

				output2dChannels[outputCol] = make(chan []*ckks.Ciphertext)
				activation2dChannels[outputCol] = make(chan []*ckks.Ciphertext)

				go func(rowIndex int, colIndex int, output2dChannel chan []*ckks.Ciphertext, activation2dChannel chan []*ckks.Ciphertext) {

					// Generate array to store output
					output1d := make([]*ckks.Ciphertext, outputSize[2])
					activatedOutput1d := make([]*ckks.Ciphertext, outputSize[2])

					// Generate array of channels for concurrent computation
					output1dChannels := make([]chan *ckks.Ciphertext, outputSize[2])

					// Loop through each kernel
					for k := range c.Kernels {

						output1dChannels[k] = make(chan *ckks.Ciphertext)

						go func(rowIndex int, colIndex int, kernelIndex int, output1dChannel chan *ckks.Ciphertext) {

							// Declare result to store the dot product of kernel and that region of input
							// var result *ckks.Ciphertext
							kernelCiphertext := []*ckks.Ciphertext{}
							inputCiphertext := []*ckks.Ciphertext{}

							for krow := 0; krow < c.Kernels[kernelIndex].Row; krow++ {
								for kcol := 0; kcol < c.Kernels[kernelIndex].Column; kcol++ {

									// Check if in padding
									if rowIndex+krow == -1 || colIndex+kcol == -1 || rowIndex+krow == c.InputSize[0] || colIndex+kcol == c.InputSize[1] {
										continue
									}

									for kdep := 0; kdep < c.Kernels[kernelIndex].Depth; kdep++ {

										kernelCiphertext = append(kernelCiphertext, c.Kernels[kernelIndex].Data[krow][kcol][kdep])
										inputCiphertext = append(inputCiphertext, input[rowIndex+krow][colIndex+kcol][kdep])

									}
								}
							}

							result := c.utils.InterDotProduct(kernelCiphertext, inputCiphertext, true, false, true)

							if len(c.Bias) != 0 {
								c.utils.Add(c.Bias[kernelIndex], result, result)
							}

							output1dChannel <- result

						}(rowIndex, colIndex, k, output1dChannels[k])

					}

					for k := range output1dChannels {
						output1d[k] = <-output1dChannels[k]
					}

					// Bootstrap output ciphertext
					if c.btspOutput[0] {
						c.utils.Bootstrap1dInPlace(output1d, true)
					}

					if c.Activation != nil {

						activatedOutput1d = (*c.Activation).Forward(output1d, c.batchSize)

						if c.btspActivation[0] {
							c.utils.Bootstrap1dInPlace(activatedOutput1d, true)
						}

					}

					output2dChannel <- output1d
					activation2dChannel <- activatedOutput1d

				}(rowIndex, col, output2dChannels[outputCol], activation2dChannels[outputCol])

				outputCol++
			}

			for col := 0; col < outputSize[1]; col++ {
				output2d[col] = <-output2dChannels[col]
				activatedOutput2d[col] = <-activation2dChannels[col]
			}

			output2dChannel <- output2d
			activation2dChannel <- activatedOutput2d

		}(row, output3dChannels[outputRow], activation3dChannels[outputRow])

		outputRow++

	}

	for row := 0; row < outputSize[0]; row++ {
		output3d[row] = <-output3dChannels[row]
		activatedOutput3d[row] = <-activation3dChannels[row]
	}

	return Output2d{output3d, activatedOutput3d}

}

func (c Conv2D) Backward(input [][][]*ckks.Ciphertext, output [][][]*ckks.Ciphertext, gradient [][][]*ckks.Ciphertext, hasPrevLayer bool) Gradient2d {

	gradients := Gradient2d{}

	// Calculate ∂L/∂Z
	if c.Activation != nil {
		if (*c.Activation).GetType() != "softmax" {

			rowChannels := make([]chan [][]*ckks.Ciphertext, len(gradient))

			for ri := range gradient {

				rowChannels[ri] = make(chan [][]*ckks.Ciphertext)

				go func(rowIndex int, rowChannel chan [][]*ckks.Ciphertext) {

					columnChannels := make([]chan []*ckks.Ciphertext, len(gradient[rowIndex]))

					for ci := range gradient[rowIndex] {

						columnChannels[ci] = make(chan []*ckks.Ciphertext)

						go func(rowIndex int, colIndex int, columnChannel chan []*ckks.Ciphertext) {
							// Calculate ∂A/∂Z
							activationGradient := (*c.Activation).Backward(output[rowIndex][colIndex], c.batchSize)

							// store data if ct has been bootstrapped, so there're not double bootstrapping
							columnBootstrapped := false

							// Bootstrapp concurrently
							if activationGradient[0].Level() == 1 && c.btspActivation[1] {
								c.utils.Bootstrap1dInPlace(activationGradient, true)
								columnBootstrapped = true
							}

							productChannels := make([]chan *ckks.Ciphertext, len(activationGradient))

							// Create go routine to multiply concurrently
							for di := range activationGradient {

								productChannels[di] = make(chan *ckks.Ciphertext)
								go func(a *ckks.Ciphertext, b *ckks.Ciphertext, utils utility.Utils, c chan *ckks.Ciphertext) {

									utils.MultiplyConcurrent(a, b, true, c)

								}(gradient[rowIndex][colIndex][di], activationGradient[di], c.utils, productChannels[di])

							}

							gradientWrtOutput := make([]*ckks.Ciphertext, len(activationGradient))

							// Wait for go routine to complete and save values sent through channels
							for di := range productChannels {
								gradientWrtOutput[di] = <-productChannels[di]
							}

							// Bootstrapp if required and haven't been done yet
							if c.btspActivation[1] && !columnBootstrapped {
								c.utils.Bootstrap1dInPlace(gradientWrtOutput, true)
							}

							columnChannel <- gradientWrtOutput

						}(rowIndex, ci, columnChannels[ci])

					}

					columnOutput := make([][]*ckks.Ciphertext, len(gradient[rowIndex]))

					for ci := range columnChannels {
						columnOutput[ci] = <-columnChannels[ci]
					}

					rowChannel <- columnOutput

				}(ri, rowChannels[ri])

			}

			for ri := range rowChannels {
				gradient[ri] = <-rowChannels[ri]
			}

		}
	}

	if len(c.Bias) != 0 {
		// Update bias using Σr(Σc(∂L/∂Z))
		gradients.BiasGradient = make([]*ckks.Ciphertext, len(c.Kernels))
		for k := range c.Kernels {
			for ri := range gradient {
				for ci := range gradient[ri] {
					if gradients.BiasGradient[k] == nil {
						gradients.BiasGradient[k] = gradient[ri][ci][k]
					} else {
						c.utils.Add(gradient[ri][ci][k], gradients.BiasGradient[k], gradients.BiasGradient[k])
					}
				}
			}
		}
	}

	// Calculate ∂Z/∂F
	gradientKernel := generate2dKernelFromArray(gradient)

	padding := 0
	if c.Padding {
		padding = 1
	}

	if c.Strides[0] != 0 && c.Strides[1] != 0 {
		gradientKernel.dilate(c.Strides)
	}

	gradients.WeightGradient = make([][][]*ckks.Ciphertext, len(c.Kernels))
	weightGradientKernelChannels := make([]chan [][]*ckks.Ciphertext, len(c.Kernels))

	// loop throught gradient of each kernel
	for k := 0; k < gradientKernel.Depth; k++ {

		weightGradientKernelChannels[k] = make(chan [][]*ckks.Ciphertext)

		go func(kernelIndex int, weightGradientKernelChannel chan [][]*ckks.Ciphertext) {

			weightGradientRowChannels := make([]chan []*ckks.Ciphertext, c.Kernels[0].Row)
			currentGradientRow := 0

			// Loop through input row
			for row := (padding * -1); row <= c.InputSize[0]-gradientKernel.Row+padding; row++ {

				weightGradientRowChannels[currentGradientRow] = make(chan []*ckks.Ciphertext)

				go func(kernelIndex int, rowIndex int, weightGradientRowChannel chan []*ckks.Ciphertext) {

					weightGradientColChannels := make([]chan *ckks.Ciphertext, c.Kernels[0].Column)
					currentGradientCol := 0

					// Loop through input column
					for col := (padding * -1); col <= c.InputSize[1]-gradientKernel.Column+padding; col++ {

						weightGradientColChannels[currentGradientCol] = make(chan *ckks.Ciphertext)

						go func(kernelIndex int, rowIndex int, colIndex int, weightGradientColChannel chan *ckks.Ciphertext) {

							kernelCiphertexts := []*ckks.Ciphertext{}
							inputCiphertexts := []*ckks.Ciphertext{}

							// loop through gradient kernel's row and column
							for krow := 0; krow < gradientKernel.Row; krow++ {
								for kcol := 0; kcol < gradientKernel.Column; kcol++ {
									// Check if in padding or is nil
									if rowIndex+krow == -1 || colIndex+kcol == -1 || gradientKernel.Data[krow][kcol][0] == nil {
										continue
									}
									// Loop through input channel
									for dep := 0; dep < c.InputSize[2]; dep++ {
										kernelCiphertexts = append(kernelCiphertexts, gradientKernel.Data[krow][kcol][kernelIndex])
										inputCiphertexts = append(inputCiphertexts, input[rowIndex+krow][colIndex+kcol][dep])
									}
								}
							}

							result := c.utils.InterDotProduct(kernelCiphertexts, inputCiphertexts, true, false, true)

							weightGradientColChannel <- result

						}(kernelIndex, rowIndex, col, weightGradientColChannels[currentGradientCol])

						currentGradientCol++
					}

					// Generate array to store weight gradient of each column in a row
					rowWeightGradient := make([]*ckks.Ciphertext, c.Kernels[0].Column)

					// Capture weight gradient of each column and save to array
					for col := range weightGradientColChannels {
						rowWeightGradient[col] = <-weightGradientColChannels[col]
					}

					// Sent row gradient back to row channel reciever
					weightGradientRowChannel <- rowWeightGradient

				}(kernelIndex, row, weightGradientRowChannels[currentGradientRow])

				currentGradientRow++
			}

			// Generate array to store weight gradient of each row in a kernel
			kernelWeightGradient := make([][]*ckks.Ciphertext, c.Kernels[0].Row)

			// Capture weight gradient of each row from channels
			for row := range weightGradientRowChannels {
				kernelWeightGradient = <-weightGradientKernelChannels[row]
			}

			// Sent kernel weight gradient back
			weightGradientKernelChannel <- kernelWeightGradient

		}(k, weightGradientKernelChannels[k])

	}

	// Capture weight gradient from channels
	for k := range weightGradientKernelChannels {
		gradients.WeightGradient[k] = <-weightGradientKernelChannels[k]
	}

	// Calculate ∂L/∂A(l-1)

	// Rotate all kernel by 180 degree clockwise
	rotatedKernels := make([]conv2dKernel, len(c.Kernels))

	for k := range c.Kernels {
		rotatedKernels[k] = c.Kernels[k].rotate180()
	}

	// Prepare loss gradient w.r.t output of this layer (∂L/∂Z(l-1)) for ∂L/∂A(l-1) convolution calculation
	lossGrad := generate2dKernelFromArray(gradient)
	lossGrad.dilate(c.Strides)
	lossGrad.addPadding([]int{c.Kernels[0].Row, c.Kernels[0].Column})

	if hasPrevLayer {

		gradients.InputGradient = make([][][]*ckks.Ciphertext, c.InputSize[0])
		inputGradientChannels := make([]chan [][]*ckks.Ciphertext, c.InputSize[0])

		// Perform convolution between loss wrt Z and filter
		for row := 0; row < lossGrad.Row-c.Kernels[0].Row; row++ {

			inputGradientChannels[row] = make(chan [][]*ckks.Ciphertext)

			go func(rowIndex int, inputGradientChannel chan [][]*ckks.Ciphertext) {

				inputGradientColumnChannels := make([]chan []*ckks.Ciphertext, c.InputSize[1])

				for col := 0; col < lossGrad.Column-c.Kernels[0].Column; col++ {

					inputGradientColumnChannels[col] = make(chan []*ckks.Ciphertext)

					go func(rowIndex int, colIndex int, inputGradientColumnChannel chan []*ckks.Ciphertext) {

						inputGradientDepthChannels := make([]chan *ckks.Ciphertext, c.InputSize[2])

						for d := 0; d < c.InputSize[2]; d++ {

							inputGradientDepthChannels[d] = make(chan *ckks.Ciphertext)

							go func(rowIndex int, colIndex int, depIndex int, inputGradientDepthChannel chan *ckks.Ciphertext) {

								rotatedKernelsCiphertexts := []*ckks.Ciphertext{}
								lossGradientCiphertexts := []*ckks.Ciphertext{}

								// Loop through each kernel
								for k := 0; k < len(c.Kernels); k++ {
									// Loop through each row in kernel
									for krow := range c.Kernels[k].Data {
										// Loop through each column in kernel
										for kcol := range c.Kernels[k].Data[krow] {

											// Check if in padding
											if rotatedKernels[k].Data[krow][kcol][depIndex] != nil && lossGrad.Data[rowIndex+krow][colIndex+kcol][k] != nil {

												rotatedKernelsCiphertexts = append(rotatedKernelsCiphertexts, rotatedKernels[k].Data[krow][kcol][depIndex])
												lossGradientCiphertexts = append(lossGradientCiphertexts, lossGrad.Data[rowIndex+krow][colIndex+kcol][k])

											}

										}
									}
								}

								// Calculate dot product and send result back through channel
								inputGradientDepthChannel <- c.utils.InterDotProduct(rotatedKernelsCiphertexts, lossGradientCiphertexts, true, false, true)

							}(rowIndex, colIndex, d, inputGradientDepthChannels[d])

						}

						// Create array to store input gradient for a column
						inputColumnGradient := make([]*ckks.Ciphertext, c.InputSize[2])

						// Capture and store values through channels
						for d := range inputGradientDepthChannels {
							inputColumnGradient[d] = <-inputGradientDepthChannels[d]
						}

						// Bootstrap if necessary
						if c.btspOutput[1] {
							c.utils.Bootstrap1dInPlace(inputColumnGradient, true)
						}

						// Send input gradient of this column back through channel
						inputGradientColumnChannel <- inputColumnGradient

					}(rowIndex, col, inputGradientColumnChannels[col])

				}

				// Create array to store input gradient for this row
				inputGradientRow := make([][]*ckks.Ciphertext, c.InputSize[1])

				// Capture and store input gradient of each column in this row and store it in this row's array
				for col := range inputGradientColumnChannels {
					inputGradientRow[col] = <-inputGradientColumnChannels[col]
				}

				// Send input gradient of this row back through channel
				inputGradientChannel <- inputGradientRow

			}(row, inputGradientChannels[row])

		}

		// Capture row from go routine channels and save to gradient struct
		for row := range inputGradientChannels {
			gradients.InputGradient[row] = <-inputGradientChannels[row]
		}

	}

	return gradients
}

func (c *Conv2D) UpdateGradient(gradient Gradient2d, lr float64) {

	batchAverager := c.utils.EncodePlaintextFromArray(c.utils.GenerateFilledArraySize(lr/float64(c.batchSize), c.batchSize))

	// create weight group
	var wg sync.WaitGroup

	for k := range c.Kernels {

		wg.Add(1)

		go func(index int, utils utility.Utils) {

			defer wg.Done()
			var biasWg sync.WaitGroup

			if len(c.Bias) != 0 {

				biasWg.Add(1)
				
				go func(biasUtils utility.Utils){
					defer biasWg.Done()

					biasUtils.SumElementsInPlace(gradient.BiasGradient[index])
					biasUtils.MultiplyPlain(gradient.BiasGradient[index], &batchAverager, gradient.BiasGradient[index], true, false)

					if gradient.BiasGradient[index].Level() < c.weightLevel{
						biasUtils.BootstrapInPlace(gradient.BiasGradient[index])
					}

					biasUtils.Sub(c.Bias[index], gradient.BiasGradient[index], c.Bias[index])

				}(utils.ShallowCopy())

			}

			// Update weight
			c.Kernels[index].updateWeight(gradient.WeightGradient[index], batchAverager, utils, c.weightLevel)

			biasWg.Wait()

		}(k, c.utils.ShallowCopy())

	}

	wg.Wait()
}

func (c *Conv2D) GetOutputSize() []int {

	padding := 0
	if c.Padding {
		padding = 1
	}

	// (W1−F+2P)/S+1
	outputRowSize := int(float64(c.InputSize[0]-c.Kernels[0].Row+(2*padding))/float64(c.Strides[0])) + 1

	// (H1−F+2P)/S+1
	outputColumnSize := int(float64(c.InputSize[1]-c.Kernels[0].Column+(2*padding))/float64(c.Strides[1])) + 1

	return []int{outputRowSize, outputColumnSize, len(c.Kernels)}

}

func (c Conv2D) IsTrainable() bool {
	return true
}

func (c Conv2D) HasActivation() bool {
	return c.Activation != nil
}

func (c *Conv2D) SetBootstrapOutput(set bool, direction string) {
	switch direction {
	case "forward":
		c.btspOutput[0] = set
	case "backward":
		c.btspOutput[1] = set
	}
}

func (c *Conv2D) SetBootstrapActivation(set bool, direction string) {
	switch direction {
	case "forward":
		c.btspActivation[0] = set
	case "backward":
		c.btspActivation[1] = set
	}
}

func (c Conv2D) GetForwardLevelConsumption() int {
	return 1
}

func (c Conv2D) GetBackwardLevelConsumption() int {
	return 1
}

func (c Conv2D) GetForwardActivationLevelConsumption() int {
	if c.HasActivation() {
		return (*c.Activation).GetForwardLevelConsumption()
	} else {
		return 0
	}
}

func (c Conv2D) GetBackwardActivationLevelConsumption() int {
	if c.HasActivation() {
		return (*c.Activation).GetBackwardLevelConsumption()
	} else {
		return 0
	}
}

func (c *Conv2D) SetWeightLevel(lvl int){
	c.weightLevel = lvl
}