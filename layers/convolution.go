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
	Row		int
	Column	int
	Depth	int
	Data 	[][][]*ckks.Ciphertext
}

func generateRandomNormal2dKernel(row int, col int, depth int, utils utility.Utils) conv2dKernel {

	randomNums := array.GenerateRandomNormalArray(row * col * depth)
	data := make([][][]*ckks.Ciphertext, row)

	for r := 0; r < row; r++ {

		data[r] = make([][]*ckks.Ciphertext, col)

		for c := 0; c < col; c++ {

			data[r][c] = make([]*ckks.Ciphertext, depth)

			for d := 0; d < depth; d++{
				encryptedRandom := utils.Encrypt(utils.GenerateFilledArray(randomNums[(r * col) + c]))
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

func (k *conv2dKernel) sgd(gradient *ckks.Ciphertext, row int, col int, lr float64, utils utility.Utils){

	lrPlain := utils.EncodePlaintextFromArray(utils.GenerateFilledArray(lr))
	lrGradient := utils.MultiplyPlainNew(gradient, &lrPlain, true, false)

	for d := range k.Data[row][col] {
		utils.Sub(*k.Data[row][col][d], lrGradient, k.Data[row][col][d])
	}

}

func (k *conv2dKernel) dialate(dialation []int, rightPadding bool, bottomPadding bool) {

	// Turn bool to int
	rPadding := 0
	if rightPadding {
		rPadding = 1
	}

	// Turn bool to int
	bPadding := 0
	if bottomPadding {
		bPadding = 1
	}

	// Calculate the number of row for dialated kernel
	newRow := k.Row + (dialation[0] * (k.Row - 1)) + rPadding

	// Calculate row modulo for checking if certain row should be empty or not
	rowMod := dialation[0] + 1

	// Calculate the number of column for dialated kernel
	newCol := k.Column + (dialation[1] * (k.Column - 1)) + bPadding

	// Calculate column modulo for checking if certain column should be empty or not
	colMod := dialation[1] + 1

	newData := make([][][]*ckks.Ciphertext, newRow)

	oldRow := 0

	for row := range newData {

		oldCol := 0
		newData[row] = make([][]*ckks.Ciphertext, newCol)

		if row % rowMod == 0 {
			for col := range newData[row]{

				newData[row][col] = make([]*ckks.Ciphertext, k.Depth)

				if col % colMod == 0 {

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

//=================================================
//		   		CONVOLUTIONAL LAYER
//=================================================

type Conv2D struct {
	utils		utility.Utils
	Kernels		[]conv2dKernel
	Bias		[]ckks.Ciphertext
	Strides		[]int
	Padding		bool
	Activation	*activations.Activation
	InputSize	[]int
	batchSize	int
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
func NewConv2D(utils utility.Utils, filters int, kernelSize []int, strides []int, padding bool, activation *activations.Activation, useBias bool, inputSize []int, batchSize int) Conv2D{

	kernels := make([]conv2dKernel, filters)

	for i := range kernels{
		kernels[i] = generateRandomNormal2dKernel(kernelSize[0], kernelSize[1], inputSize[2], utils)
	}

	bias := []ckks.Ciphertext{}
	if useBias{
		bias = make([]ckks.Ciphertext, filters)
		for i := range bias {
			bias[i] = utils.Encrypt(utils.GenerateRandomNormalArray(batchSize))
		}
	}

	return Conv2D{utils: utils, Kernels: kernels, Bias: bias, Strides: strides, Padding: padding, Activation: activation, InputSize: inputSize, batchSize: batchSize}

}

// Evaluate forward pass of the convolutional 2d layer
// input must be packed according to section 3.1.1 in https://eprint.iacr.org/2018/1056.pdf
func (c Conv2D) Forward (input [][][]*ckks.Ciphertext) [][][]*ckks.Ciphertext {

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

	// Loop through each row for kernel start position
	for row := start; row < (c.InputSize[0] - kernelDim[0]); row += c.Strides[0]{

		output[row] = make([][]*ckks.Ciphertext, outputSize[1])

		// Loop through each column for kernel start position
		for col := start; col < (c.InputSize[1] - kernelDim[1]); col += c.Strides[1]{

			output[row][col] = make([]*ckks.Ciphertext, len(c.Kernels))
			
			// Loop through each kernel
			for k := range c.Kernels{

				// Declare result to store the dot product of kernel and that region of input
				var result *ckks.Ciphertext

				for krow := 0; krow < c.Kernels[k].Row; krow++ {
					for kcol := 0; kcol < c.Kernels[k].Column; kcol++{

						// Check if in padding
						if row + krow == -1 || col + kcol == -1 {
							continue
						}

						for kdep := 0; kdep < c.Kernels[k].Depth; kdep++{

							product := c.utils.MultiplyNew(*c.Kernels[k].Data[krow][kcol][kdep], *input[row + krow][col + kcol][kdep], true, false)

							if result == nil {
								result = &product
							} else {
								c.utils.Add(product, *result, result)
							}

						}
					}
				}

				if len(c.Bias) != 0{
					c.utils.Add(c.Bias[k], *result, result)
				}

				if c.Activation != nil{
					activatedResult := (*c.Activation).Forward(*result, c.InputSize[1])
					result = &activatedResult
				}

				output[row][col][k] = result
				
			}

		}
	}

	return output

}

func (c Conv2D) Backward(input [][][]*ckks.Ciphertext, output [][][]*ckks.Ciphertext, gradient [][][]*ckks.Ciphertext, lr float64){

	// Calculate ∂L/∂Z

	// Calculate ∂Z/∂F
	gradientKernel := generate2dKernelFromArray(gradient)

	padding := 0
	if c.Padding{
		padding = 1
	}

	if c.Strides[0] == 0 && c.Strides[1] == 0 {
		// Calculate whether right padding is necessary for the gradient kernel
		dialatedRow := gradientKernel.Row + (c.Strides[0] * (gradientKernel.Row - 1))
		rightPadding := (c.InputSize[0] + (2*padding) + 1 - c.Kernels[0].Row) + 1 == dialatedRow

		// Calculate whether bottom padding is necessary for the gradient kernel
		dialatedColumn := gradientKernel.Column + (c.Strides[1] * (gradientKernel.Column - 1))
		bottomPadding := (c.InputSize[1] + (2*padding) + 1 - c.Kernels[0].Column) + 1 == dialatedColumn

		gradientKernel.dialate(c.Strides, rightPadding, bottomPadding)
	}

	// Loop through input row
	for row := (padding * -1); row < (c.InputSize[0] - gradientKernel.Row); row++{

		// Loop through input column
		for col := (padding * -1); col < (c.InputSize[1] - gradientKernel.Column); col++{
			
			// loop throught gradient of each kernel in this layer
			for k := 0; k < gradientKernel.Depth; k++{

				var result *ckks.Ciphertext

				// loop through gradient kernel's row and column
				for krow := 0; krow < gradientKernel.Row; krow++ {
					for kcol := 0; kcol < gradientKernel.Column; kcol++{

						// Check if in padding or is nil
						if row + krow == -1 || col + kcol == -1 || gradientKernel.Data[krow][kcol][0] == nil {
							continue
						}

						// Loop through input channel
						for dep := 0; dep < c.InputSize[2]; dep++{

							product := c.utils.MultiplyNew(*gradientKernel.Data[krow][kcol][k], *input[row + krow][col + kcol][dep], true, false)

							if result == nil {
								result = &product
							} else {
								c.utils.Add(product, *result, result)
							}

						}
					}
				}

				if result != nil{
					// Update gradient using SGD
					c.Kernels[k].sgd(result, row + (padding * -1), col + (padding * -1), lr, c.utils)
				}
				
			}

		}
	}
}

func (c *Conv2D) GetOutputSize() []int {

	padding := 0
	if c.Padding {
		padding = 1
	}

	// (W1−F+2P)/S+1
	outputRowSize := int(float64(c.InputSize[0] - c.Kernels[0].Row + (2 * padding)) / float64(c.Strides[0])) + 1

	// (H1−F+2P)/S+1 
	outputColumnSize := int(float64(c.InputSize[1] - c.Kernels[1].Row + (2 * padding)) / float64(c.Strides[1])) + 1

	return []int{outputRowSize, outputColumnSize, len(c.Kernels)}

}