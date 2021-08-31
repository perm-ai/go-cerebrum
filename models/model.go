package models

import (
	"fmt"
	"math"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/dataset"
	"github.com/perm-ai/go-cerebrum/layers"
	"github.com/perm-ai/go-cerebrum/losses"
	"github.com/perm-ai/go-cerebrum/utility"
)

//=================================================
//						MODEL
//=================================================

type Model struct {
	utils    utility.Utils
	Layers1d []layers.Layer1D
	Layers2d []layers.Layer2D
	Flatten  layers.Flatten2D
	Loss     losses.Loss
}

func NewModel(utils utility.Utils, layer1d []layers.Layer1D, layer2d []layers.Layer2D, loss losses.Loss) Model {

	var flatten layers.Flatten2D
	if len(layer2d) != 0 {
		flatten = layers.NewFlatten(layer2d[len(layer2d)-1].GetOutputSize())
	}

	model := Model{utils, layer1d, layer2d, flatten, loss}

	in1D, in2D := model.setForwardBootstrapping()

	model.setBackwardBootstrapping(in1D, in2D)

	return model

}

func (m Model) Forward(input2D [][][]*ckks.Ciphertext, input1D []*ckks.Ciphertext) ([]layers.Output2d, []layers.Output1d) {

	var output2D []layers.Output2d
	var output1D []layers.Output1d

	if len(m.Layers2d) != 0 {

		if input2D != nil {
			panic(fmt.Sprintf("Input2d is not given. Expect input with size %d", m.Layers2d[0].GetOutputSize()))
		}

		// Initialize slice for storing output
		output2D = make([]layers.Output2d, len(m.Layers2d)+1)

		// Insert output of input layer
		output2D[0] = layers.Output2d{Output: input2D}

		prevLayerHasActivation := false

		// Loop through each layer and calculate output
		for layer := 0; layer < len(m.Layers2d); layer++ {

			// input of this layer is output[i]
			// output of this layer i is stored at output[i+1]

			var prevOut [][][]*ckks.Ciphertext

			if prevLayerHasActivation {
				prevOut = output2D[layer].ActivationOutput
			} else {
				prevOut = output2D[layer].Output
			}

			output2D[layer+1] = m.Layers2d[layer].Forward(utility.Clone3dCiphertext(prevOut))

			prevLayerHasActivation = m.Layers2d[layer].HasActivation()

		}

	}

	// initialize slice for storing 1D output
	output1D = make([]layers.Output1d, len(m.Layers1d)+1)

	if output2D != nil {

		prevOut := output2D[len(output2D)-1].Output

		// Check if last 2D layer has activation
		if m.Layers2d[len(m.Layers2d)-1].HasActivation() {
			prevOut = output2D[len(output2D)-1].ActivationOutput
		}

		// Insert flattened 2D output in as input array
		output1D[0] = m.Flatten.Forward(utility.Clone3dCiphertext(prevOut))

	} else {
		// Insert output of input layer
		output1D[0] = layers.Output1d{Output: input1D}
	}

	if len(m.Layers1d) != 0 {

		prevLayerHasActivation := false

		for layer := 0; layer <= len(m.Layers1d); layer++ {

			var prevOut []*ckks.Ciphertext

			if prevLayerHasActivation {
				prevOut = output1D[layer].ActivationOutput
			} else {
				prevOut = output1D[layer].Output
			}

			output1D[layer+1] = m.Layers1d[layer].Forward(utility.Clone1dCiphertext(prevOut))

			prevLayerHasActivation = m.Layers1d[layer].HasActivation()

		}

	}

	return output2D, output1D

}

