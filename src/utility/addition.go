package utility

import "github.com/ldsec/lattigo/v2/ckks"

func (utils Utils) Add(a ckks.Ciphertext, b ckks.Ciphertext, destination *ckks.Ciphertext) {

	// Add two ciphertext together and save result to destination given

	utils.Evaluator.Add(a, b, destination)

}

func (utils Utils) AddNew(a ckks.Ciphertext, b ckks.Ciphertext) ckks.Ciphertext {

	// Add two ciphertext together and return result as a new ciphertext

	ct := utils.Evaluator.AddNew(a, b)

	return *ct

}
func (utils Utils) AddConst(a *ckks.Ciphertext, b []float64, destination *ckks.Ciphertext) {
	utils.Evaluator.AddConst(a, b, destination)

}

func (utils Utils) AddConstNew(a *ckks.Ciphertext, b []float64) *ckks.Ciphertext {
	ct := utils.Evaluator.AddConstNew(a, b)
	return ct
}

func (utils Utils) Sub(a ckks.Ciphertext, b ckks.Ciphertext, destination *ckks.Ciphertext) {

	// Subtract two ciphertext together and save result to destination given

	utils.Evaluator.Sub(a, b, destination)

}

func (utils Utils) SubNew(a ckks.Ciphertext, b ckks.Ciphertext) ckks.Ciphertext {

	// Subtract two ciphertext together and return result as a new ciphertext

	ct := utils.Evaluator.SubNew(a, b)

	return *ct

}
