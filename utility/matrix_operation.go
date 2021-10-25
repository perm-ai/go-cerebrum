package utility

import (
	"github.com/ldsec/lattigo/v2/ckks"
)

func (u Utils) Transpose(ciphertexts []ckks.Ciphertext, column int) []ckks.Ciphertext {

	// This function will swap the row and column of each number in a ciphertext array where an element
	// of a row is in the same ciphertext and the next column is another ciphertext
	// [ E(x11, x12, x13, ..., x1n), 	=>	[ E(x11, x21),
	//   E(x21, x22, x23, ..., x2n), ]		  E(x12, x22), ..., E(x1n, x2n) ]

	row := len(ciphertexts)
	rotated := make([]ckks.Ciphertext, row)

	// Rotate each ciphertext back by row spot (row 0 rotate 0, row 1 rotate -1)
	// Eg. (N = 8, n = 4)
	// [ E(x11, x12, x13, x14),
	//   E(x24, x21, x22, x23),
	//	 E(x33, x34, x31, x32),
	//	 E(x42, x43, x44, x41),]
	for i := 0; i < row; i++ {
		if i == 0 {
			rotated[i] = ciphertexts[i]
		} else {
			rotated[i] = u.RotateNew(&ciphertexts[i], -i)
		}
	}

	transposed := make([]ckks.Ciphertext, column)

	for c := 0; c < column; c++ {

		newRow := ckks.NewCiphertext(u.Params, 1, ciphertexts[0].Level(), ciphertexts[0].Scale)

		// Zero out non-target slot and add
		// [ E(x11,  0 ,  0 ,  0 ),
		//   E( 0 , x21,  0 ,  0 ),		=>	E(x11, x21, x31, x41)
		//	 E( 0 ,  0 , x31,  0 ),
		//	 E( 0 ,  0 ,  0 , x41),]

		for r := 0; r < row; r++ {

			if r == 0 {
				u.MultiplyPlain(&rotated[r], &u.Filters[c], newRow, true, false)
			} else {
				product := u.MultiplyPlainNew(&rotated[r], &u.Filters[r+c], true, false)
				u.Add(product, newRow, newRow)
			}

		}

		// Rotate ciphertext to align back to original position
		// Eg. E(x24, x34, x44, x14) => E(x14, x24, x34, x44)
		transposed[c] = u.RotateNew(newRow, c)

	}

	return transposed

}

func (u Utils) Outer(a *ckks.Ciphertext, b *ckks.Ciphertext, aSize int, bSize int, filterBy float64) []*ckks.Ciphertext {

	// Need to cover rotation in range [0, aSize)
	pow2rotationEvaluator := u.Get2PowRotationEvaluator()

	outerProduct := make([]*ckks.Ciphertext, aSize)

	for i := 0; i < aSize; i++ {

		var filtered *ckks.Ciphertext

		if(filterBy == 0 || filterBy == 1){

			filtered = u.MultiplyPlainNew(a, &u.Filters[i], true, false)
			
		} else {
			
			// Generate filter with given filter scale
			filterCmplx := make([]complex128, u.Params.Slots())
			filterCmplx[i] = complex(filterBy, 0)
			filter := u.Encoder.EncodeNTTNew(filterCmplx, u.Params.LogSlots())

			filtered = u.MultiplyPlainNew(a, filter, true, false)
			
		}
		

		// If aSize is more than 2^(logSlot - 2) it would be more or equally efficient to compute sumElement
		if bSize > int(u.Params.Slots())/4 {
			u.SumElementsInPlace(filtered)
		} else {

			if i != 0 {
				// Rotate data of interest to slot 0
				u.Rotate(filtered, i)
			}

			for j := 1; j < bSize; j *= 2 {
				// Rotate and add to double the amount of data each iteration
				rotated := pow2rotationEvaluator.RotateNew(filtered, -j)
				u.Add(filtered, rotated, filtered)
			}

		}

		// Calculate product
		outerProduct[i] = u.MultiplyNew(filtered, b, true, false)

	}

	return outerProduct

}

func (u Utils) PackVector(ciphertexts []ckks.Ciphertext) ckks.Ciphertext {

	result := ckks.NewCiphertext(u.Params, 1, u.Params.MaxLevel(), u.Params.Scale())

	for i := range ciphertexts{

		if i == 0 {
			u.MultiplyPlain(&ciphertexts[i], &u.Filters[i], result, true, false)
		} else {
			filtered := u.MultiplyPlainNew(&ciphertexts[i], &u.Filters[i], true, false)
			u.Add(filtered, result, result)
		}

	}

	return *result

}