func (m Model) Backward(output2D []layers.Output2d, output1D []layers.Output1d, y []*ckks.Ciphertext) ([]layers.Gradient2d, []layers.Gradient1d) {

	gradient1D := make([]layers.Gradient1d, len(m.Layers1d)+1)
	gradient2D := make([]layers.Gradient2d, len(m.Layers2d)+1)

	// Calculate loss gradient
	var finalOutput []*ckks.Ciphertext

	// Get output of last layer
	if len(m.Layers1d) != 0 {
		if m.Layers1d[len(m.Layers1d)-1].HasActivation() {
			finalOutput = output1D[len(m.Layers1d)].ActivationOutput
		} else {
			finalOutput = output1D[len(m.Layers1d)].Output
		}
	} else {
		finalOutput = output1D[len(output1D)-1].Output
	}

	gradient1D[len(gradient1D)-1] = layers.Gradient1d{InputGradient: m.Loss.Backward(utility.Clone1dCiphertext(finalOutput), utility.Clone1dCiphertext(y), len(y))}

	if len(m.Layers1d) != 0 {

		for layer := len(m.Layers1d) - 1; layer >= 0; layer-- {

			// input of layer index i is output[i]
			// output of layer i is output[i+1]

			// Get next layer's input gradient
			nextLayerInputGrad := gradient1D[layer+1].InputGradient

			// Get layer's input
			layerInput := output1D[layer].Output
			if layer != 0 {
				if m.Layers1d[layer-1].HasActivation() {
					layerInput = output1D[layer].ActivationOutput
				}
			}

			// Get layer's output
			layerOutput := output1D[layer+1].Output
			if m.Layers1d[layer].HasActivation() {
				layerOutput = output1D[layer+1].ActivationOutput
			}

			gradient1D[layer] = m.Layers1d[layer].Backward(utility.Clone1dCiphertext(layerInput), utility.Clone1dCiphertext(layerOutput), utility.Clone1dCiphertext(nextLayerInputGrad), (layer != 0 || len(m.Layers2d) != 0))

		}

	}

	if len(m.Layers2d) != 0 {

		// Backward flatten layer
		gradient2D[len(gradient2D)-1] = m.Flatten.Backward(utility.Clone1dCiphertext(gradient1D[0].InputGradient))

		for layer := len(m.Layers2d) - 1; layer >= 0; layer-- {

			// input of layer index i is output[i]
			// output of layer i is output[i+1]

			// Get next layer's input gradient
			nextLayerInputGrad := gradient2D[layer+1].InputGradient

			// Get layer's input
			layerInput := output2D[layer].Output
			if layer != 0 {
				if m.Layers2d[layer-1].HasActivation() {
					layerInput = output2D[layer].ActivationOutput
				}
			}

			// Get layer's output
			layerOutput := output2D[layer+1].Output
			if m.Layers2d[layer].HasActivation() {
				layerOutput = output2D[layer+1].ActivationOutput
			}

			gradient2D[layer] = m.Layers2d[layer].Backward(utility.Clone3dCiphertext(layerInput), utility.Clone3dCiphertext(layerOutput), utility.Clone3dCiphertext(nextLayerInputGrad), (layer != 0 || len(m.Layers2d) != 0))

		}

	}

	return gradient2D, gradient1D

}

func (m *Model) UpdateGradient(gradients1d []layers.Gradient1d, gradients2d []layers.Gradient2d, lr float64) {

	for layer := range m.Layers2d {
		if m.Layers2d[layer].IsTrainable() {
			m.Layers2d[layer].UpdateGradient(gradients2d[layer], lr)
		}
	}

	for layer := range m.Layers1d {
		if m.Layers1d[layer].IsTrainable() {
			m.Layers1d[layer].UpdateGradient(gradients1d[layer], lr)
		}
	}

}

func (m *Model) Train2D(dataLoader dataset.Loader, learningRate float64, batchSize int, epoch int) {

	for i := 0; i < int(dataLoader.GetLength()/batchSize); i++ {

		x, y := dataLoader.Load2D(i*batchSize, batchSize)

		outputs2D, outputs1D := m.Forward(x, []*ckks.Ciphertext{})
		gradients2D, gradients1D := m.Backward(outputs2D, outputs1D, y)
		m.UpdateGradient(gradients1D, gradients2D, learningRate)

	}

}

