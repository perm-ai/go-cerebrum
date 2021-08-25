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
	BiasGradient      []*ckks.Ciphertext
	WeightGradient    [][]*ckks.Ciphertext
	PrevLayerGradient []*ckks.Ciphertext
}

//=================================================
//					DENSE LAYER
//=================================================

type Dense struct {
	utils      utility.Utils
	InputUnit  int
	OutputUnit int
	Weights    [][]*ckks.Ciphertext
	Bias       []*ckks.Ciphertext
	Activation *activations.Activation
	batchSize  int
}

func NewDense(utils utility.Utils, inputUnit int, outputUnit int, activation *activations.Activation, useBias bool, batchSize int) Dense {

	// Generate random weights and biases
	weights := make([][]*ckks.Ciphertext, outputUnit)
	bias := make([]*ckks.Ciphertext, outputUnit)

	randomBias := array.GenerateRandomNormalArray(outputUnit)

	for node := 0; node < outputUnit; node++ {

		randomWeight := array.GenerateRandomNormalArray(inputUnit)
		weights[node] = make([]*ckks.Ciphertext, inputUnit)

		if useBias {
			bias[node] = utils.EncryptToPointer(utils.GenerateFilledArraySize(randomBias[node], batchSize))
		}

		for weight := 0; weight < inputUnit; weight++ {

			weights[node][weight] = utils.EncryptToPointer(utils.GenerateFilledArray(randomWeight[weight]))

		}

	}

	return Dense{utils, inputUnit, outputUnit, weights, bias, activation, batchSize}

}

func (d Dense) Forward(input []*ckks.Ciphertext) []*ckks.Ciphertext {

	output := make([]*ckks.Ciphertext, d.OutputUnit)

	for node := range d.Weights {

		output[node] = d.utils.InterDotProduct(input, d.Weights[node], true, false)

		if len(d.Bias) != 0 {
			d.utils.Add(*output[node], *d.Bias[node], output[node])
		}

	}

	return output

}

// input is A(l-1) - activation of previous layer
// output is Z(l) - output of this layer
// gradient is ∂L/∂A(l) - influence that the activation of this layer has on the next layer
// hasPrevLayer determine whether this function calculates the gradient ∂L/∂A(l-1)
func (d *Dense) Backward(input []*ckks.Ciphertext, output []*ckks.Ciphertext, gradient []*ckks.Ciphertext, hasPrevLayer bool) DenseGradient {

	gradients := DenseGradient{}

	// Calculate gradients for last layer
	if d.Activation != nil {
		for b := range d.Bias {
			activationGradient := (*d.Activation).Backward(*output[b], d.OutputUnit)
			d.utils.Multiply(*gradient[b], activationGradient, gradient[b], true, false)
		}
	}

	gradients.BiasGradient = gradient
	gradients.WeightGradient = d.utils.InterOuter(gradients.BiasGradient, input)
	gradients.PrevLayerGradient = make([]*ckks.Ciphertext, d.InputUnit)

	if hasPrevLayer {

		// Calculate ∂L/∂A(l-1)
		transposedWeight := d.utils.InterTranspose(d.Weights)

		for xi := range transposedWeight {
			gradients.PrevLayerGradient[xi] = d.utils.InterDotProduct(transposedWeight[xi], gradients.BiasGradient, true, false)
		}

	}

	return gradients

}

func (d *Dense) UpdateGradient(gradient DenseGradient, lr float64){

	batchAverager := d.utils.EncodePlaintextFromArray(d.utils.GenerateFilledArraySize(lr / float64(d.batchSize), d.batchSize))

	for node := range d.Weights{

		if len(d.Bias) != 0{
			averagedLrBias := d.utils.MultiplyPlainNew(gradient.BiasGradient[node], &batchAverager, true, false)
			d.utils.Sub(*d.Bias[node], averagedLrBias, d.Bias[node])
		}

		for w := range d.Weights[node]{
			averagedLrWeight := d.utils.MultiplyPlainNew(gradient.WeightGradient[node][w], &batchAverager, true, false)
			d.utils.Sub(*d.Weights[node][w], averagedLrWeight, d.Weights[node][w])
		}

	}

}