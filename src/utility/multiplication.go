package utility

import (
	"math"

	"github.com/ldsec/lattigo/v2/ckks"
)

var RESCALE_THRESHOLD = math.Pow(2.0, 40.0)

func (u Utils) Multiply(a ckks.Ciphertext, b ckks.Ciphertext, destination *ckks.Ciphertext, rescale bool, bootstrap bool) {

	u.SwitchToSameModCoeff(&a, &b)
	u.Evaluator.MulRelin(a, b, &u.RelinKey, destination)
	
	if bootstrap {
		u.BootstrapIfNecessary(destination)
	}

	if rescale {
		u.Evaluator.Rescale(destination, RESCALE_THRESHOLD, destination)
	}

}

func (u Utils) MultiplyNew(a ckks.Ciphertext, b ckks.Ciphertext, rescale bool, bootstrap bool) ckks.Ciphertext {

	u.SwitchToSameModCoeff(&a, &b)
	result := u.Evaluator.MulRelinNew(a, b, &u.RelinKey)

	if bootstrap {
		u.BootstrapIfNecessary(result)
	}

	if rescale {
		u.Evaluator.Rescale(result, RESCALE_THRESHOLD, result)
	}

	return *result

}

func (u Utils) MultiplyPlain(a *ckks.Ciphertext, b *ckks.Plaintext, destination *ckks.Ciphertext, rescale bool, bootstrap bool){

	u.ReEncodeAsNTT(b);
	u.Evaluator.MulRelin(a, b, &u.RelinKey, destination)
	
	if bootstrap {
		u.BootstrapIfNecessary(destination)
	}

	if rescale {
		u.Evaluator.Rescale(destination, RESCALE_THRESHOLD, destination)
	}

}

func (u Utils) MultiplyPlainNew(a *ckks.Ciphertext, b *ckks.Plaintext, rescale bool, bootstrap bool) ckks.Ciphertext {

	u.ReEncodeAsNTT(b);
	result := u.Evaluator.MulRelinNew(a, b, &u.RelinKey)

	if bootstrap {
		u.BootstrapIfNecessary(result)
	}

	if rescale {
		u.Evaluator.Rescale(result, RESCALE_THRESHOLD, result)
	}

	return *result

}

func (u Utils) MultiplyConst(a *ckks.Ciphertext, b []float64, destination *ckks.Ciphertext, rescale bool, bootstrap bool){

	cmplx := u.Float64ToComplex128(b)
	encoded := u.Encoder.EncodeNTTAtLvlNew(a.Level(), cmplx, u.Params.LogSlots())
	u.Evaluator.MulRelin(a, encoded, &u.RelinKey, destination)
	
	if bootstrap {
		u.BootstrapIfNecessary(destination)
	}

	if rescale {
		u.Evaluator.Rescale(destination, RESCALE_THRESHOLD, destination)
	}

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

func (u Utils) ReEncodeAsNTT(a *ckks.Plaintext) {

	if !a.IsNTT() {

		// Reencode as ntt
		data := u.Encoder.Decode(a, u.Params.LogSlots())
		u.Encoder.EncodeNTT(a, data, u.Params.LogSlots())

	}

}