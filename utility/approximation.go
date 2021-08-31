package utility

import (
	"math"

	"github.com/ldsec/lattigo/v2/ckks"
)

func (u Utils) ExpNew(ciphertext *ckks.Ciphertext) *ckks.Ciphertext {
	//1 + x + x^2/2 + x^3/6 + x^4/24 + x^5/120 + x^6/720 + x^7/5040
	// coeffs := []complex128{
	// 	complex(1.0, 0),
	// 	complex(1.0, 0),
	// 	complex(1.0/2, 0),
	// 	complex(1.0/6, 0),
	// 	complex(1.0/24, 0),
	// 	complex(1.0/120, 0),
	// 	complex(1.0/720, 0),
	// 	complex(1.0/5040, 0),
	// }

	//deg1
	x := ciphertext
	deg1 := x.CopyNew()

	//deg2
	x2 := u.MultiplyNew(*x.CopyNew(), *x.CopyNew(), true, false)
	deg2 := u.MultiplyConstNew(x2.CopyNew(), 0.5, true, false)
	sum := u.AddNew(*deg1, deg2)

	//deg3
	x3 := u.MultiplyNew(*x.CopyNew(), *x2.CopyNew(), true, false)
	deg3 := u.MultiplyConstNew(x3.CopyNew(), 1.0/6, true, false)
	sum = u.AddNew(deg3, sum)

	//deg4
	x4 := u.MultiplyNew(*x2.CopyNew(), *x2.CopyNew(), true, false)
	deg4 := u.MultiplyConstNew(x4.CopyNew(), 1.0/24, true, false)
	sum = u.AddNew(deg4, sum)

	//deg5
	x5 := u.MultiplyNew(*x.CopyNew(), *x4.CopyNew(), true, false)
	deg5 := u.MultiplyConstNew(x5.CopyNew(), 1.0/120, true, false)
	sum = u.AddNew(deg5, sum)

	//deg6
	x6 := u.MultiplyNew(*x4.CopyNew(), *x2.CopyNew(), true, false)
	deg6 := u.MultiplyConstNew(x6.CopyNew(), 1.0/720, true, false)
	sum = u.AddNew(deg6, sum)

	//deg7
	x7 := u.MultiplyNew(*x4.CopyNew(), *x3.CopyNew(), true, false) //now deg6
	deg7 := u.MultiplyConstNew(x7.CopyNew(), 1.0/5040, true, false)
	sum = u.AddNew(deg7, sum)

	//add 1
	sum = *u.AddConstNew(&sum, u.GenerateFilledArray(1.0))

	return &sum

}

// This function is used to calculate 1 / (n * stretchScale)
func (u Utils) InverseApproxNew(ciphertext *ckks.Ciphertext, stretchScale float64) *ckks.Ciphertext {

	// Degree 7 approximation of inverse function

	//deg1
	x := ciphertext
	deg1 := u.MultiplyConstNew(x.CopyNew(), stretchScale*-28.0, true, false)

	//deg2
	x2 := u.MultiplyNew(*x.CopyNew(), *x.CopyNew(), true, false)
	deg2 := u.MultiplyConstNew(x2.CopyNew(), math.Pow(stretchScale, 2)*56.0, true, false)
	sum := u.AddNew(deg1, deg2)

	//deg3
	x3 := u.MultiplyNew(*x.CopyNew(), *x2.CopyNew(), true, false)
	deg3 := u.MultiplyConstNew(x3.CopyNew(), math.Pow(stretchScale, 3)*-70.0, true, false)
	sum = u.AddNew(deg3, sum)

	//deg4
	x4 := u.MultiplyNew(*x2.CopyNew(), *x2.CopyNew(), true, false)
	deg4 := u.MultiplyConstNew(x4.CopyNew(), math.Pow(stretchScale, 4)*56.0, true, false)
	sum = u.AddNew(deg4, sum)

	//deg5
	x5 := u.MultiplyNew(*x.CopyNew(), *x4.CopyNew(), true, false)
	deg5 := u.MultiplyConstNew(x5.CopyNew(), math.Pow(stretchScale, 5)*-28.0, true, false)
	sum = u.AddNew(deg5, sum)

	//deg6
	x6 := u.MultiplyNew(*x4.CopyNew(), *x2.CopyNew(), true, false)
	deg6 := u.MultiplyConstNew(x6.CopyNew(), math.Pow(stretchScale, 6)*8.0, true, false)
	sum = u.AddNew(deg6, sum)

	//deg7
	x7 := u.MultiplyNew(*x4.CopyNew(), *x3.CopyNew(), true, false) //now deg6
	deg7 := u.MultiplyConstNew(x7.CopyNew(), math.Pow(stretchScale, 7)*-1.0, true, false)
	sum = u.AddNew(deg7, sum)

	//add 1
	sum = *u.AddConstNew(&sum, u.GenerateFilledArray(8.0))

	return &sum

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

// Polynomial degree 3 approximation function of square root. Bound between [0, bound]. Set scaleBoundBack to true to get sqrt(ct) when bound > 1.75.
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
