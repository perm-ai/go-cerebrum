package ml

import (

	"github.com/ldsec/lattigo/v2/ckks"
	// "github.com/perm-ai/GO-HEML-prototype/src/logger"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"

)

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

func (model Dense) Forward(input *ckks.Ciphertext) ckks.Ciphertext{

	results := make([]ckks.Ciphertext, model.OutputUnit)

	// Loop through each node in the layer and apply it to the data
	for node := 0; node < model.OutputUnit; node++ {

		results[node] = model.utils.DotProductNew(*input, model.Weights[node], false)

	}

	result := model.utils.PackVector(results)

	model.utils.Add(result, model.Bias, &result)

	return result

}

func (model Dense) Backward(input *ckks.Ciphertext, output *ckks.Ciphertext, gradient ckks.Ciphertext, nextLayer *Dense) NeuralNetworkGradient {

	gradients := NeuralNetworkGradient{}

	if !nextLayer.Declared {

		if model.UseActivation {

			activationGradient := model.Activation.backward(*output)
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
			activationGradient := model.Activation.backward(*output)
			gradients.BiasGradient = model.utils.MultiplyNew(activationGradient, sum, true, false)
		} else {
			gradients.BiasGradient = sum
		}

		// Calculate weight gradient by doing outer product
		gradients.WeightGradient = model.utils.Outer(&gradients.BiasGradient, input, model.OutputUnit, model.InputUnit)

		return gradients

	}

}