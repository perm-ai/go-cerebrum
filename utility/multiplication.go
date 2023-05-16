package utility

import (
	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

func (u Utils) Multiply(a *rlwe.Ciphertext, b *rlwe.Ciphertext, destination *rlwe.Ciphertext, rescale bool, bootstrap bool) {

	a, b = u.SwitchToSameModCoeff(a, b)
	u.Evaluator.MulRelin(a, b, destination)

	if bootstrap {
		u.BootstrapIfNecessary(destination)
	}

	if rescale {
		u.Evaluator.Rescale(destination, rlwe.NewScale(u.Scale), destination)
	}

}

func (u Utils) MultiplyNew(a *rlwe.Ciphertext, b *rlwe.Ciphertext, rescale bool, bootstrap bool) *rlwe.Ciphertext {

	a, b = u.SwitchToSameModCoeff(a, b)
	result := u.Evaluator.MulRelinNew(a, b)

	if bootstrap {
		u.BootstrapIfNecessary(result)
	}

	if rescale {
		u.Evaluator.Rescale(result, rlwe.NewScale(u.Scale), result)
	}

	return result

}

func (u Utils) MultiplyPlain(a *rlwe.Ciphertext, b *rlwe.Plaintext, destination *rlwe.Ciphertext, rescale bool, bootstrap bool) {

	u.ReEncodeAsNTT(b)
	u.Evaluator.MulRelin(a, b, destination)

	if bootstrap {
		u.BootstrapIfNecessary(destination)
	}

	if rescale {
		u.Evaluator.Rescale(destination, rlwe.NewScale(u.Scale), destination)
	}

}

func (u Utils) MultiplyPlainNew(a *rlwe.Ciphertext, b *rlwe.Plaintext, rescale bool, bootstrap bool) *rlwe.Ciphertext {

	u.ReEncodeAsNTT(b)
	result := u.Evaluator.MulRelinNew(a, b)

	if bootstrap {
		u.BootstrapIfNecessary(result)
	}

	if rescale {
		u.Evaluator.Rescale(result, rlwe.NewScale(u.Scale), result)
	}

	return result

}

func (u Utils) MultiplyConstArray(a *rlwe.Ciphertext, b []float64, destination *rlwe.Ciphertext, rescale bool, bootstrap bool) {

	cmplx := u.Float64ToComplex128(b)
	encoded := u.Encoder.EncodeNew(cmplx, a.Level(), rlwe.NewScale(u.Params.DefaultScale()), u.Params.LogSlots())
	u.Evaluator.MulRelin(a, encoded, destination)

	if bootstrap {
		u.BootstrapIfNecessary(destination)
	}

	if rescale {
		u.Evaluator.Rescale(destination, rlwe.NewScale(u.Scale), destination)
	}

}

func (u Utils) MultiplyConstArrayNew(a rlwe.Ciphertext, b []float64, rescale bool, bootstrap bool) *rlwe.Ciphertext {

	cmplx := u.Float64ToComplex128(b)
	encoded := u.Encoder.EncodeNew(cmplx, a.Level(), rlwe.NewScale(u.Params.DefaultScale()), u.Params.LogSlots())
	result := u.Evaluator.MulRelinNew(&a, encoded)

	if bootstrap {
		u.BootstrapIfNecessary(result)
	}

	if rescale {
		u.Evaluator.Rescale(result, rlwe.NewScale(u.Scale), result)
	}
	return result
}

func (u Utils) SwitchToSameModCoeff(a *rlwe.Ciphertext, b *rlwe.Ciphertext) (*rlwe.Ciphertext, *rlwe.Ciphertext) {

	if a.Level() != b.Level() {

		var requireSwitch *rlwe.Ciphertext
		var constant *rlwe.Ciphertext

		if a.Level() > b.Level() {
			requireSwitch = a.CopyNew()
			constant = b
		} else {
			requireSwitch = b.CopyNew()
			constant = a
		}

		u.Evaluator.DropLevel(requireSwitch, requireSwitch.Level()-constant.Level())

		return requireSwitch, constant

	}

	return a, b

}

func (u Utils) MultiplyConst(a *rlwe.Ciphertext, b float64, destination *rlwe.Ciphertext, rescale bool, bootstrap bool) {

	originalScale := a.Scale.Float64()
	u.Evaluator.MultByConst(a, b, destination)

	if bootstrap {
		u.BootstrapIfNecessary(destination)
	}

	if rescale && destination.Scale.Float64() != originalScale {
		u.Evaluator.Rescale(destination, rlwe.NewScale(u.Scale), destination)
	}

}

func (u Utils) MultiplyConstNew(a *rlwe.Ciphertext, b float64, rescale bool, bootstrap bool) *rlwe.Ciphertext {

	destination := ckks.NewCiphertext(u.Params, a.Degree(), a.Level())
	u.MultiplyConst(a, b, destination, rescale, bootstrap)

	return destination

}

func (u Utils) ReEncodeAsNTT(a *rlwe.Plaintext) {

	if !a.IsNTT {

		// Reencode as ntt
		data := u.Encoder.Decode(a, u.Params.LogSlots())
		u.Encoder.Encode(data, a, u.Params.LogSlots())

	}

}

func (u Utils) MultiplyConcurrent(a *rlwe.Ciphertext, b *rlwe.Ciphertext, rescale bool, c chan *rlwe.Ciphertext) {

	eval := u.Evaluator.ShallowCopy()
	a, b = u.SwitchToSameModCoeff(a, b)
	result := eval.MulRelinNew(a, b)

	if rescale {
		eval.Rescale(result, rlwe.NewScale(u.Scale), result)
	}

	c <- result
}

func (u Utils) MultiplyPlainConcurrent(a *rlwe.Ciphertext, b *rlwe.Plaintext, rescale bool, c chan *rlwe.Ciphertext) {

	u.ReEncodeAsNTT(b)
	eval := u.Evaluator.ShallowCopy()
	result := eval.MulRelinNew(a, b)

	if rescale {
		eval.Rescale(result, rlwe.NewScale(u.Scale), result)
	}

	c <- result
}
