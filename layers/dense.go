package layers

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/activations"
	"github.com/perm-ai/go-cerebrum/array"
	"github.com/perm-ai/go-cerebrum/utility"
)

//=================================================
//			  DENSE GRADIENT
//=================================================

type DenseGradient struct {
	BiasGradient   		[]*ckks.Ciphertext
	WeightGradient 		[][]*ckks.Ciphertext
	PrevLayerGradient	[]*ckks.Ciphertext
}

//=================================================
//					DENSE LAYER
//=================================================

type Dense struct {

	utils         utility.Utils
	InputUnit     int
	OutputUnit    int
	Weights       [][]*ckks.Ciphertext
	Bias          []*ckks.Ciphertext
	Activation    *activations.Activation

}

func NewDense(utils utility.Utils, inputUnit int, outputUnit int, activation *activations.Activation, useBias bool) Dense {

	// Generate random weights and biases
	weights := make([][]*ckks.Ciphertext, outputUnit)
	bias := make([]*ckks.Ciphertext, outputUnit)

	randomBias := array.GenerateRandomNormalArray(outputUnit)

	for node := 0; node < outputUnit; node++ {

		randomWeight := array.GenerateRandomNormalArray(inputUnit)
		weights[node] = make([]*ckks.Ciphertext, inputUnit)

		if useBias {
			bias[node] = utils.EncryptToPointer(utils.GenerateFilledArray(randomBias[node]))
		}

		for weight := 0; weight < inputUnit; weight++ {

			weights[node][weight] = utils.EncryptToPointer(utils.GenerateFilledArray(randomWeight[weight]))

		}

	}

	return Dense{utils, inputUnit, outputUnit, weights, bias, activation}

}

func (d Dense) Forward(input []*ckks.Ciphertext) []*ckks.Ciphertext {

	output := make([]*ckks.Ciphertext, d.OutputUnit)

	for node := range d.Weights{

		for datapoint := range d.Weights[node]{
			
			product := d.utils.MultiplyNew(*input[datapoint], *d.Weights[node][datapoint], true, false)

			if output[node] == nil{
				output[node] = &product
			} else {
				d.utils.Add(*output[node], product, output[node])
			}

		}

		if d.Bias[node] != nil{
			d.utils.Add(*output[node], *d.Bias[node], output[node])
		}

	}

	return output

}

// input is A(l-1) - activation of previous layer
// output is Z(l) - output of this layer
// gradient is ∂L/∂A(l) - influence that the activation of this layer has on the next layer
// hasPrevLayer determine whether this function calculates the gradient ∂L/∂A(l-1)
func (d *Dense) Backward(input []*ckks.Ciphertext, output []*ckks.Ciphertext, gradient []*ckks.Ciphertext, hasPrevLayer bool) DenseGradient{

	gradients := DenseGradient{}

	// Calculate gradients for last layer
	if d.Activation != nil {
		for b := range d.Bias{
			activationGradient := (*d.Activation).Backward(*output[b], d.OutputUnit)
			d.utils.Multiply(*gradient[b], activationGradient, gradient[b], true, false)
		}
	} 

	gradients.BiasGradient = gradient
	gradients.WeightGradient = d.utils.Outer2d(gradients.BiasGradient, input)
	gradients.PrevLayerGradient = make([]*ckks.Ciphertext, d.InputUnit)
	
	if hasPrevLayer {

		// Calculate ∂L/∂A(l-1)
		transposedWeight := d.utils.Transpose2d(d.Weights)

		for row := range transposedWeight{
			
			for datapoint := range gradients.BiasGradient{

				grad := d.utils.MultiplyNew(*transposedWeight[row][datapoint], *gradients.BiasGradient[datapoint], true, false)

				if gradients.PrevLayerGradient[row] == nil{
					gradients.PrevLayerGradient[row] = &grad
				} else {
					d.utils.Add(*gradients.PrevLayerGradient[row], grad, gradients.PrevLayerGradient[row])
				}

			}
		}

	}

	return gradients

}