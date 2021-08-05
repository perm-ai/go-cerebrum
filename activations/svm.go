package activations

import (
	"math"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

type SVM struct {
	U        		utility.Utils
}

func NewSvmActivation(u utility.Utils) SVM {
	return SVM{u}
}

func (s SVM) Forward(input ckks.Ciphertext, bound float64) ckks.Ciphertext {

	coeffs := []complex128{
		complex(0.136205, 0),
		complex(0.37125, 0),
		complex(0.297, 0),
		complex(0.0264, 0),
		complex(-0.04928, 0),
		complex(-0.0033792, 0),
		complex(0.0045056, 0),
		complex(-0.000514926, 0),
	}

	stretchScale := 1.25 / bound

	if stretchScale < 1.25 {
		for i, coeff := range coeffs {
			coeffs[i] = complex(real(coeff) * math.Pow(stretchScale, float64(i)), 0)
		}
	}

	poly := ckks.NewPoly(coeffs)

	var err error
	var result *ckks.Ciphertext

	if result, err = s.U.Evaluator.EvaluatePoly(&input, poly, input.Scale); err != nil {
		panic(err)
	}

	return *result

}

func (s SVM) Backward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {
	return ckks.Ciphertext{}
}