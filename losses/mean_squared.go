package losses

import (
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/utility"
)

type MSE struct {
	U utility.Utils
}

func (m MSE) Forward(pred []*rlwe.Ciphertext, y []*rlwe.Ciphertext, predLength int) []*rlwe.Ciphertext {
	result := make([]*rlwe.Ciphertext, len(pred))
	error := make([]*rlwe.Ciphertext, len(pred))
	for i := range pred{
		error[i] = m.U.SubNew(pred[i], y[i])
	}
	error_squared := make([]*rlwe.Ciphertext, len(pred))
	for i := range pred{
		error_squared[i] = m.U.MultiplyNew(error[i], error[i], true, false)
	}

	// sum := make([]*rlwe.Ciphertext, len(pred))
	for i := range pred{
		m.U.SumElementsInPlace(error_squared[i])
	}

	mean := m.U.EncodePlaintextFromArray(m.U.GenerateFilledArray(float64(1/predLength)))

	for i := range pred{
		result[i] = m.U.MultiplyPlainNew(error_squared[i], mean, true, false)
	}
	
	return result
}

func (m MSE) Backward(pred []*rlwe.Ciphertext, y []*rlwe.Ciphertext, predLength int) []*rlwe.Ciphertext {
	result := make([]*rlwe.Ciphertext, len(pred))

	for i := range pred{
		result[i] = m.U.SubNew(pred[i], y[i])
	}


	return result
}
