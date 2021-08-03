package utility

import "github.com/ldsec/lattigo/v2/ckks"

// Approximate square root function. Recommended D value is 6 which works with numbers between [0, 2].
func (u Utils) Sqrt(x *ckks.Ciphertext, d int, size int) *ckks.Ciphertext {

	ones := ckks.NewPlaintext(u.Params, x.Level(), x.Scale)
	u.Encoder.EncodeNTT(ones, u.Float64ToComplex128(u.GenerateFilledArraySize(1, size)), u.Params.LogSlots())
	threes := u.Encoder.EncodeNTTNew(u.Float64ToComplex128(u.GenerateFilledArraySize(3, size)), u.Params.LogSlots())

	a := x.CopyNew()
	b := u.SubPlainNew(*x, *ones)

	for i := 0; i < d; i++ {

		// Calculate A[i+1]
		aMultiplier := u.MultiplyConstNew(&b, -0.5, true, false)
		u.AddPlain(&aMultiplier, ones, &aMultiplier)
		u.Multiply(*a, aMultiplier, a, true, false)

		if i < (d - 1) {
			// Calculate B[i+1]
			bMultiplier := u.SubPlainNew(b, *threes)
			u.MultiplyConst(&bMultiplier, 0.25, &bMultiplier, true, false)

			bSquared := u.MultiplyNew(b, b, true, false)
			u.Multiply(bSquared, bMultiplier, &b, true, false)
		}

	}

	return a

}
