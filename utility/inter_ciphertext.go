package utility

import "github.com/ldsec/lattigo/v2/ckks"

// This files houses inter-ciphertext operations

func (u Utils) InterDotProduct(a []*ckks.Ciphertext, b []*ckks.Ciphertext, rescale bool, bootstrap bool) *ckks.Ciphertext{

	if len(a) != len(b){
		panic("Unequal length")
	}

	var sum ckks.Ciphertext

	for i := range a{

		prod := u.MultiplyNew(*a[i], *b[i], rescale, bootstrap)

		if i == 0{
			sum = prod
		} else {
			u.Add(sum, prod, &sum)
		}

	}

	return &sum

}

func (u Utils) InterOuter(a []*ckks.Ciphertext, b []*ckks.Ciphertext) [][]*ckks.Ciphertext {

	output := make([][]*ckks.Ciphertext, len(a))

	for i := range a {

		output[i] = make([]*ckks.Ciphertext, len(b))

		for j := range b {

			product := u.MultiplyNew(*a[i], *b[j], true, false)
			output[i][j] = &product

		}

	}

	return output

}

func (u Utils) InterTranspose(a [][]*ckks.Ciphertext) [][]*ckks.Ciphertext {

	transposed := make([][]*ckks.Ciphertext, len(a[0]))

	for row := range transposed {

		transposed[row] = make([]*ckks.Ciphertext, len(a))

		for col := range transposed {
			transposed[row][col] = a[col][row]
		}

	}

	return transposed

}
