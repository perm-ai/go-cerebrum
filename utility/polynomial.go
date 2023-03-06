package utility

import (
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

type Coefficient struct {
	Degree  int
	Value   float64
	Encoded map[int]*rlwe.Plaintext
}

type Polynomial struct {
	u     Utils
	Coeff []Coefficient
}

// Create new polynomial struct
// index of coeff is the degree that that coeff serve (e.g. coeff[0] is for degree 0)
func NewPolynomial(coeff []float64, utils Utils) Polynomial {

	coeffs := make([]Coefficient, len(coeff))

	for i := range coeff {

		encoded := make(map[int]*rlwe.Plaintext)

		if i != 0 {
			pt := utils.EncodePlaintextFromArray(utils.GenerateFilledArray(coeff[i]))
			encoded[utils.Params.Slots()] = pt
		}

		coeffs[i] = Coefficient{Degree: i, Value: coeff[i], Encoded: encoded}

	}

	return Polynomial{u: utils, Coeff: coeffs}

}

func (poly Polynomial) EvaluateDegree7(x *rlwe.Ciphertext, size int) *rlwe.Ciphertext {

	channels := make([]chan *rlwe.Ciphertext, 8)
	slots := poly.u.Params.Slots()

	x2 := poly.u.MultiplyNew(x.CopyNew(), x.CopyNew(), true, false)
	x4 := poly.u.MultiplyNew(x2.CopyNew(), x2.CopyNew(), true, false)

	// Calc degree 7
	if poly.Coeff[7].Value != 0 {
		channels[7] = make(chan *rlwe.Ciphertext)
		go func(u Utils, c chan *rlwe.Ciphertext) {

			c7x1 := u.MultiplyPlainNew(x.CopyNew(), poly.Coeff[7].Encoded[slots], true, false)
			c7x3 := u.MultiplyNew(c7x1, x2.CopyNew(), true, false)
			c7x7 := u.MultiplyNew(c7x3, x4.CopyNew(), true, false)
			c <- c7x7

		}(poly.u.CopyWithClonedEval(), channels[7])
	}

	// Calc degree 6
	if poly.Coeff[6].Value != 0 {
		channels[6] = make(chan *rlwe.Ciphertext)
		go func(u Utils, c chan *rlwe.Ciphertext) {

			c6x2 := u.MultiplyPlainNew(x2.CopyNew(), poly.Coeff[6].Encoded[slots], true, false)
			c6x6 := u.MultiplyNew(c6x2, x4.CopyNew(), true, false)
			c <- c6x6

		}(poly.u.CopyWithClonedEval(), channels[6])
	}

	// Calc degree 5
	if poly.Coeff[5].Value != 0 {
		channels[5] = make(chan *rlwe.Ciphertext)
		go func(u Utils, c chan *rlwe.Ciphertext) {

			c5x1 := u.MultiplyPlainNew(x.CopyNew(), poly.Coeff[5].Encoded[slots], true, false)
			c5x5 := u.MultiplyNew(c5x1, x4.CopyNew(), true, false)
			c <- c5x5

		}(poly.u.CopyWithClonedEval(), channels[5])
	}

	// Calc degree 4
	if poly.Coeff[4].Value != 0 {
		channels[4] = make(chan *rlwe.Ciphertext)
		go func(u Utils, c chan *rlwe.Ciphertext) {

			c4x2 := u.MultiplyPlainNew(x2.CopyNew(), poly.Coeff[4].Encoded[slots], true, false)
			c4x4 := u.MultiplyNew(c4x2, x2.CopyNew(), true, false)
			c <- c4x4

		}(poly.u.CopyWithClonedEval(), channels[4])
	}

	// Calc degree 3
	if poly.Coeff[3].Value != 0 {
		channels[3] = make(chan *rlwe.Ciphertext)
		go func(u Utils, c chan *rlwe.Ciphertext) {

			c3x1 := u.MultiplyPlainNew(x.CopyNew(), poly.Coeff[3].Encoded[slots], true, false)
			c3x3 := u.MultiplyNew(c3x1, x2.CopyNew(), true, false)
			c <- c3x3

		}(poly.u.CopyWithClonedEval(), channels[3])
	}

	// Calc degree 2
	if poly.Coeff[2].Value != 0 {
		channels[2] = make(chan *rlwe.Ciphertext)
		go func(u Utils, c chan *rlwe.Ciphertext) {

			c2x2 := u.MultiplyPlainNew(x2.CopyNew(), poly.Coeff[2].Encoded[slots], true, false)
			c <- c2x2

		}(poly.u.CopyWithClonedEval(), channels[2])
	}

	// Calc degree 1
	if poly.Coeff[1].Value != 0 {
		channels[1] = make(chan *rlwe.Ciphertext)
		go func(u Utils, c chan *rlwe.Ciphertext) {

			c1x1 := u.MultiplyPlainNew(x.CopyNew(), poly.Coeff[1].Encoded[slots], true, false)
			c <- c1x1

		}(poly.u.CopyWithClonedEval(), channels[1])
	}

	sum := poly.u.EncryptToPointer(poly.u.GenerateFilledArray(0))
	for c := range poly.Coeff {
		if poly.Coeff[c].Value != 0 && c != 0 {
			result := <-channels[c]
			poly.u.Add(result, sum, sum)
		} else if c == 0 {
			if _, ok := poly.Coeff[c].Encoded[size]; !ok {

				encoded := poly.u.EncodePlaintextFromArray(poly.u.GenerateFilledArraySize(poly.Coeff[c].Value, size))
				poly.Coeff[c].Encoded[size] = encoded

			}
			poly.u.AddPlain(sum, poly.Coeff[c].Encoded[size], sum)
		}
	}

	return sum

}
