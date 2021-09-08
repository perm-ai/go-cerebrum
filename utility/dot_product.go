package utility

import (
	"github.com/ldsec/lattigo/v2/ckks"
)

func (u Utils) rotateAndAdd(ct *ckks.Ciphertext, size float64, evaluator ckks.Evaluator) ckks.Ciphertext {

	midpoint := size / 2

	rotated := evaluator.RotateNew(ct, int(midpoint))
	u.Add(*ct, *rotated, ct)

	if midpoint == 1 {
		return *ct
	} else {
		return u.rotateAndAdd(ct, midpoint, evaluator)
	}

}

func (u Utils) SumElementsInPlace(ct *ckks.Ciphertext) {

	rotationEvaluator := u.Get2PowRotationEvaluator()
	u.rotateAndAdd(ct, float64(u.Params.Slots()), rotationEvaluator)

}

func (u Utils) SumElementsNew(ct ckks.Ciphertext) ckks.Ciphertext {

	rotationEvaluator := u.Get2PowRotationEvaluator()
	return u.rotateAndAdd(&ct, float64(u.Params.Slots()), rotationEvaluator)

}

func (u Utils) DotProduct(a ckks.Ciphertext, b ckks.Ciphertext, destination *ckks.Ciphertext, bootstrap bool) {

	u.Multiply(a, b, destination, true, bootstrap)
	u.SumElementsInPlace(destination)

}

func (u Utils) DotProductNew(a ckks.Ciphertext, b ckks.Ciphertext, bootstrap bool) ckks.Ciphertext {

	result := u.MultiplyNew(a, b, true, bootstrap)
	u.SumElementsInPlace(&result)

	return result

}


