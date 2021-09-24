package utility

import (
	"github.com/ldsec/lattigo/v2/ckks"
)

// This files houses inter-ciphertext operations

func (u Utils) InterDotProduct(a []*ckks.Ciphertext, b []*ckks.Ciphertext, rescale bool, bootstrap bool, concurrent bool) *ckks.Ciphertext {

	if len(a) != len(b) {
		panic("Unequal length")
	}

	var sum *ckks.Ciphertext

	if concurrent {

		channels := make([]chan *ckks.Ciphertext, len(a))

		for i := range a {

			channels[i] = make(chan *ckks.Ciphertext)

			go u.MultiplyConcurrent(a[i], b[i], true, channels[i])

		}

		for c := range channels {
			if c == 0 {
				sum = <-channels[c]
			} else {
				u.Add(*sum, *<-channels[c], sum)
			}
		}

	} else {

		for i := range a {

			prod := u.MultiplyNew(*a[i], *b[i], rescale, bootstrap)

			if i == 0 {
				sum = &prod
			} else {
				u.Add(*sum, prod, sum)
			}

		}
	}

	return sum

}

func (u Utils) InterOuter(a []*ckks.Ciphertext, b []*ckks.Ciphertext, concurrent bool) [][]*ckks.Ciphertext {

	output := make([][]*ckks.Ciphertext, len(a))

	if concurrent {

		outputChannels := make([]chan []*ckks.Ciphertext, len(a))

		for i := range a {

			outputChannels[i] = make(chan []*ckks.Ciphertext)

			go func(row int, rowChannel chan []*ckks.Ciphertext) {

				colOutput := make([]*ckks.Ciphertext, len(b))
				colChannels := make([]chan *ckks.Ciphertext, len(b))

				for j := range b {

					colChannels[j] = make(chan *ckks.Ciphertext)

					go u.MultiplyConcurrent(a[row], b[j], true, colChannels[j])

				}

				for j := range colChannels {
					colOutput[j] = <-colChannels[j]
				}

				rowChannel <- colOutput

			}(i, outputChannels[i])

		}

		for i := range outputChannels {
			output[i] = <-outputChannels[i]
		}

	} else {
		for i := range a {

			output[i] = make([]*ckks.Ciphertext, len(b))

			for j := range b {

				product := u.MultiplyNew(*a[i], *b[j], true, false)
				output[i][j] = &product

			}

		}
	}

	return output

}

func (u Utils) InterTranspose(a [][]*ckks.Ciphertext) [][]*ckks.Ciphertext {

	transposed := make([][]*ckks.Ciphertext, len(a[0]))

	for row := range transposed {

		transposed[row] = make([]*ckks.Ciphertext, len(a))

		for col := range transposed[row] {
			transposed[row][col] = a[col][row]
		}

	}

	return transposed

}
