package utility

import (
	"fmt"

	"github.com/ldsec/lattigo/v2/ckks"
)

// This files houses inter-ciphertext operations

func (u Utils) InterDotProduct(a []*ckks.Ciphertext, b []*ckks.Ciphertext, rescale bool, bootstrap bool, concurrent bool) *ckks.Ciphertext {

	if len(a) != len(b) {
		panic("Unequal length")
	}

	sum := u.Encrypt(u.GenerateFilledArraySize(0, len(a)))

	if concurrent {
		ans := make([]ckks.Ciphertext, len(a))

		channels := make([]chan ckks.Ciphertext, len(a))

		for i := range a {

			fmt.Printf("starting element number %d \n", i)
			channels[i] = make(chan ckks.Ciphertext)

			go u.MultiplyConcurrent(*a[i], *b[i], true, channels[i])

		}

		for c := range channels {
			ans[c] = <- channels[c]
		}

		for i := range ans {
			sum = u.AddNew(sum, ans[i])
		}
		
	} else {

		for i := range a {

			prod := u.MultiplyNew(*a[i], *b[i], rescale, bootstrap)

			if i == 0 {
				sum = prod
			} else {
				u.Add(sum, prod, &sum)
			}

		}
	}

	return &sum

}

func (u Utils) InterOuter(a []ckks.Ciphertext, b []ckks.Ciphertext, concurrent bool) [][]ckks.Ciphertext {

	output := make([][]ckks.Ciphertext, len(a))

	if concurrent {
		channels := make([][]chan ckks.Ciphertext, len(a))
		for i := range a {
			channels[i] = make([]chan ckks.Ciphertext, len(b))
			for j := range b {
				fmt.Printf("i = %d, and j = %d", i, j)
				output[i][j] = u.Encrypt(u.GenerateFilledArraySize(0, len(a)))
				go u.MultiplyConcurrent(a[i], b[i], true, channels[i][j])
				for c := range channels[i] {
					// output is currently data
					output[i][c] = <-channels[i][c]
				}
			}
		}

	} else {
		for i := range a {

			output[i] = make([]ckks.Ciphertext, len(b))

			for j := range b {

				product := u.MultiplyNew(a[i], b[j], true, false)
				output[i][j] = product

			}

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
