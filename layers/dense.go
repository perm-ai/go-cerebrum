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
//					DENSE LAYER
//=================================================

type Dense struct {
	utils          utility.Utils
	InputUnit      int
	OutputUnit     int
	Weights        [][]*ckks.Ciphertext
	Bias           []*ckks.Ciphertext
	Activation     *activations.Activation
	btspOutput     []bool
	btspActivation []bool
	batchSize      int
}

func NewDense(utils utility.Utils, inputUnit int, outputUnit int, activation *activations.Activation, useBias bool, batchSize int) Dense {

	// Generate random weights and biases
	weights := make([][]*ckks.Ciphertext, outputUnit)
	bias := make([]*ckks.Ciphertext, outputUnit)

	// Determine the standard deviation of initial random weight distribution
	weightStdDev := 0.0

	if (*activation).GetType() == "relu" {
		weightStdDev = math.Sqrt(1.0 / float64(inputUnit))
	} else {
		weightStdDev = math.Sqrt(1.0 / float64(inputUnit+outputUnit))
	}

	randomBias := array.GeneratePlainArray(0.0, outputUnit)

	for node := 0; node < outputUnit; node++ {

		randomWeight := array.GenerateRandomNormalArray(inputUnit, weightStdDev)
		weights[node] = make([]*ckks.Ciphertext, inputUnit)

		if useBias {
			bias[node] = utils.EncryptToPointer(utils.GenerateFilledArraySize(randomBias[node], batchSize))
		}

		for weight := 0; weight < inputUnit; weight++ {

			weights[node][weight] = utils.EncryptToPointer(utils.GenerateFilledArray(randomWeight[weight]))

		}

	}

	return Dense{utils, inputUnit, outputUnit, weights, bias, activation, []bool{false, false}, []bool{false, false}, batchSize}

}

func (d Dense) Forward(input []*ckks.Ciphertext) Output1d {

	output := make([]*ckks.Ciphertext, d.OutputUnit)
	activatedOutput := make([]*ckks.Ciphertext, d.OutputUnit)

	for node := range d.Weights {

		output[node] = d.utils.InterDotProduct(input, d.Weights[node], true, false, true)

		if len(d.Bias) != 0 {
			d.utils.Add(*output[node], *d.Bias[node], output[node])
		}

		if d.btspOutput[0] {
			d.utils.BootstrapInPlace(output[node])
		}

	}

	if d.Activation != nil {
		activatedOutput = (*d.Activation).Forward(output, d.batchSize)

		if d.btspActivation[0] {
			for a := range activatedOutput {
				d.utils.BootstrapInPlace(activatedOutput[a])
			}
		}

	}

	return Output1d{output, activatedOutput}

}

// input is A(l-1) - activation of previous layer
// output is Z(l) - output of this layer
// gradient is ∂L/∂A(l) - influence that the activation of this layer has on the next layer
// hasPrevLayer determine whether this function calculates the gradient ∂L/∂A(l-1)
func (d *Dense) Backward(input []*ckks.Ciphertext, output []*ckks.Ciphertext, gradient []*ckks.Ciphertext, hasPrevLayer bool) Gradient1d {

	gradients := Gradient1d{}

	// Calculate gradients for last layer
	if d.Activation != nil {

		if (*d.Activation).GetType() != "softmax" {
			activationGradient := (*d.Activation).Backward(output, d.OutputUnit)

			hasBootstrapped := false

			if activationGradient[0].Level() == 1 && d.btspActivation[1] {
				d.utils.Bootstrap1dInPlace(activationGradient, true)
				hasBootstrapped = true
			}

			var wg sync.WaitGroup

			channels := make([]chan *ckks.Ciphertext, len(d.Bias))
			for b := range d.Bias {

				wg.Add(1)
				channels[b] = make(chan *ckks.Ciphertext)

				go func(index int, utils utility.Utils) {
					defer wg.Done()
					utils.Multiply(*gradient[index], *activationGradient[index], gradient[index], true, false)

				}(b, d.utils.CopyWithClonedEval())

			}

			wg.Wait()

			if d.btspActivation[1] && !hasBootstrapped {
				d.utils.Bootstrap1dInPlace(activationGradient, true)
			}

		}

	}

	gradients.BiasGradient = gradient
	gradients.WeightGradient = d.utils.InterOuter(gradients.BiasGradient, input, true)
	gradients.InputGradient = make([]*ckks.Ciphertext, d.InputUnit)

	if hasPrevLayer {

		// Calculate ∂L/∂A(l-1)
		transposedWeight := d.utils.InterTranspose(d.Weights)

		for xi := range transposedWeight {
			gradients.InputGradient[xi] = d.utils.InterDotProduct(transposedWeight[xi], gradients.BiasGradient, true, false, true)
			if d.btspOutput[1] {
				d.utils.BootstrapInPlace(gradients.InputGradient[xi])
			}
		}

	}

	return gradients

}

func (d *Dense) UpdateGradient(gradient Gradient1d, lr float64) {

	batchAverager := d.utils.EncodePlaintextFromArray(d.utils.GenerateFilledArraySize(lr/float64(d.batchSize), d.batchSize))

	for node := range d.Weights {

		if len(d.Bias) != 0 {
			d.utils.SumElementsInPlace(gradient.BiasGradient[node])
			averagedLrBias := d.utils.MultiplyPlainNew(gradient.BiasGradient[node], &batchAverager, true, false)
			d.utils.Sub(*d.Bias[node], averagedLrBias, d.Bias[node])
		}

		for w := range d.Weights[node] {
			d.utils.SumElementsInPlace(gradient.WeightGradient[node][w])
			averagedLrWeight := d.utils.MultiplyPlainNew(gradient.WeightGradient[node][w], &batchAverager, true, false)
			d.utils.Sub(*d.Weights[node][w], averagedLrWeight, d.Weights[node][w])
		}

	}

}

func (d Dense) GetOutputSize() int {
	return d.OutputUnit
}

func (d Dense) IsTrainable() bool {
	return true
}

func (d Dense) HasActivation() bool {
	return d.Activation != nil
}

func (d Dense) GetForwardLevelConsumption() int {
	return 1
}

func (d Dense) GetBackwardLevelConsumption() int {
	return 1
}

func (d Dense) GetForwardActivationLevelConsumption() int {
	if d.HasActivation() {
		return (*d.Activation).GetForwardLevelConsumption()
	} else {
		return 0
	}
}

func (d Dense) GetBackwardActivationLevelConsumption() int {
	if d.HasActivation() {
		return (*d.Activation).GetBackwardLevelConsumption()
	} else {
		return 0
	}
}

func (d *Dense) SetBootstrapOutput(set bool, direction string) {
	switch direction {
	case "forward":
		d.btspOutput[0] = set
	case "backward":
		d.btspOutput[1] = set
	}
}

func (d *Dense) SetBootstrapActivation(set bool, direction string) {
	switch direction {
	case "forward":
		d.btspActivation[0] = set
	case "backward":
		d.btspActivation[1] = set
	}
}
