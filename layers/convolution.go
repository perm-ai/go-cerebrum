package layers

import (
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

	randomNums := array.GenerateRandomNormalArray(row * col * depth)
	data := make([][][]*ckks.Ciphertext, row)

	for r := 0; r < row; r++ {

		data[r] = make([][]*ckks.Ciphertext, col)

		for c := 0; c < col; c++ {

			data[r][c] = make([]*ckks.Ciphertext, depth)

			for d := 0; d < depth; d++ {
				encryptedRandom := utils.Encrypt(utils.GenerateFilledArray(randomNums[(r*col)+c]))
				data[r][c][d] = &encryptedRandom
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

func (k *conv2dKernel) updateWeight(gradient [][]*ckks.Ciphertext, lr ckks.Plaintext, utils utility.Utils) {

	for row := range gradient {
		for col := range gradient[row] {
			averagedLrGradient := utils.MultiplyPlainNew(gradient[row][col], &lr, true, false)
			for d := range k.Data[row][col] {
				utils.Sub(*k.Data[row][col][d], averagedLrGradient, k.Data[row][col][d])
			}
		}
	}

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
	Bias           []ckks.Ciphertext
	Strides        []int
	Padding        bool
	Activation     *activations.Activation
	InputSize      []int
	btspOutput     []bool
	btspActivation []bool
	batchSize      int
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

	bias := []ckks.Ciphertext{}
	if useBias {
		randomBias := utils.GenerateRandomNormalArray(filters)
		bias = make([]ckks.Ciphertext, filters)
		for i := range bias {
			bias[i] = utils.Encrypt(utils.GenerateFilledArraySize(randomBias[i], batchSize))
		}
	}

	return Conv2D{utils: utils, Kernels: kernels, Bias: bias, Strides: strides, Padding: padding, Activation: activation, InputSize: inputSize, batchSize: batchSize, btspOutput: []bool{false, false}, btspActivation: []bool{false, false}}

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
	output := make([][][]*ckks.Ciphertext, outputSize[0])
	activatedOutput := make([][][]*ckks.Ciphertext, outputSize[0])

	// Store the current row of output
	outputRow := 0

	// Loop through each row for kernel start position
	for row := start; row+kernelDim[0]-1 <= c.InputSize[0]+(-1*start)-1; row += c.Strides[0] {

		output[outputRow] = make([][]*ckks.Ciphertext, outputSize[1])
		activatedOutput[outputRow] = make([][]*ckks.Ciphertext, outputSize[1])

		// Store the current column of output
		outputCol := 0

		// Loop through each column for kernel start position
		for col := start; col+kernelDim[1]-1 <= c.InputSize[1]+(-1*start)-1; col += c.Strides[1] {

			output[outputRow][outputCol] = make([]*ckks.Ciphertext, len(c.Kernels))
			activatedOutput[outputRow][outputCol] = make([]*ckks.Ciphertext, len(c.Kernels))

			// Loop through each kernel
			for k := range c.Kernels {

				// Declare result to store the dot product of kernel and that region of input
				var result *ckks.Ciphertext

				for krow := 0; krow < c.Kernels[k].Row; krow++ {
					for kcol := 0; kcol < c.Kernels[k].Column; kcol++ {

						// Check if in padding
						if row+krow == -1 || col+kcol == -1 || row+krow == c.InputSize[0] || col+kcol == c.InputSize[1] {
							continue
						}

						for kdep := 0; kdep < c.Kernels[k].Depth; kdep++ {

							product := c.utils.MultiplyNew(*c.Kernels[k].Data[krow][kcol][kdep], *input[row+krow][col+kcol][kdep], true, false)

							if result == nil {
								result = &product
							} else {
								c.utils.Add(product, *result, result)
							}

						}
					}
				}

				if len(c.Bias) != 0 {
					c.utils.Add(c.Bias[k], *result, result)
				}

				output[outputRow][outputCol][k] = result

				// Bootstrap output ciphertext
				if c.btspOutput[0]{
					c.utils.BootstrapInPlace(output[outputRow][outputCol][k])
				}
			}

			if c.Activation != nil {

				activatedOutput[outputRow][outputCol] = (*c.Activation).Forward(output[outputRow][outputCol], c.batchSize)

				if c.btspActivation[0]{
					c.utils.Bootstrap1dInPlace(activatedOutput[outputRow][outputCol], false)
				}

			}

			outputCol++
		}
		outputRow++
	}

	return Output2d{output, activatedOutput}

}

func (c Conv2D) Backward(input [][][]*ckks.Ciphertext, output [][][]*ckks.Ciphertext, gradient [][][]*ckks.Ciphertext, hasPrevLayer bool) Gradient2d {

	gradients := Gradient2d{}

	// Calculate ∂L/∂Z
	if c.Activation != nil{
		if (*c.Activation).GetType() != "softmax"{

			for ri := range gradient {
				for ci := range gradient[ri] {
	
					// Calculate ∂A/∂Z
					activationGradient := (*c.Activation).Backward(output[ri][ci], c.batchSize)

					// store data if ct has been bootstrapped, so there're not double bootstrapping
					columnBootstrapped := false

					// Bootstrapp concurrently
					if activationGradient[0].Level() == 1 && c.btspActivation[1] {
						c.utils.Bootstrap1dInPlace(activationGradient, true)
						columnBootstrapped = true
					}

					productChannels := make([]chan ckks.Ciphertext, len(activationGradient))

					// Create go routine to multiply concurrently
					for di := range activationGradient {

						productChannels[di] = make(chan ckks.Ciphertext)
						go func(a *ckks.Ciphertext, b *ckks.Ciphertext, utils utility.Utils, c chan ckks.Ciphertext){

							utils.MultiplyConcurrent(*a, *b, true, c)

						}(gradient[ri][ci][di], activationGradient[di], c.utils, productChannels[di])

					}

					// Wait for go routine to complete and save values sent through channels
					for di := range productChannels {
						prod := <-productChannels[di]
						gradient[ri][ci][di] = &prod
					}

					// Bootstrapp if required and haven't been done yet
					if c.btspActivation[1] && !columnBootstrapped{
						c.utils.Bootstrap1dInPlace(activationGradient, true)
					}
	
				}
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
						c.utils.Add(*gradient[ri][ci][k], *gradients.BiasGradient[k], gradients.BiasGradient[k])
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

	// loop throught gradient of each kernel
	for k := 0; k < gradientKernel.Depth; k++ {

		gradients.WeightGradient[k] = make([][]*ckks.Ciphertext, c.Kernels[0].Row)
		currentGradientRow := 0

		// Loop through input row
		for row := (padding * -1); row <= c.InputSize[0]-gradientKernel.Row+padding; row++ {

			gradients.WeightGradient[k][currentGradientRow] = make([]*ckks.Ciphertext, c.Kernels[0].Column)
			currentGradientCol := 0

			// Loop through input column
			for col := (padding * -1); col <= c.InputSize[1]-gradientKernel.Column+padding; col++ {

				var result *ckks.Ciphertext

				// loop through gradient kernel's row and column
				for krow := 0; krow < gradientKernel.Row; krow++ {
					for kcol := 0; kcol < gradientKernel.Column; kcol++ {

						// Check if in padding or is nil
						if row+krow == -1 || col+kcol == -1 || gradientKernel.Data[krow][kcol][0] == nil {
							continue
						}

						// Loop through input channel
						for dep := 0; dep < c.InputSize[2]; dep++ {

							product := c.utils.MultiplyNew(*gradientKernel.Data[krow][kcol][k], *input[row+krow][col+kcol][dep], true, false)

							if result == nil {
								result = &product
							} else {
								c.utils.Add(product, *result, result)
							}

						}
					}
				}
				gradients.WeightGradient[k][currentGradientRow][currentGradientCol] = result
				currentGradientCol++
			}
			currentGradientRow++
		}
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

		// Perform convolution between loss wrt Z and filter
		for row := 0; row < lossGrad.Row-c.Kernels[0].Row; row++ {

			gradients.InputGradient[row] = make([][]*ckks.Ciphertext, c.InputSize[1])

			for col := 0; col < lossGrad.Column-c.Kernels[0].Column; col++ {

				gradients.InputGradient[row][col] = make([]*ckks.Ciphertext, c.InputSize[2])

				for d := 0; d < c.InputSize[3]; d++ {

					// Loop through each kernel
					for k := 0; k < len(c.Kernels); k++ {
						// Loop through each row in kernel
						for krow := range c.Kernels[k].Data {
							// Loop through each column in kernel
							for kcol := range c.Kernels[k].Data[krow] {

								// Check if in padding
								if rotatedKernels[k].Data[krow][kcol][d] != nil && lossGrad.Data[row+krow][col+kcol][k] != nil {

									product := c.utils.MultiplyNew(*rotatedKernels[k].Data[krow][kcol][d], *lossGrad.Data[row+krow][col+kcol][k], true, false)

									if gradients.InputGradient[row][col][d] == nil {
										gradients.InputGradient[row][col][d] = &product
									} else {
										c.utils.Add(*gradients.InputGradient[row][col][d], product, gradients.InputGradient[row][col][d])
									}

								}

							}
						}
					}

					if c.btspOutput[1]{
						c.utils.BootstrapInPlace(gradients.InputGradient[row][col][d])
					}
					
				}
			}
		}
	}

	return gradients
}

func (c *Conv2D) UpdateGradient(gradient Gradient2d, lr float64) {

	batchAverager := c.utils.EncodePlaintextFromArray(c.utils.GenerateFilledArraySize(lr/float64(c.batchSize), c.batchSize))

	for k := range c.Kernels {

		if len(c.Bias) != 0 {
			// Calculate average gradient of bias in a batch
			c.utils.SumElementsInPlace(gradient.BiasGradient[k])
			c.utils.MultiplyPlain(gradient.BiasGradient[k], &batchAverager, gradient.BiasGradient[k], true, false)

			// Update bias
			c.utils.Sub(c.Bias[k], *gradient.BiasGradient[k], &c.Bias[k])
		}

		// Update weight
		c.Kernels[k].updateWeight(gradient.WeightGradient[k], batchAverager, c.utils)
	}

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

func (c *Conv2D) SetBootstrapOutput(set bool, direction string){
	switch direction{
	case "forward":
		c.btspOutput[0] = set
	case "backward":
		c.btspOutput[1] = set
	}
}

func (c *Conv2D) SetBootstrapActivation(set bool, direction string){
	switch direction{
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
