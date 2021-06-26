package utility

import "github.com/ldsec/lattigo/v2/ckks"

func (u Utils) BootstrapIfNecessary(ct *ckks.Ciphertext) {

	if ct.Level() == 1 {

		*ct = *u.Bootstrapper.Bootstrapp(ct)

	}

}
