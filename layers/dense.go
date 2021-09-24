package layers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"sync"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/activations"
	"github.com/perm-ai/go-cerebrum/array"
	"github.com/perm-ai/go-cerebrum/logger"
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
	weightLevel    int
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
	logger := logger.NewLogger(true)

	var wg sync.WaitGroup

	for node := 0; node < outputUnit; node++ {

		wg.Add(1)

		go func(nodeIndex int, u utility.Utils) {

			defer wg.Done()

			logger.Log(fmt.Sprintf("Generating weight for node %d", nodeIndex))
			randomWeight := array.GenerateRandomNormalArray(inputUnit, weightStdDev)
			weights[nodeIndex] = make([]*ckks.Ciphertext, inputUnit)

			if useBias {
				bias[nodeIndex] = u.EncryptToLevel(u.GenerateFilledArraySize(randomBias[nodeIndex], batchSize), 9)
			}

			var weightWg sync.WaitGroup

			for weight := 0; weight < inputUnit; weight++ {

				weightWg.Add(1)

				go func(weightIndex int, wUtils utility.Utils){
					defer weightWg.Done()
					weights[nodeIndex][weightIndex] = wUtils.EncryptToLevel(u.GenerateFilledArray(randomWeight[weightIndex]), 9)
				}(weight, u.CopyWithClonedEncryptor())

			}

			weightWg.Wait()

		}(node, utils.CopyWithClonedEncryptor())

	}

	wg.Wait()

	return Dense{utils, inputUnit, outputUnit, weights, bias, activation, []bool{false, false}, []bool{false, false}, batchSize, 9}

}

func (d Dense) Forward(input []*ckks.Ciphertext) Output1d {

	output := make([]*ckks.Ciphertext, d.OutputUnit)
	activatedOutput := make([]*ckks.Ciphertext, d.OutputUnit)

	for node := range d.Weights {

		fmt.Printf("Starting dot product for node %d\n", node)
		output[node] = d.utils.InterDotProduct(input, d.Weights[node], true, false, true)
		fmt.Printf("Dot product for node %d completed\n", node)

		if len(d.Bias) != 0 {
			d.utils.Add(*output[node], *d.Bias[node], output[node])
		}

	}

	if d.btspOutput[0] {
		fmt.Printf("Bootstrapping node\n")
		d.utils.Bootstrap1dInPlace(output, true)
		fmt.Printf("Bootstrapping node completed\n")
	}

	if d.Activation != nil {
		activatedOutput = (*d.Activation).Forward(output, d.batchSize)

		if d.btspActivation[0] {
			d.utils.Bootstrap1dInPlace(activatedOutput, true)
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

	// create weight group
	var wg sync.WaitGroup

	for node := range d.Weights {

		wg.Add(1)

		go func(nodeIndex int, utils utility.Utils) {

			defer wg.Done()

			updatedBiasChannel := make(chan *ckks.Ciphertext)

			if len(d.Bias) != 0 {

				// Calculate updated bias concurrently
				go func(utils utility.Utils, c chan *ckks.Ciphertext) {

					utils.SumElementsInPlace(gradient.BiasGradient[nodeIndex])
					averagedLrBias := utils.MultiplyPlainNew(gradient.BiasGradient[nodeIndex], &batchAverager, true, false)

					if averagedLrBias.Level() < d.weightLevel {
						utils.BootstrapInPlace(&averagedLrBias)
					}

					result := utils.SubNew(*d.Bias[nodeIndex], averagedLrBias)
					c <- &result

				}(d.utils.ShallowCopy(), updatedBiasChannel)

			}

			updatedWeightChannels := make([]chan *ckks.Ciphertext, len(d.Weights[nodeIndex]))

			for w := range d.Weights[nodeIndex] {

				// Update weight concurrently
				updatedWeightChannels[w] = make(chan *ckks.Ciphertext)

				go func(weightIndex int, utils utility.Utils, c chan *ckks.Ciphertext) {

					utils.SumElementsInPlace(gradient.WeightGradient[nodeIndex][weightIndex])
					averagedLrWeight := utils.MultiplyPlainNew(gradient.WeightGradient[nodeIndex][weightIndex], &batchAverager, true, false)

					// Bootstrap if gradient's level is lower than it's suppose to be
					if averagedLrWeight.Level() < d.weightLevel {
						utils.BootstrapInPlace(&averagedLrWeight)
					}

					result := utils.SubNew(*d.Weights[nodeIndex][weightIndex], averagedLrWeight)
					c <- &result

				}(w, d.utils.ShallowCopy(), updatedWeightChannels[w])

			}

			if len(d.Bias) != 0 {
				d.Bias[nodeIndex] = <-updatedBiasChannel
			}

			for w := range updatedWeightChannels {
				d.Weights[nodeIndex][w] = <-updatedWeightChannels[w]
			}

		}(node, d.utils.ShallowCopy())

	}

	wg.Wait()

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

func (d *Dense) SetWeightLevel(lvl int) {
	d.weightLevel = lvl
}

type denseWeight struct {
	Weight [][]float64
	Bias   []float64
}

func (d *Dense) ExportWeights(filename string) {

	plainWeights := make([][]float64, len(d.Weights))

	var wg sync.WaitGroup

	for node := range d.Weights {
		plainWeights[node] = make([]float64, len(d.Weights[node]))
		for i := range plainWeights[node] {
			wg.Add(1)
			go func(nodeIdx int, weightIdx int, utils utility.Utils){
				defer wg.Done()
				plainWeights[nodeIdx][weightIdx] = utils.Decrypt(d.Weights[nodeIdx][weightIdx])[0]
			}(node, i, d.utils.CopyWithClonedDecryptor())
		}
	}

	wg.Wait()

	bias := make([]float64, len(d.Bias))

	for node := range bias {
		bias[node] = d.utils.Decrypt(d.Bias[node])[0]
	}

	weight := denseWeight{plainWeights, bias}

	file, _ := json.MarshalIndent(weight, "", " ")

	_ = ioutil.WriteFile(filename, file, 0644)

}
