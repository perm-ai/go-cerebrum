package utility

import "github.com/ldsec/lattigo/v2/ckks"

func (utils Utils) Add(a ckks.Ciphertext, b ckks.Ciphertext, destination *ckks.Ciphertext) {

	utils.EqualizeScale(&a, &b)
	utils.Evaluator.Add(a, b, destination)

}

func (utils Utils) AddNew(a ckks.Ciphertext, b ckks.Ciphertext) ckks.Ciphertext {

	utils.EqualizeScale(&a, &b)
	ct := utils.Evaluator.AddNew(a, b)

	return *ct

}

func (utils Utils) Sub(a ckks.Ciphertext, b ckks.Ciphertext, destination *ckks.Ciphertext) {

	utils.EqualizeScale(&a, &b)
	utils.Evaluator.Sub(a, b, destination)

}

func (utils Utils) SubNew(a ckks.Ciphertext, b ckks.Ciphertext) ckks.Ciphertext {

	utils.EqualizeScale(&a, &b)
	ct := utils.Evaluator.SubNew(a, b)

	return *ct

}

func (utils Utils) EqualizeScale(a *ckks.Ciphertext, b *ckks.Ciphertext) {

	var constant ckks.Ciphertext
	var requiredMult ckks.Ciphertext

	if a.Scale() != b.Scale() {

		if a.Scale() > b.Scale() {
			constant = *a
			requiredMult = *b
		} else {
			constant = *b
			requiredMult = *a
		}

		rescaleBy := constant.Scale() / requiredMult.Scale()
		rescaler := utils.Float64ToComplex128(utils.GenerateFilledArray(1))

		encodedRescaler := ckks.NewPlaintext(&utils.Params, requiredMult.Level(), rescaleBy)
		utils.Encoder.EncodeNTT(encodedRescaler, rescaler, utils.Params.LogSlots())

		utils.Evaluator.MulRelin(&requiredMult, encodedRescaler, &utils.RelinKey, &requiredMult)

	}

}
