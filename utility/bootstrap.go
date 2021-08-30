package utility

import "github.com/ldsec/lattigo/v2/ckks"

func (u Utils) BootstrapIfNecessary(ct *ckks.Ciphertext) bool {

	if ct.Level() == 1 {

		u.log.Log("Bootstrapping")
		*ct = *u.Bootstrapper.Bootstrapp(ct)
		return true

	}

	return false

}

func (u Utils) BootstrapInPlace(ct *ckks.Ciphertext) {

	u.log.Log("Bootstrapping")
	*ct = *u.Bootstrapper.Bootstrapp(ct)

}

func (u Utils) Bootstrap1dInPlace(ct []*ckks.Ciphertext) {

	for i := range ct{
		*ct[i] = *u.Bootstrapper.Bootstrapp(ct[i])
	}

}

func (u Utils) Bootstrap3dInPlace(ct [][][]*ckks.Ciphertext) {

	for r := range ct{
		for c := range ct[r]{
			for d := range ct[r][c]{
				*ct[r][c][d] = *u.Bootstrapper.Bootstrapp(ct[r][c][d])
			}
		}
	}

}