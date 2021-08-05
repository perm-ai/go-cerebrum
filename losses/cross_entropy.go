package losses

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

type CrossEntropy struct {
	U utility.Utils
}

func (c CrossEntropy) Forward(pred ckks.Ciphertext, y ckks.Ciphertext, predLength int) ckks.Ciphertext {
	return pred
}

func (c CrossEntropy) Backward(pred ckks.Ciphertext, y ckks.Ciphertext, predLength int) ckks.Ciphertext {
	return c.U.SubNew(pred, y)
}
