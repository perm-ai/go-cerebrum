package utility

import (
	"math"

	"github.com/ldsec/lattigo/v2/ckks"
)

func (u Utils) Multiply(a ckks.Ciphertext, b ckks.Ciphertext, destination *ckks.Ciphertext) {

	u.SwitchToSameModCoeff(&a, &b)
	u.Evaluator.MulRelin(a, b, &u.RelinKey, destination)
	u.BootstrapIfNecessary(destination)

}

func (u Utils) MultiplyNew(a *ckks.Ciphertext, b *ckks.Ciphertext) ckks.Ciphertext {

	result := ckks.NewCiphertext(&u.Params, 1, u.Params.MaxLevel(), a.Scale())
	u.Multiply(*a, *b, result)
	u.BootstrapIfNecessary(result)

	return *result

}

func (u Utils) MultiplyRescale(a ckks.Ciphertext, b ckks.Ciphertext, destination *ckks.Ciphertext) {

	u.Multiply(a, b, destination)
	u.BootstrapIfNecessary(destination)
	u.Evaluator.Rescale(destination, math.Pow(2.0, 40.0), destination)

}

func (u Utils) MultiplyRescaleNew(a *ckks.Ciphertext, b *ckks.Ciphertext) ckks.Ciphertext {

	result := u.MultiplyNew(a, b)
	u.BootstrapIfNecessary(&result)
	u.Evaluator.Rescale(&result, math.Pow(2.0, 40.0), &result)

	return result

}

func (u Utils) SwitchToSameModCoeff(a *ckks.Ciphertext, b *ckks.Ciphertext) {

	if a.Level() != b.Level() {

		var requireSwitch *ckks.Ciphertext
		var constant *ckks.Ciphertext

		if a.Level() > b.Level() {
			requireSwitch = a
			constant = b
		} else {
			requireSwitch = b
			constant = a
		}

		u.Evaluator.DropLevel(requireSwitch, requireSwitch.Level()-constant.Level())

	}

}
