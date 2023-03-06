package utility

import (
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/tuneinsight/lattigo/v4/ckks"
)

func (u Utils) rotateAndAdd(ct *rlwe.Ciphertext, size float64, evaluator *ckks.Evaluator) *rlwe.Ciphertext {

	midpoint := size / 2

	rotated := (*evaluator).RotateNew(ct, int(midpoint))
	u.Add(ct, rotated, ct)

	if midpoint == 1 {
		return ct
	} else {
		return u.rotateAndAdd(ct, midpoint, evaluator)
	}

}

func (u Utils) SumElementsInPlace(ct *rlwe.Ciphertext) {

	rotationEvaluator := u.Get2PowRotationEvaluator()
	u.rotateAndAdd(ct, float64(u.Params.Slots()), &rotationEvaluator)

}

func (u Utils) SumElementsNew(ct rlwe.Ciphertext) *rlwe.Ciphertext {

	rotationEvaluator := u.Get2PowRotationEvaluator()
	newCt := ct.CopyNew()
	return u.rotateAndAdd(newCt, float64(u.Params.Slots()), &rotationEvaluator)

}

func (u Utils) DotProduct(a *rlwe.Ciphertext, b *rlwe.Ciphertext, destination *rlwe.Ciphertext, bootstrap bool) {

	u.Multiply(a, b, destination, true, bootstrap)
	u.SumElementsInPlace(destination)

}

func (u Utils) DotProductNew(a *rlwe.Ciphertext, b *rlwe.Ciphertext, bootstrap bool) *rlwe.Ciphertext {

	result := u.MultiplyNew(a, b, true, bootstrap)
	u.SumElementsInPlace(result)

	return result

}


