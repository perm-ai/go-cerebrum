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
