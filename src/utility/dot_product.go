package utility

import "github.com/ldsec/lattigo/v2/ckks"

// func (u Utils) rotateAndAdd(ct *ckks.Ciphertext, size float64) ckks.Ciphertext {

// 	midpoint := size / 2

// 	rotated := u.Evaluator.RotateNew(ct, uint64(midpoint))
// 	u.Add(*ct, *rotated, ct)

// 	if midpoint == 1 {
// 		return *ct
// 	} else {
// 		return u.rotateAndAdd(ct, midpoint)
// 	}

// }

func (u Utils) SumElementsInPlace(ct *ckks.Ciphertext) {

	// u.rotateAndAdd(ct, float64(u.Params.Slots()))
	u.Evaluator.InnerSum(ct, 1, u.Params.Slots(), ct)

}

func (u Utils) SumElementsNew(ct ckks.Ciphertext) ckks.Ciphertext {

	destination := ckks.NewCiphertext(u.Params, 1, u.Params.MaxLevel(), u.Params.Scale())
	u.Evaluator.InnerSum(&ct, 1, u.Params.Slots(), destination)

	return *destination

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