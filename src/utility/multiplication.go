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

func (u Utils) MultiplyNew(a ckks.Ciphertext, b ckks.Ciphertext) ckks.Ciphertext {

	u.SwitchToSameModCoeff(&a, &b)
	result := u.Evaluator.MulRelinNew(a, b, &u.RelinKey)
	u.BootstrapIfNecessary(result)

	return *result

}

func (u Utils) MultiplyRescale(a ckks.Ciphertext, b ckks.Ciphertext, destination *ckks.Ciphertext) {

	u.Multiply(a, b, destination)
	u.BootstrapIfNecessary(destination)
	u.Evaluator.Rescale(destination, math.Pow(2.0, 40.0), destination)

}

func (u Utils) MultiplyRescaleNew(a *ckks.Ciphertext, b *ckks.Ciphertext) ckks.Ciphertext {

	result := u.MultiplyNew(*a, *b)
	u.BootstrapIfNecessary(&result)
	u.Evaluator.Rescale(&result, math.Pow(2.0, 40.0), &result)

	return result

}

func (u Utils) MultiplyPlainRescale(a *ckks.Ciphertext, b *ckks.Plaintext, destination *ckks.Ciphertext){

	u.ReEncodeAsNTT(b);
	u.Evaluator.MulRelin(a, b, &u.RelinKey, destination)
	u.BootstrapIfNecessary(destination)
	u.Evaluator.Rescale(destination, math.Pow(2.0, 40.0), destination)

}

func (u Utils) MultiplyPlainRescaleNew(a *ckks.Ciphertext, b *ckks.Plaintext) ckks.Ciphertext {

	u.ReEncodeAsNTT(b);
	result := u.Evaluator.MulRelinNew(a, b, &u.RelinKey)
	u.BootstrapIfNecessary(result)
	u.Evaluator.Rescale(result, math.Pow(2.0, 40.0), result)

	return *result

}

func (u Utils) MultiplyConstRescale(a *ckks.Ciphertext, b []float64, destination *ckks.Ciphertext){

	cmplx := u.Float64ToComplex128(b)
	encoded := u.Encoder.EncodeNTTAtLvlNew(u.Params.MaxLevel(), cmplx, u.Params.LogSlots())
	u.Evaluator.MulRelin(a, encoded, &u.RelinKey, destination)
	u.BootstrapIfNecessary(destination)
	u.Evaluator.Rescale(destination, math.Pow(2.0, 40.0), destination)

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