func (m Model) setForwardBootstrapping() ([]int, []int) {

	inputLevel2D := make([]int, len(m.Layers2d)+1)
	inputLevel1D := make([]int, len(m.Layers1d)+1)
	inputLevel2D[0] = 9

	// Calculate when to bootstrap for 2D forward propagation
	for l := 0; l < len(m.Layers2d); l++ {

		// Calculate required level for this layer
		requiredLevel := m.Layers2d[l].GetForwardLevelConsumption() + m.Layers2d[l].GetForwardActivationLevelConsumption()

		if (inputLevel2D[l] - requiredLevel) < 1 {

			// If not enough level bootstrap output of previous layer
			if m.Layers2d[l-1].HasActivation() {
				m.Layers2d[l-1].SetBootstrapActivation(true, "forward")
			} else {
				m.Layers2d[l-1].SetBootstrapOutput(true, "forward")
			}

			// Set input to this layer to 9 (highest)
			inputLevel2D[l] = 9

		}

		// Set output level of this layer (input of next layer)
		inputLevel2D[l+1] = inputLevel2D[l] - requiredLevel

	}

	if len(m.Layers2d) == 0 {
		inputLevel1D[0] = 9
	} else {
		inputLevel1D[0] = inputLevel2D[len(inputLevel2D)-1]
	}

	// Calculate when to bootstrap for 1D forward propagation
	for l := 0; l < len(m.Layers1d); l++ {

		// Calculate required level for this layer
		requiredLevel := m.Layers1d[l].GetForwardLevelConsumption() + m.Layers1d[l].GetForwardActivationLevelConsumption()

		if (inputLevel1D[l] - requiredLevel) < 1 {

			// If not enough level bootstrap output of previous layer
			if m.Layers1d[l-1].HasActivation() {
				m.Layers1d[l-1].SetBootstrapActivation(true, "forward")
			} else {
				m.Layers1d[l-1].SetBootstrapOutput(true, "forward")
			}

			// Set input to this layer to 9 (highest)
			inputLevel1D[l] = 9

		}

		// Set output level of this layer (input of next layer)
		inputLevel1D[l+1] = inputLevel1D[l] - requiredLevel

	}

	// Set bootstrap output for last layer
	if m.Layers1d[len(m.Layers1d)-1].HasActivation() {
		m.Layers1d[len(m.Layers1d)-1].SetBootstrapActivation(true, "forward")
	} else {
		m.Layers1d[len(m.Layers1d)-1].SetBootstrapOutput(true, "forward")
	}

	inputLevel1D[len(inputLevel1D)-1] = 9

	return inputLevel1D, inputLevel2D

}

