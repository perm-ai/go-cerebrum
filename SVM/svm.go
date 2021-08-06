package svm

import (
	"fmt"
	"math"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/activations"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/utility"
)

type SVM struct {
	u      *utility.Utils
	Weight *ckks.Ciphertext
}

func NewSVM(u utility.Utils) SVM {

	weightPlain := make([]float64, u.Params.Slots())
	encryptedWeight := u.Encrypt(weightPlain)

	return SVM{&u, &encryptedWeight}

}

func (model SVM) Forward(x ckks.Ciphertext) ckks.Ciphertext {

	return model.u.DotProductNew(x, *model.Weight, false)

}

// Backward propagation of SVM to get gradient
// X and Y should be formatted through svm.PackSvmTrainingData
func (model SVM) Backward(x ckks.Ciphertext, y ckks.Ciphertext, featureSize int, ceilPow2 int, dataPoints int, C float64, bound float64, learningRate float64) ckks.Ciphertext {

	// Calculate xw
	distance := model.u.MultiplyNew(x, *model.Weight, true, false)

	// Calculate sum(x2)
	var rotated ckks.Ciphertext

	for rot := ceilPow2 / 2; rot >= 1; rot /= 2 {
		rotated = model.u.RotateNew(&distance, rot)
		model.u.Add(rotated, distance, &distance)
	}

	rotated = model.u.RotateNew(&distance, (-1 * featureSize))
	model.u.Add(rotated, distance, &distance)

	// Calculate sum(xw) * y
	model.u.Multiply(distance, y, &distance, true, false)

	ones := make([][]float64, model.u.Params.Slots())

	for i := 0; i < dataPoints; i++ {
		ones[i] = make([]float64, featureSize)
		for j := range ones[i] {
			ones[i][j] = 1
		}
	}

	// Calculate distance = 1 - (sum(xw) * y)
	spacedOne, _ := PackSvmTrainingData(ones, ones[0], model.u.Params.Slots())
	onePlain := ckks.NewPlaintext(model.u.Params, distance.Level(), distance.Scale)
	model.u.Encoder.EncodeNTT(onePlain, model.u.Float64ToComplex128(spacedOne), model.u.Params.LogSlots())

	model.u.Evaluator.Sub(onePlain, &distance, &distance)

	// Calculate activation
	activation := activations.NewSvmActivation(*model.u)
	activated := activation.Forward(distance, bound)

	// Calculate margin * Y * X
	m := model.u.MultiplyConstNew(&y, C * learningRate, true, false)
	model.u.Multiply(m, x, &m, true, false)

	// Calculate dw = w - (m * activated_distance)
	dw := model.u.MultiplyNew(m, activated, true, false)
	model.u.Sub(*model.Weight, dw, &dw)

	// Calculate sum of weight from all data size
	for rot := model.u.Params.Slots() / 2; rot > ceilPow2; rot /= 2 {
		rotated = model.u.RotateNew(&dw, rot)
		model.u.Add(dw, rotated, &dw)
	}

	// Calculate average dw
	model.u.MultiplyConst(&dw, (1.0 / float64(dataPoints)), &dw, true, false)

	return dw

}

func (model *SVM) UpdateGradient(dw ckks.Ciphertext){

	model.u.Sub(*model.Weight, dw, model.Weight)
	model.u.BootstrapInPlace(model.Weight)

}

func (model *SVM) Train(x ckks.Ciphertext, y ckks.Ciphertext, featureSize int, ceilPow2 int, dataPoints int, C float64, bound float64, learningRate float64, epoch int){

	for i := 0; i < epoch; i++ {
		dw := model.Backward(x, y, featureSize, ceilPow2, dataPoints, C, bound, learningRate)
		model.UpdateGradient(dw)
	}

}

func PackSvmTrainingData(x [][]float64, y []float64, slots int) (packedX []float64, packedY []float64) {

	pow2 := key.GetPow2K(int(math.Log2(float64(slots))))
	dataLen := len(x[0]) + 1
	var ceilPow2 int

	for _, pow2 := range pow2 {

		if pow2 > dataLen {
			ceilPow2 = pow2
			break
		} else {
			ceilPow2 = pow2
		}

	}
	fmt.Println(ceilPow2)

	spacePerData := ceilPow2 * 2
	fitable := float64(slots) / float64(spacePerData)
	packedX = make([]float64, slots)
	packedY = make([]float64, slots)

	for i := 0; float64(i) < fitable && i < len(x); i++ {
		start := spacePerData * i
		for j, n := range x[i] {
			packedX[start+j] = n
			packedY[start+j] = y[i]
		}
		packedX[start+len(x[i])] = 1
		packedY[start+len(x[i])] = y[i]
	}

	return packedX, packedY

}
