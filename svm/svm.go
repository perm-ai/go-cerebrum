package svm

import (
	"math/rand"
	"time"

	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/utility"
)

type SVM struct {
	u      		*utility.Utils
	Features	int
	Weights 	[]*rlwe.Ciphertext
	Alphas		*rlwe.Ciphertext
	kernel 		Kernel
}

func NewSVM(u utility.Utils, feature int, kernel Kernel) SVM {

	weightPlain := make([]float64, u.Params.Slots())

	encryptedWeight := make([]*rlwe.Ciphertext, feature)

	for i := range encryptedWeight {
		encryptedWeight[i] = u.EncryptToPointer(weightPlain)
	}

	return SVM{&u, feature, encryptedWeight, &rlwe.Ciphertext{}, kernel}

}

// Optimize the alpha value of an SVM model, if kernel is linear compute weight.
// This optimization problem implements a Pegasos method to optimize primal SVM with kernel function
func (model *SVM) Fit(x []*rlwe.Ciphertext, y *rlwe.Ciphertext, dataLength int, iterations int, lambda float64){

	model.Alphas = model.u.EncryptToPointer(model.u.GenerateFilledArray(0))

	for t := 1; t <= iterations; t++{

		// Get random index
		rand.Seed(time.Now().UnixNano())
		it := rand.Intn(dataLength)

		// Create an array of ciphertext to store each feature of data with random index
		xi := make([]*rlwe.Ciphertext, model.Features)

		// Create filter to filter out data at random index
		filter := make([]float64, model.u.Params.Slots())
		filter[it] = 1
		encodedFilter := model.u.EncodePlaintextFromArray(filter)

		// Loop through training data set and apply filter to extract data from that random index out
		// then use sum element in place to make that ciphertext filled with that randomed index
		for feature := range x{
			xi[feature] = model.u.MultiplyPlainNew(x[feature], encodedFilter, true, true)
			model.u.SumElementsInPlace(xi[feature])
		}

		// Calculate alpha * y
		scalar := model.u.MultiplyNew(model.Alphas, y, true, false)

		// Calculate kernel
		decision, kernelErr := model.kernel.Calculate(xi, x)

		// Catch kernel error
		if kernelErr != nil {
			panic(kernelErr)
		}

		// Calculate decision
		model.u.Multiply(decision, scalar, decision, true, false)
		model.u.SumElementsInPlace(decision)

		// Calculate and encode eta
		etaArray := make([]float64, model.u.Params.Slots())
		etaArray[it] = 1.0 / (lambda * float64(t))
		eta := model.u.EncodePlaintextFromArray(etaArray)

		// Apply eta
		model.u.MultiplyPlain(decision, eta, decision, true, false)

		// Pass through decision function turning number <1 into 1 and >1 into 0
		// TODO: Add decision function

		// Add decision to alpha
		model.u.Add(decision, model.Alphas, model.Alphas)
	}

	if model.kernel.Type() == "linear" {

		// Calculate weight if it is a linear kernel
		for feature := range model.Weights{

			model.Weights[feature] = model.u.MultiplyNew(model.Alphas, y, true, false)
			model.u.Multiply(x[feature], model.Weights[feature], model.Weights[feature], true, false)
			model.u.SumElementsInPlace(model.Weights[feature])
			
			regularizationConst := 1.0 / (lambda * float64(iterations))
			regularization := model.u.EncodePlaintextFromArray(model.u.GenerateFilledArray(regularizationConst))

			model.u.MultiplyPlain(model.Weights[feature], regularization, model.Weights[feature], true, false)

		}

	}

}