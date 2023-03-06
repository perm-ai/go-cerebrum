package utility

import (
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/tuneinsight/lattigo/v4/ckks/bootstrapping"
)

func (u Utils) BootstrapIfNecessary(ct *rlwe.Ciphertext) bool {

	if ct.Level() == 1 {

		*ct = *u.Bootstrapper.Bootstrap(ct)
		return true

	}

	return false

}

func (u Utils) BootstrapInPlace(ct *rlwe.Ciphertext) {

	*ct = *u.Bootstrapper.Bootstrap(ct)

}

func (u Utils) Bootstrap1dInPlace(ct []*rlwe.Ciphertext, concurrent bool) {

	if concurrent{

		channels := make([]chan rlwe.Ciphertext, len(ct))

		for i := range ct{

			channels[i] = make(chan rlwe.Ciphertext)
			newBtp := u.Bootstrapper.ShallowCopy()

			go bootstrapGoRoutine(ct[i], *newBtp, channels[i])

		}

		for c := range channels{
			*ct[c] = <- channels[c]
		}

	} else {
		for i := range ct{
			*ct[i] = *u.Bootstrapper.Bootstrap(ct[i])
		}
	}

}

func bootstrapGoRoutine (ciphertext *rlwe.Ciphertext, btp bootstrapping.Bootstrapper, c chan rlwe.Ciphertext){

	c <- *btp.Bootstrap(ciphertext)

}

func (u Utils) Bootstrap3dInPlace(ct [][][]*rlwe.Ciphertext) {

	for r := range ct{
		for c := range ct[r]{
			for d := range ct[r][c]{
				*ct[r][c][d] = *u.Bootstrapper.Bootstrap(ct[r][c][d])
			}
		}
	}

}
