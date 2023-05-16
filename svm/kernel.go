package svm

import (
	"errors"

	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/utility"
)


type Kernel interface {
	Calculate(xi []*rlwe.Ciphertext, xj []*rlwe.Ciphertext) (*rlwe.Ciphertext, error)
	Type() string
}

type Linear struct {
	u			utility.Utils
}

func (l Linear) Calculate(xi []*rlwe.Ciphertext, xj []*rlwe.Ciphertext) (*rlwe.Ciphertext, error) {

	if len(xi) != len(xj){
		return nil, errors.New("INVALID INPUT: Lenght of xi and xj aren't equal")
	}

	var result *rlwe.Ciphertext

	for i := range xi {
		if i == 0{
			result = l.u.MultiplyNew(xi[i], xj[i], true, false)
		} else {
			product := l.u.MultiplyNew(xi[i], xj[i], true, false)
			l.u.Add(result, product, result)
		}
	}

	return result, nil

}

func (l Linear) Type() string {
	return "linear"
}

type RBF struct {
	u 			utility.Utils
	Gamma 		float64
	encGamma 	rlwe.Plaintext
}

// Generate new RBF struct where gamma = 1 / (2(σ^2))
func NewRBF(u utility.Utils, gamma float64) RBF {

	encGamma := u.Encoder.EncodeNew(u.Float64ToComplex128(u.GenerateFilledArray((-1) * gamma)), u.Params.MaxLevel(), u.Params.DefaultScale(), u.Params.LogSlots())

	return RBF{u, gamma, *encGamma}

}

// Calculate the out put from the RBF kernel. Each ciphertext in xi and xj array should represents a feature of a data
func (r RBF) Calculate(xi []*rlwe.Ciphertext, xj []*rlwe.Ciphertext) (*rlwe.Ciphertext, error) {

	if len(xi) != len(xj){
		return nil, errors.New("INVALID INPUT: Lenght of xi and xj aren't equal")
	}

	var result *rlwe.Ciphertext

	for i := range xi {
		sub := r.u.SubNew(xi[i], xj[i])
		if i == 0{
			result = r.u.MultiplyNew(sub, sub, true, false)
		} else {
			squared := r.u.MultiplyNew(sub, sub, true, false)
			r.u.Add(result, squared, result)
		}
	}

	r.u.MultiplyPlain(result, &r.encGamma, result, true, false)

	return result, nil
	
}

func (r RBF) Type() string {
	return "rbf"
}