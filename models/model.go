package models

import (
	"fmt"
	"math"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/importer"
	"github.com/perm-ai/go-cerebrum/utility"
	"github.com/perm-ai/go-cerebrum/layers"
	"github.com/perm-ai/go-cerebrum/activations"
	"github.com/perm-ai/go-cerebrum/losses"
)

//=================================================
//			  NEURAL NETWORK GRADIENT
//=================================================

type NeuralNetworkGradient struct {
	BiasGradient   ckks.Ciphertext
	WeightGradient []ckks.Ciphertext
}

func (n NeuralNetworkGradient) GroupGradients(utils utility.Utils, weightLength int, biasLenght int) utility.CiphertextGroup {

	ctData := make([]utility.CiphertextData, len(n.WeightGradient)+1)

	for i := range n.WeightGradient {

		ctData[i] = utility.CiphertextData{Ciphertext: n.WeightGradient[i], Length: weightLength}

	}

	ctData[len(n.WeightGradient)] = utility.CiphertextData{Ciphertext: n.BiasGradient, Length: biasLenght}

	return utility.NewCiphertextGroup(ctData, utils)

}

func (n *NeuralNetworkGradient) LoadFromGroup(group utility.CiphertextGroup, rescale bool) {

	ct := group.BreakGroup(rescale)

	n.WeightGradient = ct[0 : len(ct)-1]
	n.BiasGradient = ct[len(ct)-1]

}


//=================================================
//						MODEL
//=================================================

type Model struct {
	utils          utility.Utils
	Layers         []layers.Dense
	Loss           losses.Loss
	transposeCache []map[int][]ckks.Ciphertext
}

func NewModel(utils utility.Utils) Model {

	modelLayers := make([]layers.Dense, 3)

	tanh := activations.Tanh{U: utils}
	softmax := activations.Softmax{U: utils}
	crossEntropy := losses.CrossEntropy{U: utils}

	modelLayers[0] = layers.NewDense(784, 128, tanh, true, utils)
	modelLayers[1] = layers.NewDense(128, 64, tanh, true, utils)
	modelLayers[2] = layers.NewDense(64, 10, softmax, true, utils)

	transposeCache := make([]map[int][]ckks.Ciphertext, len(modelLayers))

	return Model{utils: utils, Layers: modelLayers, Loss: crossEntropy, transposeCache: transposeCache}

}

func (m Model) Forward(input ckks.Ciphertext) map[string]*ckks.Ciphertext {

	outputs := map[string]*ckks.Ciphertext{}

	outputs["A0"] = input.CopyNew()

	z1, a1 := m.Layers[0].Forward(outputs["A0"])
	outputs["Z1"] = &z1
	outputs["A1"] = &a1

	z2, a2 := m.Layers[0].Forward(outputs["A1"])
	outputs["Z2"] = &z2
	outputs["A2"] = &a2

	z3, a3 := m.Layers[0].Forward(outputs["A2"])
	outputs["Z3"] = &z3
	outputs["A3"] = &a3

	return outputs

}

func (m Model) Backward(outputs map[string]*ckks.Ciphertext, y ckks.Ciphertext, lr float64, batchNumber int) []NeuralNetworkGradient {

	// Calculate gradient from loss function
	gradient := m.Loss.Backward(*outputs["A3"], y, m.Layers[2].OutputUnit)

	// Create array to store gradients of each layer
	denseGradients := make([]NeuralNetworkGradient, len(m.Layers))

	// Create empty ciphertext array for last layer that doesn't need next layer's transposed weight
	empty := []ckks.Ciphertext{}

	// Loop through each layer to calculate backward gradient
	for layer := len(m.Layers) - 1; layer >= 0; layer-- {

		// Get the string that map to input and output of the layer (activation of last layer as in, non-activated output as out)
		layerIn := fmt.Sprintf("A%d", layer)
		layerOut := fmt.Sprintf("Z%d", layer+1)

		// Check if last layer
		if layer == len(m.Layers)-1 {

			// If last layer put in empty transposed weight and calculate backward gradient
			denseGradients[layer] = m.Layers[layer].Backward(outputs[layerIn], outputs[layerOut], gradient, empty, lr)

		} else {

			// Check if there is cached transposed weight
			if _, ok := m.transposeCache[layer+1][batchNumber]; !ok {

				// If not exist, create one and add to cache
				m.transposeCache[layer+1][batchNumber] = m.utils.Transpose(m.Layers[layer+1].Weights, m.Layers[layer+1].InputUnit)

			}

			// Pull transposed weight from cache
			transposedWeight := m.transposeCache[layer+1][batchNumber]

			// Calculate backward gradient
			denseGradients[layer] = m.Layers[layer].Backward(outputs[layerIn], outputs[layerOut], gradient, transposedWeight, lr)

		}

	}

	return denseGradients

}

func (m *Model) UpdateGradient(gradients []NeuralNetworkGradient) {

	for i, layer := range m.Layers {

		layer.UpdateGradient(gradients[i])

	}

}

