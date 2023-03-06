package utility

import (
	"math"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

func (u Utils) ExpNew(ciphertext *rlwe.Ciphertext, size int) *rlwe.Ciphertext {

	// 1 + x + x^2/2 + x^3/6 + x^4/24 + x^5/120 + x^6/720 + x^7/5040
	coeffs := []float64{
		1.0,
		1.0,
		1.0/2.0,
		1.0/6.0,
		1.0/24.0,
		1.0/120.0,
		1.0/720.0,
		1.0/5040.0,
	}

	poly := NewPolynomial(coeffs, u)
	return poly.EvaluateDegree7(ciphertext, size)

}

// This function is used to calculate 1 / (n * stretchScale)
func (u Utils) InverseApproxNew(ciphertext *rlwe.Ciphertext, stretchScale float64, size int) *rlwe.Ciphertext {

	// Degree 7 approximation of inverse function

	coeff := []float64{
		8,
		-28,
		56,
		-70,
		56,
		-28,
		8,
		-1,
	}

	for c := range coeff{
		coeff[c] = math.Pow(stretchScale, float64(c)) * coeff[c]
	}

	poly := NewPolynomial(coeff, u)
	return poly.EvaluateDegree7(ciphertext, size)

}

func (u Utils) InverseNew(ct *rlwe.Ciphertext, horizontalStretchScale float64, size int) *rlwe.Ciphertext {

	// Calculate approximate inverse of a ciphertext

	if horizontalStretchScale != 1 {

		// Calculate inverse
		inversed := u.InverseApproxNew(ct, horizontalStretchScale, size)

		// Apply vertical stretch
		return u.MultiplyConstNew(inversed, horizontalStretchScale, true, false)

	} else {
		return u.InverseApproxNew(ct, 1, size)
	}

}

// Polynomial degree 3 approximation function of square root. Bound between [0, bound]. Set scaleBoundBack to true to get sqrt(ct) when bound > 1.75.
// Please note that as bound increase the accuracy of approximation will decrease
func (u Utils) SqrtApprox(ct *rlwe.Ciphertext, degree int, bound float64, scaleBoundBack bool) *rlwe.Ciphertext {

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
	var result *rlwe.Ciphertext

	if result, err = u.Evaluator.EvaluatePoly(ct, poly, rlwe.NewScale(math.Pow(2, 35))); err != nil {
		panic(err)
	}

	if scaleBoundBack && bound > originalBound {
		scalingBackFactor := (1 / math.Sqrt(m))
		u.MultiplyConst(result, scalingBackFactor, result, true, false)
	}

	return result

}
