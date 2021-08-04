package utility

import (

	"math"

	"github.com/ldsec/lattigo/v2/ckks"
)

func (u Utils) ExpNew(ciphertext *ckks.Ciphertext) *ckks.Ciphertext {

	coeffs := []complex128{
		complex(1.0, 0),
		complex(1.0, 0),
		complex(1.0/2, 0),
		complex(1.0/6, 0),
		complex(1.0/24, 0),
		complex(1.0/120, 0),
		complex(1.0/720, 0),
		complex(1.0/5040, 0),
	}

	poly := ckks.NewPoly(coeffs)

	var err error
	var result *ckks.Ciphertext

	if result, err = u.Evaluator.EvaluatePoly(ciphertext, poly, ciphertext.Scale); err != nil {
		panic(err)
	}

	return result

}

func (u Utils) InverseApproxNew(ciphertext *ckks.Ciphertext, stretchScale float64) *ckks.Ciphertext {

	// This function is used to calculate 1 / (n * stretchScale)

	// Degree 7 approximation of inverse function

	coeffs := []complex128{
		complex(8, 0),
		complex(-28, 0),
		complex(56, 0),
		complex(-70, 0),
		complex(56, 0),
		complex(-28, 0),
		complex(8, 0),
		complex(-1, 0),
	}

	if stretchScale != 1 {

		for i, coeff := range coeffs {

			coeffs[i] = complex(real(coeff)*math.Pow(stretchScale, float64(i)), 0)

		}

	}

	poly := ckks.NewPoly(coeffs)

	var err error
	var result *ckks.Ciphertext

	if result, err = u.Evaluator.EvaluatePoly(ciphertext, poly, ciphertext.Scale); err != nil {
		panic(err)
	}

	return result

}

func (u Utils) InverseNew(ct *ckks.Ciphertext, horizontalStretchScale float64) ckks.Ciphertext {

	// Calculate approximate inverse of a ciphertext. only works within the bound of [0.175, 1.5]
	// Costs 3 mutiplicative depth
	// horizontalStretchScale is applied to get the number between wanted bound
	// This function, if used with correct ciphertext and stretching params, will return 1/n.

	if horizontalStretchScale != 1 {

		// Calculate inverse
		inversed := u.InverseApproxNew(ct, horizontalStretchScale)

		// Apply vertical stretch
		return u.MultiplyConstNew(inversed, horizontalStretchScale, true, false)

	} else {
		return *u.InverseApproxNew(ct, 1)
	}

}

// Polynomial degree 3 approximation function of square root. Bound between [0, bound]. Set scaleBoundBack to true to get sqrt(ct) when bound > 1.75
// Please note that as bound increase the accuracy of approximation will decrease
func (u Utils) SqrtApprox(ct *ckks.Ciphertext, degree int, bound float64, scaleBoundBack bool) *ckks.Ciphertext {

	if degree != 3 && degree != 7 {
		panic("Only degree 3 or 7 is available.")
	}

	originalBound := 1.75

	coeffs := map[int][]complex128{}

	coeffs[3] = []complex128{
		complex(0.3125, 0),
		complex((15.0 / 16.0), 0),
		complex((-5.0 / 16.0), 0),
		complex((1.0 / 16.0), 0),
	}

	// TODO: IMPLEMENT DEG 7 APPROXIMATION
	// coeffs[7] = []complex128{
	// 	complex(1, 0),
	// 	complex(0.5, 0),
	// 	complex(-0.125, 0),
	// 	complex(0.0625, 0),
	// 	complex((-5 / 128), 0),
	// 	complex((7 / 256), 0),
	// 	complex((-21 / 1024), 0),
	// 	complex((231 / 14336), 0),
	// }

	var m float64

	if bound > originalBound {

		// Declare coeff rescaling factor
		m = (originalBound / bound)

		for i, coeff := range coeffs[degree] {
			coeffs[degree][i] = complex(real(coeff)*math.Pow(m, float64(i)), 0)
		}

	}

	poly := ckks.NewPoly(coeffs[degree])

	var err error
	var result *ckks.Ciphertext

	if result, err = u.Evaluator.EvaluatePoly(ct, poly, math.Pow(2, 35)); err != nil {
		panic(err)
	}

	if scaleBoundBack && bound > originalBound {
		scalingBackFactor := (1 / math.Sqrt(m))
		u.MultiplyConst(result, scalingBackFactor, result, true, false)
	}

	return result

}
