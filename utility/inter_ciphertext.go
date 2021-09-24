package utility

import (
	"sync"

	"github.com/ldsec/lattigo/v2/ckks"
)

// This files houses inter-ciphertext operations

type SafeSum struct{
	Ct *ckks.Ciphertext
	mu sync.Mutex
}

func (s *SafeSum) Add(ct *ckks.Ciphertext, utils Utils){
	s.mu.Lock()
	if s.Ct == nil {
		s.Ct = ct
	} else {
		utils.Add(s.Ct, ct, s.Ct)
	}
	s.mu.Unlock()
}

func (u Utils) InterDotProduct(a []*ckks.Ciphertext, b []*ckks.Ciphertext, rescale bool, bootstrap bool, concurrent bool) *ckks.Ciphertext {

	if len(a) != len(b) {
		panic("Unequal length")
	}

	sum := SafeSum{}

	if concurrent {

		var wg sync.WaitGroup

		for i := range a {

			wg.Add(1)

			go func(index int, utils Utils){

				defer wg.Done()

				product := utils.MultiplyNew(a[index], b[index], true, false)
				sum.Add(product, utils)

			}(i, u.CopyWithClonedEval())
			
		}

		wg.Wait()

	} else {

		for i := range a {

			prod := u.MultiplyNew(a[i], b[i], rescale, bootstrap)
			sum.Add(prod, u)

		}
	}

	return sum.Ct

}

func (u Utils) InterOuter(a []*ckks.Ciphertext, b []*ckks.Ciphertext, concurrent bool) [][]*ckks.Ciphertext {

	output := make([][]*ckks.Ciphertext, len(a))

	if concurrent {

		var rowWg sync.WaitGroup

		for i := range a {

			rowWg.Add(1)
			output[i] = make([]*ckks.Ciphertext, len(b))

			go func(row int) {

				defer rowWg.Done()
				var colWg sync.WaitGroup

				for j := range b {

					colWg.Add(1)

					go func(column int, utils Utils){

						defer colWg.Done()
						output[row][column] = utils.MultiplyNew(a[row], b[column], true, false)

					}(j, u.CopyWithClonedEval())

				}

				colWg.Wait()

			}(i)

		}

		rowWg.Wait()

	} else {
		for i := range a {

			output[i] = make([]*ckks.Ciphertext, len(b))

			for j := range b {

				output[i][j] = u.MultiplyNew(a[i], b[j], true, false)

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
