package layers

import "github.com/tuneinsight/lattigo/v4/rlwe"

//=================================================
//		   		2 DIMENTIONAL GRADIENT
//=================================================
type Gradient2d struct {
	BiasGradient   []*rlwe.Ciphertext
	WeightGradient [][][]*rlwe.Ciphertext
	InputGradient  [][][]*rlwe.Ciphertext
}

//=================================================
//		   		2 DIMENTIONAL OUTPUT
//=================================================
type Output2d struct {
	Output           [][][]*rlwe.Ciphertext
	ActivationOutput [][][]*rlwe.Ciphertext
}

//=================================================
//			  1 DIMENTIONAL GRADIENT
//=================================================

type Gradient1d struct {
	BiasGradient   []*rlwe.Ciphertext
	WeightGradient [][]*rlwe.Ciphertext
	InputGradient  []*rlwe.Ciphertext
}

//=================================================
//		   		1 DIMENTIONAL OUTPUT
//=================================================
type Output1d struct {
	Output           []*rlwe.Ciphertext
	ActivationOutput []*rlwe.Ciphertext
}

//=================================================
//			  1 DIMENTIONAL LAYER
//=================================================

type Layer1D interface {
	Forward(input []*rlwe.Ciphertext) Output1d
	Backward(input []*rlwe.Ciphertext, output []*rlwe.Ciphertext, gradient []*rlwe.Ciphertext, hasPrevLayer bool) Gradient1d
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
	Forward(input [][][]*rlwe.Ciphertext) Output2d
	Backward(input [][][]*rlwe.Ciphertext, output [][][]*rlwe.Ciphertext, gradient [][][]*rlwe.Ciphertext, hasPrevLayer bool) Gradient2d
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
