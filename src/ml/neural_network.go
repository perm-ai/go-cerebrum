package ml

import (

	"github.com/ldsec/lattigo/v2/ckks"
	// "github.com/perm-ai/GO-HEML-prototype/src/logger"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"

)

//=================================================
//					DENSE LAYER
//=================================================

type Dense struct {

	InputUnit		int
	OutputUnit		int
	Weights			[]ckks.Ciphertext
	Bias			ckks.Ciphertext
	Activation 		Activation
	UseActivation	bool
	utils 			utility.Utils

	Declared		bool

}

type NeuralNetworkGradient struct {

	BiasGradient	ckks.Ciphertext
	WeightGradient	[]ckks.Ciphertext

}

func NewDense(inputUnit int, outputUnit int, activation Activation, useActivation bool, utils utility.Utils) Dense {

	weights := make([]ckks.Ciphertext, outputUnit)

	for i := range weights{

		weights[i] = utils.Encrypt(utils.GenerateRandomNormalArray(inputUnit))

	}

	bias := utils.Encrypt(utils.GenerateRandomNormalArray(outputUnit))

	return Dense{inputUnit, outputUnit, weights, bias, activation, useActivation, utils, true}

}

func (model Dense) Forward(input *ckks.Ciphertext) (ckks.Ciphertext, ckks.Ciphertext){

	results := make([]ckks.Ciphertext, model.OutputUnit)

	// Loop through each node in the layer and apply it to the data
	for node := 0; node < model.OutputUnit; node++ {

		results[node] = model.utils.DotProductNew(*input, model.Weights[node], false)

	}

	z := model.utils.PackVector(results)
	model.utils.Add(z, model.Bias, &z)

	if model.UseActivation {
		a := model.Activation.Forward(z, model.OutputUnit)
		return z, a
	} else {
		return z, z
	}

}

func (model Dense) Backward(input *ckks.Ciphertext, output *ckks.Ciphertext, gradient ckks.Ciphertext, nextLayer *Dense) NeuralNetworkGradient {

	gradients := NeuralNetworkGradient{}

	if !nextLayer.Declared {

		if model.UseActivation {

			activationGradient := model.Activation.Backward(*output, model.OutputUnit)
			gradients.BiasGradient = model.utils.MultiplyNew(gradient, activationGradient, true, false)

		} else {

			gradients.BiasGradient = gradient

		}
		// Calculate outer product of bias_gradient and input to layer to calculate weight gradient
		gradients.WeightGradient = model.utils.Outer(&gradients.BiasGradient, input, model.OutputUnit, model.InputUnit)
		
		return gradients

	} else {

		transposedWeight := model.utils.Transpose(nextLayer.Weights, model.OutputUnit)

		var sum ckks.Ciphertext

		for i := range transposedWeight{
			
			if i == 0 {

				// Calculate dot product and replace sum with new ciphertext
				sum = model.utils.DotProductNew(transposedWeight[i], gradient, false)

				// Apply filter to dot product
				model.utils.MultiplyPlain(&sum, &model.utils.Filters[i], &sum, true, false)

			} else {

				// Calculate dot product
				product := model.utils.DotProductNew(transposedWeight[i], gradient, false)

				// Apply filter to dot product
				model.utils.MultiplyPlain(&product, &model.utils.Filters[i], &product, true, false)

				// Add filtered product to sum ciphertext
				model.utils.Add(product, sum, &sum)

			}

		}

		// Apply activation gradient
		if model.UseActivation {
			activationGradient := model.Activation.Backward(*output, model.OutputUnit)
			gradients.BiasGradient = model.utils.MultiplyNew(activationGradient, sum, true, false)
		} else {
			gradients.BiasGradient = sum
		}

		// Calculate weight gradient by doing outer product
		gradients.WeightGradient = model.utils.Outer(&gradients.BiasGradient, input, model.OutputUnit, model.InputUnit)

		return gradients

	}

}

//=================================================
//						MODEL
//=================================================

type Model struct {

	Layers	[]Dense
	Loss 	Loss

}

func NewModel(utils utility.Utils) Model {

	layers := make([]Dense, 3)

	tanh := Tanh{utils: utils}
	softmax := Softmax{utils: utils}
	crossEntropy := CrossEntropy{utils: utils}

	layers[0] = NewDense(784, 128, tanh, true, utils)
	layers[1] = NewDense(128, 64, tanh, true, utils)
	layers[2] = NewDense(64, 10, softmax, true, utils)

	return Model{layers, crossEntropy}

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

func (m Model) Backward(outputs map[string]*ckks.Ciphertext, y ckks.Ciphertext, lr float64) []NeuralNetworkGradient {

	gradient := m.Loss.Backward(*outputs["A3"], y, m.Layers[2].OutputUnit)

	denseGradients := make([]NeuralNetworkGradient, len(m.Layers))

	denseGradients[2] = m.Layers[2].Backward(outputs["A2"], outputs["Z3"], gradient, &Dense{})
	denseGradients[1] = m.Layers[2].Backward(outputs["A2"], outputs["Z2"], denseGradients[2].BiasGradient, &m.Layers[2])
	denseGradients[0] = m.Layers[2].Backward(outputs["A1"], outputs["Z1"], denseGradients[1].BiasGradient, &m.Layers[1])
	
	return denseGradients

}