package utility

import (
	"github.com/ldsec/lattigo/v2/ckks"
)

func (utils Utils) Add(a *ckks.Ciphertext, b *ckks.Ciphertext, destination *ckks.Ciphertext) {

	// Add two ciphertext together and save result to destination given
	// utils.EqualizeScale(&a, &b)
	utils.Evaluator.Add(a, b, destination)

}

func (utils Utils) AddNew(a *ckks.Ciphertext, b *ckks.Ciphertext) *ckks.Ciphertext {

	// Add two ciphertext together and return result as a new ciphertext
	utils.EqualizeScale(a, b)
	ct := utils.Evaluator.AddNew(a, b)

	return ct

}

func (u Utils) AddPlain(a *ckks.Ciphertext, b *ckks.Plaintext, destination *ckks.Ciphertext) {

	u.ReEncodeAsNTT(b)
	u.Evaluator.Add(a, b, destination)

}

func (utils Utils) AddPlainNew(a *ckks.Ciphertext, b *ckks.Plaintext) *ckks.Ciphertext {

	// Add two ciphertext together and return result as a new ciphertext

	ct := utils.Evaluator.AddNew(a, b)

	return ct

}

func (utils Utils) AddConst(a *ckks.Ciphertext, b []float64, destination *ckks.Ciphertext) {

	// Add overwrite ciphertext and constant

	utils.Evaluator.AddConst(a, &b, destination)

}

func (utils Utils) AddConstNew(a *ckks.Ciphertext, b []float64) *ckks.Ciphertext {

	// Add and create a new ciphertext

	ct := utils.Evaluator.AddConstNew(a, &b)
	return ct
}

func (utils Utils) Sub(a *ckks.Ciphertext, b *ckks.Ciphertext, destination *ckks.Ciphertext) {

	// Subtract two ciphertext together and save result to destination given
	utils.EqualizeScale(a, b)
	utils.Evaluator.Sub(a, b, destination)

}

// Subtract two ciphertext together and return result as a new ciphertext
func (utils Utils) SubNew(a *ckks.Ciphertext, b *ckks.Ciphertext) *ckks.Ciphertext {

	ct := utils.Evaluator.SubNew(a, b)

	return ct

}

// Subtract ciphertext and plaintext and save result to destination given
func (utils Utils) SubPlain(a *ckks.Ciphertext, b *ckks.Plaintext, destination *ckks.Ciphertext) {

	utils.Evaluator.Sub(a, b, destination)

}

// Subtract ciphertext and plaintext and return result as a new ciphertext
func (utils Utils) SubPlainNew(a *ckks.Ciphertext, b *ckks.Plaintext) *ckks.Ciphertext {

	ct := utils.Evaluator.SubNew(a, b)

	return ct

}

func (utils Utils) EqualizeScale(a *ckks.Ciphertext, b *ckks.Ciphertext) {

	var higherScale *ckks.Ciphertext
	var lowerScale *ckks.Ciphertext

	if a.Scale != b.Scale {

		if a.Scale > b.Scale {
			higherScale = a
			lowerScale = b
		} else {
			higherScale = b
			lowerScale = a
		}

		if higherScale.Scale / lowerScale.Scale < 4{
			rescaler := ckks.NewPlaintext(utils.Params, higherScale.Level(), higherScale.Scale/lowerScale.Scale)
			utils.Encoder.EncodeNTT(rescaler, utils.Float64ToComplex128(utils.GenerateFilledArray(1)), utils.Params.LogSlots())
			utils.MultiplyPlain(lowerScale, rescaler, lowerScale, false, false)
		}

	}

}
