package ml

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

type Loss interface {
	Forward(pred ckks.Ciphertext, y ckks.Ciphertext, predLength int) ckks.Ciphertext
	Backward(pred ckks.Ciphertext, y ckks.Ciphertext, predLength int) ckks.Ciphertext
}

type CrossEntropy struct {
	utils utility.Utils
}

func (c CrossEntropy) Forward(pred ckks.Ciphertext, y ckks.Ciphertext, predLength int) ckks.Ciphertext {
	return pred
}

func (c CrossEntropy) Backward(pred ckks.Ciphertext, y ckks.Ciphertext, predLength int) ckks.Ciphertext {
	return c.utils.SubNew(pred, y)
}
