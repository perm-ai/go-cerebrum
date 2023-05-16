package utility

import (
	// "github.com/ldsec/lattigo/v2/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"

)

func (utils Utils) Add(a *rlwe.Ciphertext, b *rlwe.Ciphertext, destination *rlwe.Ciphertext) {

	// Add two ciphertext together and save result to destination given
	// utils.EqualizeScale(&a, &b)
	utils.Evaluator.Add(a, b, destination)

}

func (utils Utils) AddNew(a *rlwe.Ciphertext, b *rlwe.Ciphertext) *rlwe.Ciphertext {

	// Add two ciphertext together and return result as a new ciphertext
	utils.EqualizeScale(a, b)
	ct := utils.Evaluator.AddNew(a, b)

	return ct

}

func (u Utils) AddPlain(a *rlwe.Ciphertext, b *rlwe.Plaintext, destination *rlwe.Ciphertext) {

	u.ReEncodeAsNTT(b)
	u.Evaluator.Add(a, b, destination)

}

func (utils Utils) AddPlainNew(a *rlwe.Ciphertext, b *rlwe.Plaintext) *rlwe.Ciphertext {

	// Add two ciphertext together and return result as a new ciphertext

	ct := utils.Evaluator.AddNew(a, b)

	return ct

}

func (utils Utils) AddConst(a *rlwe.Ciphertext, b []float64, destination *rlwe.Ciphertext) {

	// Add overwrite ciphertext and constant

	utils.Evaluator.AddConst(a, &b, destination)

}

func (utils Utils) AddConstNew(a *rlwe.Ciphertext, b []float64) *rlwe.Ciphertext {

	// Add and create a new ciphertext

	ct := utils.Evaluator.AddConstNew(a, &b)
	return ct
}

func (utils Utils) Sub(a *rlwe.Ciphertext, b *rlwe.Ciphertext, destination *rlwe.Ciphertext) {

	// Subtract two ciphertext together and save result to destination given
	utils.EqualizeScale(a, b)
	utils.Evaluator.Sub(a, b, destination)

}

// Subtract two ciphertext together and return result as a new ciphertext
func (utils Utils) SubNew(a *rlwe.Ciphertext, b *rlwe.Ciphertext) *rlwe.Ciphertext {

	ct := utils.Evaluator.SubNew(a, b)

	return ct

}

// Subtract ciphertext and plaintext and save result to destination given
func (utils Utils) SubPlain(a *rlwe.Ciphertext, b *rlwe.Plaintext, destination *rlwe.Ciphertext) {

	utils.Evaluator.Sub(a, b, destination)

}

// Subtract ciphertext and plaintext and return result as a new ciphertext
func (utils Utils) SubPlainNew(a *rlwe.Ciphertext, b *rlwe.Plaintext) *rlwe.Ciphertext {

	ct := utils.Evaluator.SubNew(a, b)

	return ct

}

func (utils Utils) EqualizeScale(a *rlwe.Ciphertext, b *rlwe.Ciphertext) {

	var higherScale *rlwe.Ciphertext
	var lowerScale *rlwe.Ciphertext

	if a.Scale.Float64() != b.Scale.Float64() {

		if a.Scale.Float64() > b.Scale.Float64() {
			higherScale = a
			lowerScale = b
		} else {
			higherScale = b
			lowerScale = a
		}

		if higherScale.Scale.Float64() / lowerScale.Scale.Float64() < 4{
			rescaler := utils.Encoder.EncodeNew(utils.Float64ToComplex128(utils.GenerateFilledArray(1)), higherScale.Level(), rlwe.NewScale(higherScale.Scale.Float64()/lowerScale.Scale.Float64()), utils.Params.LogSlots())
			utils.MultiplyPlain(lowerScale, rescaler, lowerScale, false, false)
		}

	}

}
