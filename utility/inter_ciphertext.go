package utility

import (
	"sync"

	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/logger"
)

// This files houses inter-ciphertext operations

type SafeSum struct{
	Ct *rlwe.Ciphertext
	mu sync.Mutex
}

func (s *SafeSum) Add(ct *rlwe.Ciphertext, utils Utils){
	s.mu.Lock()
	if s.Ct == nil {
		s.Ct = ct.CopyNew()
	} else {
		utils.Add(s.Ct, ct, s.Ct)
	}
	s.mu.Unlock()
}

func (u Utils) InterDotProduct(a []*rlwe.Ciphertext, b []*rlwe.Ciphertext, rescale bool, concurrent bool, counter *logger.OperationsCounter) *rlwe.Ciphertext {

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

				product := utils.MultiplyNew(a[index], b[index], rescale, false)
				sum.Add(product, utils)
				if counter != nil {
					counter.Increment()
				}

			}(i, u.CopyWithClonedEval())
			
		}

		wg.Wait()

	} else {

		for i := range a {

			prod := u.MultiplyNew(a[i], b[i], rescale, false)
			sum.Add(prod, u)

		}
	}

	return sum.Ct

}

func (u Utils) InterOuter(a []*rlwe.Ciphertext, b []*rlwe.Ciphertext, concurrent bool) [][]*rlwe.Ciphertext {

	output := make([][]*rlwe.Ciphertext, len(a))

	if concurrent {

		var rowWg sync.WaitGroup

		for i := range a {

			rowWg.Add(1)
			output[i] = make([]*rlwe.Ciphertext, len(b))

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

			output[i] = make([]*rlwe.Ciphertext, len(b))

			for j := range b {

				output[i][j] = u.MultiplyNew(a[i], b[j], true, false)

			}

		}
	}

	return output

}

func (u Utils) InterTranspose(a [][]*rlwe.Ciphertext) [][]*rlwe.Ciphertext {

	transposed := make([][]*rlwe.Ciphertext, len(a[0]))

	for row := range transposed {

		transposed[row] = make([]*rlwe.Ciphertext, len(a))

		for col := range transposed[row] {
			transposed[row][col] = a[col][row]
		}

	}

	return transposed

}
