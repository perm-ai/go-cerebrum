package layers

import "github.com/ldsec/lattigo/v2/ckks"

//=================================================
//		   		2 DIMENTIONAL GRADIENT
//=================================================
type Gradient2d struct {
	BiasGradient   []*ckks.Ciphertext
	WeightGradient [][][]*ckks.Ciphertext
	InputGradient  [][][]*ckks.Ciphertext
}

//=================================================
//		   		2 DIMENTIONAL OUTPUT
//=================================================
type Output2d struct {
	Output           [][][]*ckks.Ciphertext
	ActivationOutput [][][]*ckks.Ciphertext
}

//=================================================
//			  1 DIMENTIONAL GRADIENT
//=================================================

type Gradient1d struct {
	BiasGradient   []*ckks.Ciphertext
	WeightGradient [][]*ckks.Ciphertext
	InputGradient  []*ckks.Ciphertext
}

//=================================================
//		   		1 DIMENTIONAL OUTPUT
//=================================================
type Output1d struct {
	Output           []*ckks.Ciphertext
	ActivationOutput []*ckks.Ciphertext
}

//=================================================
//			  1 DIMENTIONAL LAYER
//=================================================

type Layer1D interface {
	Forward(input []*ckks.Ciphertext) Output1d
	Backward(input []*ckks.Ciphertext, output []*ckks.Ciphertext, gradient []*ckks.Ciphertext, hasPrevLayer bool) Gradient1d
	UpdateGradient(gradient Gradient1d, lr float64)

	GetOutputSize() int
	IsTrainable() bool
	HasActivation() bool

	SetBootstrapOutput(set bool, direction string) // Direction "forward" or "backward"
	SetBootstrapActivation(set bool, direction string) // Direction "forward" or "backward"
	GetForwardLevelConsumption() int
	GetBackwardLevelConsumption() int
	GetForwardActivationLevelConsumption() int
	GetBackwardActivationLevelConsumption() int
	SetWeightLevel(lvl int)

	ExportWeights(filename string)
}

//=================================================
//			  2 DIMENTIONAL LAYER
//=================================================

type Layer2D interface {
	Forward(input [][][]*ckks.Ciphertext) Output2d
	Backward(input [][][]*ckks.Ciphertext, output [][][]*ckks.Ciphertext, gradient [][][]*ckks.Ciphertext, hasPrevLayer bool) Gradient2d
	UpdateGradient(gradient Gradient2d, lr float64)
	GetOutputSize() []int
	IsTrainable() bool
	HasActivation() bool

	SetBootstrapOutput(set bool, direction string) // Direction "forward" or "backward"
	SetBootstrapActivation(set bool, direction string) // Direction "forward" or "backward"
	GetForwardLevelConsumption() int
	GetBackwardLevelConsumption() int
	GetForwardActivationLevelConsumption() int
	GetBackwardActivationLevelConsumption() int
	SetWeightLevel(lvl int)
}