func (m Model) setBackwardBootstrapping(inputLevel1D []int, inputLevel2D []int) ([]int, []int) {

	gradientLevel1D := make([]int, len(m.Layers1d)+1)
	gradientLevel2D := make([]int, len(m.Layers2d)+1)

	gradientLevel1D[len(gradientLevel1D)-1] = inputLevel1D[len(inputLevel1D)-1]

	// Loop throught each layer backward
	for l := len(m.Layers1d) - 1; l >= 0; l-- {

		// Get the loss gradient wrt input on the next layer
		gradientLevel := gradientLevel1D[l+1]

		// Get level of input of this layer
		inputLevel := inputLevel1D[l]

		if m.Layers1d[l].GetBackwardActivationLevelConsumption() > 0 && m.Layers1d[l].IsTrainable() {

			activationLevel := inputLevel - m.Layers1d[l].GetForwardLevelConsumption() - m.Layers1d[l].GetBackwardActivationLevelConsumption()

			// Check if loss gradient wrt input has enough level to multiply once with gradient of activation wrt output
			if gradientLevel < 2 {
				// if not enough bootstrap output of previous layer
				m.Layers1d[l+1].SetBootstrapOutput(true, "backward")
				gradientLevel1D[l+1] = 9
				gradientLevel = gradientLevel1D[l+1]
			}

			activationLossGradientLevel := 0
			if activationLevel > 1 {
				// Calculate level of loss gradient wrt output
				activationLossGradientLevel = int(math.Min(float64(gradientLevel), float64(activationLevel)) - 1)
			} else {
				// Bootstrap activation gradient wrt output before computing loss gradient wrt output if not enough level is reached
				m.Layers1d[l].SetBootstrapActivation(true, "backward")
				activationLossGradientLevel = int(math.Min(float64(gradientLevel), 9.0) - 1)
			}

			// Calculate loss gradient wrt input
			gradientLevel1D[l] = int(math.Min(float64(inputLevel), float64(activationLossGradientLevel)) - float64(m.Layers1d[l].GetBackwardLevelConsumption()))

			// Check if loss gradient wrt input is less than one
			if gradientLevel1D[l] < 1 && activationLossGradientLevel < inputLevel {
				// Bootstrap loss gradient wrt output if it is responsible for making the level of loss wrt input less than 1
				m.Layers1d[l].SetBootstrapActivation(true, "backward")
				activationLossGradientLevel = 9

				// recalculate level of loss gradient wrt input
				gradientLevel1D[l] = int(math.Min(float64(inputLevel), float64(activationLossGradientLevel)) - float64(m.Layers1d[l].GetBackwardLevelConsumption()))
			}

		} else if m.Layers1d[l].IsTrainable() {

			// calculate level of loss gradient wrt input
			gradientLevel1D[l] = int(math.Min(float64(inputLevel), float64(gradientLevel)) - float64(m.Layers1d[l].GetBackwardLevelConsumption()))

			// Bootstrap loss gradient wrt input of next layer if level is not enough
			if gradientLevel1D[l] < 1 {
				m.Layers1d[l-1].SetBootstrapOutput(true, "backward")
				gradientLevel1D[l-1] = 9
				gradientLevel1D[l] = int(math.Min(float64(inputLevel), float64(gradientLevel1D[l-1])) - float64(m.Layers1d[l].GetBackwardLevelConsumption()))
			}

		} else {

			// calculate level of loss gradient wrt input
			gradientLevel1D[l] = gradientLevel - m.Layers1d[l].GetBackwardLevelConsumption()

			// Bootstrap loss gradient wrt input of next layer if level is not enough
			if gradientLevel1D[l] < 1 {
				m.Layers1d[l-1].SetBootstrapOutput(true, "backward")
				gradientLevel1D[l-1] = 9
				gradientLevel1D[l] = int(math.Min(float64(inputLevel), float64(gradientLevel1D[l-1])) - float64(m.Layers1d[l].GetBackwardLevelConsumption()))
			}

		}

	}

	gradientLevel2D[len(gradientLevel2D)-1] = gradientLevel1D[0]

	if len(m.Layers2d) == 0 {
		return gradientLevel1D, gradientLevel2D
	}

	// Loop throught each layer backward
	for l := len(m.Layers2d) - 1; l >= 0; l-- {

		// Get the loss gradient wrt input on the next layer
		gradientLevel := gradientLevel2D[l+1]

		// Get level of input of this layer
		inputLevel := inputLevel2D[l]

		if m.Layers2d[l].HasActivation() && m.Layers2d[l].IsTrainable() {

			activationLevel := inputLevel - m.Layers2d[l].GetForwardLevelConsumption() - m.Layers2d[l].GetBackwardActivationLevelConsumption()

			// Check if loss gradient wrt input has enough level to multiply once with gradient of activation wrt output
			if gradientLevel < 2 {
				// if not enough bootstrap output of previous layer
				m.Layers2d[l+1].SetBootstrapOutput(true, "backward")
				gradientLevel2D[l+1] = 9
				gradientLevel = gradientLevel2D[l+1]
			}

			activationLossGradientLevel := 0
			if activationLevel > 1 {
				// Calculate level of loss gradient wrt output
				activationLossGradientLevel = int(math.Min(float64(gradientLevel), float64(activationLevel)) - 1)
			} else {
				// Bootstrap activation gradient wrt output before computing loss gradient wrt output if not enough level is reached
				m.Layers2d[l].SetBootstrapActivation(true, "backward")
				activationLossGradientLevel = int(math.Min(float64(gradientLevel), 9.0) - 1)
			}

			// Calculate loss gradient wrt input
			gradientLevel2D[l] = int(math.Min(float64(inputLevel), float64(activationLossGradientLevel)) - float64(m.Layers2d[l].GetBackwardLevelConsumption()))

			// Check if loss gradient wrt input is less than one
			if gradientLevel2D[l] < 1 && activationLossGradientLevel < inputLevel {
				// Bootstrap loss gradient wrt output if it is responsible for making the level of loss wrt input less than 1
				m.Layers2d[l].SetBootstrapActivation(true, "backward")
				activationLossGradientLevel = 9

				// recalculate level of loss gradient wrt input
				gradientLevel2D[l] = int(math.Min(float64(inputLevel), float64(activationLossGradientLevel)) - float64(m.Layers2d[l].GetBackwardLevelConsumption()))
			}

		} else if m.Layers2d[l].IsTrainable() {

			// calculate level of loss gradient wrt input
			gradientLevel2D[l] = int(math.Min(float64(inputLevel), float64(gradientLevel)) - float64(m.Layers2d[l].GetBackwardLevelConsumption()))

			// Bootstrap loss gradient wrt input of next layer if level is not enough
			if gradientLevel2D[l] < 1 {
				m.Layers2d[l-1].SetBootstrapOutput(true, "backward")
				gradientLevel2D[l-1] = 9
				gradientLevel2D[l] = int(math.Min(float64(inputLevel), float64(gradientLevel2D[l-1])) - float64(m.Layers2d[l].GetBackwardLevelConsumption()))
			}

		} else {

			// calculate level of loss gradient wrt input
			gradientLevel2D[l] = gradientLevel - m.Layers2d[l].GetBackwardLevelConsumption()

			// Bootstrap loss gradient wrt input of next layer if level is not enough
			if gradientLevel2D[l] < 1 {
				m.Layers2d[l-1].SetBootstrapOutput(true, "backward")
				gradientLevel2D[l-1] = 9
				gradientLevel2D[l] = int(math.Min(float64(inputLevel), float64(gradientLevel2D[l-1])) - float64(m.Layers2d[l].GetBackwardLevelConsumption()))
			}

		}

	}

	return gradientLevel1D, gradientLevel2D

}