func (m Model) Train(dataLoader importer.MnistDataLoader, learningRate float64, batchSize int, miniBatchSize int, epoch int) {

	// TODO: find the best way to incorperate learning rate into gradient

	if !(batchSize%miniBatchSize == 0) {
		panic("Batch size must be divisable by mini batch size")
	}

	// Loop through each epoch
	for e := 0; e < epoch; e++ {

		totalBatches := dataLoader.TrainingDataPoint / batchSize

		// Loop through each batch
		for batch := 0; batch < totalBatches; batch++ {

			batchData := dataLoader.GetDataAsBatch(batch, batchSize)
			totalMiniBatches := batchSize / miniBatchSize
			miniBatchGradient := make([][]NeuralNetworkGradient, totalMiniBatches)

			// Loop through each mini bacth
			for miniBatch := 0; miniBatch < totalMiniBatches; miniBatch++ {

				// Store the sum of neural netowrk gradient
				backwardGradients := make([][]NeuralNetworkGradient, miniBatchSize)

				// Loop through each data in mini batch
				for i, data := range batchData[(miniBatch * miniBatchSize):(miniBatch + miniBatchSize)] {

					// Calculate forward outputs
					forwardOutputs := m.Forward(data.Image)

					// Get gradients
					backwardOutputs := m.Backward(forwardOutputs, data.Label, learningRate, batch)

					// Save gradients to array
					backwardGradients[i] = backwardOutputs

				}

				// Calculate minibatch average gradient
				averagedMiniBatch := m.AverageNeuralNetworkGradients(backwardGradients, true)

				// Bootstrap efficiently by combining weights of each node into one ciphertext using CiphertextGroup struct
				for layer := range averagedMiniBatch {

					if layer == 0 {

						// Layer that require highest level possible weight gradient will be bootstrapped normally
						// since ungrouping require 1 multiplicative depth

						// Can be optimized further by bootstrapping more when forward propagating making the calculation starts at level 8
						for i := range averagedMiniBatch[layer].WeightGradient {
							m.utils.BootstrapInPlace(&averagedMiniBatch[layer].WeightGradient[i])
						}

						m.utils.BootstrapInPlace(&averagedMiniBatch[layer].BiasGradient)

					} else {

						// Group weights into a ciphertext and bootstrap
						grouped := averagedMiniBatch[layer].GroupGradients(m.utils, m.Layers[layer].InputUnit, m.Layers[layer].OutputUnit)
						grouped.Bootstrap()
						averagedMiniBatch[layer].LoadFromGroup(grouped, true)

					}

				}

				// Add this bootstrapped minibatch gradient into array
				miniBatchGradient[miniBatch] = averagedMiniBatch
			}

			// Average batch gradient
			batchGradientAverage := m.AverageNeuralNetworkGradients(miniBatchGradient, false)

			// Update model's weights and biases
			m.UpdateGradient(batchGradientAverage)

		}

	}

}

func (m Model) AverageNeuralNetworkGradients(gradients [][]NeuralNetworkGradient, rescale bool) []NeuralNetworkGradient {

	var result []NeuralNetworkGradient

	for i, gradient := range gradients {

		if i == 0 {
			result = gradient
		} else {

			for layer := range result {
				// Combine mini batch bias gradient
				m.utils.Add(result[layer].BiasGradient, gradient[layer].BiasGradient, &result[layer].BiasGradient)

				// Loop through each node weight gradient and add
				for weightIndex := range result[layer].WeightGradient {
					m.utils.Add(result[layer].WeightGradient[weightIndex], gradient[layer].WeightGradient[weightIndex], &result[layer].WeightGradient[weightIndex])
				}

			}

		}

	}

	for layer := range result {

		m.utils.MultiplyConst(&result[layer].BiasGradient, (1 / float64(len(gradients))), &result[layer].BiasGradient, rescale, false)

		// Calculate scale that will allow bootstrapping at level 0
		desiredScale := math.Exp2(math.Round(math.Log2(float64(m.utils.Params.Q()[0]) / m.utils.Bootstrapper.MessageRatio)))

		averager := ckks.NewPlaintext(m.utils.Params, m.utils.Params.MaxLevel(), desiredScale)

		if result[layer].WeightGradient[0].Level() == 1 {

			// Encode plaintext of 1/n with scale that when multiply with weight gradient will result in ct with desired scale
			m.utils.Encoder.EncodeNTT(averager, m.utils.Float64ToComplex128(m.utils.GenerateFilledArraySize(1/float64(len(gradients)), m.Layers[layer].InputUnit)), m.utils.Params.LogSlots())

		}

		for _, weight := range result[layer].WeightGradient {

			var averager ckks.Plaintext

			if weight.Level() == 1 {

				m.utils.MultiplyPlain(&weight, &averager, &weight, false, false)

			} else {

				m.utils.MultiplyConst(&weight, 1/float64(len(gradients)), &weight, rescale, false)

			}

		}

	}

	return result

}
