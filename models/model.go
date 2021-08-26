package models

import (
	"fmt"

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
	utils          	utility.Utils
	Layers1d		[]layers.Layer1D
	Layers2d        []layers.Layer2D
	Flatten 		layers.Flatten2D
	Loss           	losses.Loss
}

func NewModel(utils utility.Utils, layer1d []layers.Layer1D, layer2d []layers.Layer2D, loss losses.Loss) Model {

	var flatten layers.Flatten2D
	if len(layer2d) != 0{
		flatten = layers.NewFlatten(layer2d[len(layer2d) -1].GetOutputSize())
	}

	return Model{utils, layer1d, layer2d, flatten, loss}

}

func (m Model) Forward(input2D [][][]*ckks.Ciphertext, input1D []*ckks.Ciphertext) ([]layers.Output2d, []layers.Output1d) {

	var output2D []layers.Output2d
	var output1D []layers.Output1d

	if len(m.Layers2d) != 0 {

		if input2D != nil {
			panic(fmt.Sprintf("Input2d is not given. Expect input with size %d", m.Layers2d[0].GetOutputSize()))
		}

		// Initialize slice for storing output
		output2D = make([]layers.Output2d, len(m.Layers2d) + 1)

		// Insert output of input layer
		output2D[0] = layers.Output2d{Output: input2D}

		prevLayerHasActivation := false

		// Loop through each layer and calculate output
		for layer := 0; layer < len(m.Layers2d); layer++{

			// input of this layer is output[i]
			// output of this layer i is stored at output[i+1]
			
			var prevOut [][][]*ckks.Ciphertext

			if prevLayerHasActivation {
				prevOut = output2D[layer].ActivationOutput
			} else {
				prevOut = output2D[layer].Output
			}

			output2D[layer + 1] = m.Layers2d[layer].Forward(prevOut)

			prevLayerHasActivation = m.Layers2d[layer].HasActivation()

		}

	}

	// initialize slice for storing 1D output
	output1D = make([]layers.Output1d, len(m.Layers1d) + 1)

	if output2D != nil{

		prevOut := output2D[len(output2D) - 1].Output

		// Check if last 2D layer has activation
		if m.Layers2d[len(m.Layers2d) - 1].HasActivation() {
			prevOut = output2D[len(output2D) - 1].ActivationOutput
		}

		// Insert flattened 2D output in as input array
		output1D[0] = m.Flatten.Forward(prevOut)

	} else {
		// Insert output of input layer
		output1D[0] = layers.Output1d{Output: input1D}
	}

	if len(m.Layers1d) != 0 {

		prevLayerHasActivation := false

		for layer := 0; layer <= len(m.Layers1d); layer++{
			
			prevOut := output1D[layer].Output

			if prevLayerHasActivation {
				prevOut = output1D[layer].ActivationOutput
			}

			output1D[layer+1] = m.Layers1d[layer].Forward(prevOut)

			prevLayerHasActivation = m.Layers1d[layer].HasActivation()

		}

	}

	return output2D, output1D

}

func (m Model) Backward(output2D []layers.Output2d, output1D []layers.Output1d, y []*ckks.Ciphertext) ([]layers.Gradient2d, []layers.Gradient1d) {

	gradient1D := make([]layers.Gradient1d, len(m.Layers1d) + 1)
	gradient2D := make([]layers.Gradient2d, len(m.Layers2d) + 1)

	// Calculate loss gradient
	var finalOutput []*ckks.Ciphertext

	// Get output of last layer
	if len(m.Layers1d) != 0{
		if m.Layers1d[len(m.Layers1d) - 1].HasActivation(){
			finalOutput = output1D[len(m.Layers1d)].ActivationOutput
		} else {
			finalOutput = output1D[len(m.Layers1d)].Output
		}
	}else{
		finalOutput = output1D[len(output1D)-1].Output
	}

	gradient1D[len(gradient1D) - 1] = layers.Gradient1d{InputGradient: m.Loss.Backward(finalOutput, y, len(y))}

	if len(m.Layers1d) != 0 {

		for layer := len(m.Layers1d) - 1; layer >= 0; layer--{

			// input of layer index i is output[i]
			// output of layer i is output[i+1]

			// Get next layer's input gradient
			nextLayerInputGrad := gradient1D[layer + 1].InputGradient

			// Get layer's input
			layerInput := output1D[layer].Output
			if layer != 0{
				if m.Layers1d[layer - 1].HasActivation(){
					layerInput = output1D[layer].ActivationOutput
				}
			}
			
			// Get layer's output
			layerOutput := output1D[layer+1].Output
			if m.Layers1d[layer].HasActivation(){
				layerOutput = output1D[layer+1].ActivationOutput
			}

			gradient1D[layer] = m.Layers1d[layer].Backward(layerInput, layerOutput, nextLayerInputGrad, (layer != 0 || len(m.Layers2d) != 0))

		}

	}

	if len(m.Layers2d) != 0 {

		// Backward flatten layer
		gradient2D[len(gradient2D) - 1] = m.Flatten.Backward(gradient1D[0].InputGradient)

		for layer := len(m.Layers2d) - 1; layer >= 0; layer--{

			// input of layer index i is output[i]
			// output of layer i is output[i+1]

			// Get next layer's input gradient
			nextLayerInputGrad := gradient2D[layer + 1].InputGradient

			// Get layer's input
			layerInput := output2D[layer].Output
			if layer != 0{
				if m.Layers2d[layer - 1].HasActivation(){
					layerInput = output2D[layer].ActivationOutput
				}
			}
			
			// Get layer's output
			layerOutput := output2D[layer+1].Output
			if m.Layers2d[layer].HasActivation(){
				layerOutput = output2D[layer+1].ActivationOutput
			}

			gradient2D[layer] = m.Layers2d[layer].Backward(layerInput, layerOutput, nextLayerInputGrad, (layer != 0 || len(m.Layers2d) != 0))

		}

	}

	return gradient2D, gradient1D

}

func (m *Model) UpdateGradient(gradients1d []layers.Gradient1d, gradients2d []layers.Gradient2d, lr float64) {

	for layer := range m.Layers2d{
		if m.Layers2d[layer].IsTrainable(){
			m.Layers2d[layer].UpdateGradient(gradients2d[layer], lr)
		}
	}

	for layer := range m.Layers1d{
		if m.Layers1d[layer].IsTrainable(){
			m.Layers1d[layer].UpdateGradient(gradients1d[layer], lr)
		}
	}

}

func (m *Model) Train2D(dataLoader dataset.Loader, learningRate float64, batchSize int, epoch int) {

	for i := 0; i < int(dataLoader.GetLength() / batchSize); i++{

		x, y := dataLoader.Load2D(i * batchSize, batchSize)
		
		outputs2D, outputs1D := m.Forward(x, []*ckks.Ciphertext{})
		gradients2D, gradients1D := m.Backward(outputs2D, outputs1D, y)
		m.UpdateGradient(gradients1D, gradients2D, learningRate)

	}

}
