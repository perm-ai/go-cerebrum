package losses

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

type MSE struct {
	U utility.Utils
}

func (m MSE) Forward(pred []*ckks.Ciphertext, y []*ckks.Ciphertext, predLength int) []*ckks.Ciphertext {
	result := make([]*ckks.Ciphertext, len(pred))
	error := make([]*ckks.Ciphertext, len(pred))
	for i := range pred{
		error[i] = m.U.SubNew(pred[i], y[i])
	}
	error_squared := make([]*ckks.Ciphertext, len(pred))
	for i := range pred{
		error_squared[i] = m.U.MultiplyNew(error[i], error[i], true, false)
	}

	// sum := make([]*ckks.Ciphertext, len(pred))
	for i := range pred{
		m.U.SumElementsInPlace(error_squared[i])
	}

	mean := m.U.EncodePlaintextFromArray(m.U.GenerateFilledArray(float64(1/predLength)))

	for i := range pred{
		result[i] = m.U.MultiplyPlainNew(error_squared[i], mean, true, false)
	}
	
	return result
}

func (m MSE) Backward() {

}
