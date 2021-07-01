package utility

import (
	// "math"
	"github.com/ldsec/lattigo/v2/ckks"
)

func (u Utils) Transpose(ciphertexts []*ckks.Ciphertext, column int) []*ckks.Ciphertext {

	// This function will swap the row and column of each number in a ciphertext array where an element
	// of a row is in the same ciphertext and the next column is another ciphertext
	// [ E(x11, x12, x13, ..., x1n), 	=>	[ E(x11, x21),
	//   E(x21, x22, x23, ..., x2n), ]		  E(x12, x22), ..., E(x1n, x2n) ]

	row := len(ciphertexts)
	rotated := make([]*ckks.Ciphertext, row)

	// Rotate each ciphertext back by row spot (row 0 rotate 0, row 1 rotate -1)
	// Eg. (N = 8, n = 4)
	// [ E(x11, x12, x13, x14),
	//   E(x24, x21, x22, x23),
	//	 E(x33, x34, x31, x32),
	//	 E(x42, x43, x44, x41),]
	for i := 0; i < row; i++ {
		rotated[i] = u.Evaluator.RotateNew(ciphertexts[i], uint64(int(u.Params.Slots()) - 1 - i), &u.GaloisKey)
	}

	transposed := make([]*ckks.Ciphertext, column)

	for c := 0; c < column; c++ {

		newRow := ckks.NewCiphertext(&u.Params, 1, rotated[0].Level(), rotated[0].Scale())

		// Zero out non-target slot and add
		// [ E(x11,  0 ,  0 ,  0 ),
		//   E( 0 , x21,  0 ,  0 ),		=>	E(x11, x21, x31, x41)
		//	 E( 0 ,  0 , x31,  0 ),
		//	 E( 0 ,  0 ,  0 , x41),]

		for r := 0; r < row; r++ {
			
			if r == 0{
				u.MultiplyPlain(rotated[r], &u.Filters[c], newRow, true, false)
			} else {
				product := u.MultiplyPlainNew(rotated[r], &u.Filters[r + c], true, false)
				u.Add(product, *newRow, newRow)
			}

		}

		// Rotate ciphertext to align back to original position
		// Eg. E(x24, x34, x44, x14) => E(x14, x24, x34, x44)
		transposed[c] = u.Evaluator.RotateNew(newRow, uint64(c), &u.GaloisKey)

	}

	return transposed

}

func (u Utils) Outer(a *ckks.Ciphertext, b *ckks.Ciphertext, aSize int, bSize int) []ckks.Ciphertext {

	outerProduct := make([]ckks.Ciphertext, aSize)

	for i := 0; i < aSize; i++{

		filtered := u.MultiplyPlainNew(a, &u.Filters[i], true, false)

		// If aSize is more than 2^(logSlot - 2) it would be more or equally efficient to compute sumElement
		if(bSize > int(u.Params.Slots()) / 4) {
			u.SumElementsInPlace(&filtered)
		} else {

			if(i != 0) {
				// Rotate data of interest to slot 0
				u.Evaluator.Rotate(&filtered, uint64(i), &u.GaloisKey, &filtered)
			}

			for j := 1; j < bSize; j *= 2{
				// Rotate and add to double the amount of data each iteration
				rotated := u.Evaluator.RotateNew(&filtered, uint64(int(u.Params.Slots()) - 1 - j), &u.GaloisKey)
				u.Add(filtered, *rotated, &filtered)
			}

		}

		// Calculate product
		outerProduct[i] = u.MultiplyNew(filtered, *b, true, false)

	}

	return outerProduct

}