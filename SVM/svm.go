package svm

import (
	"fmt"
	"math"

	"github.com/perm-ai/go-cerebrum/array"
)

// NORMALIZE DATA BEFORE TRAINING

type SVM struct {
	weights []float64
}

func NewSVMModel() SVM {
	weights := array.GeneratePlainArray(0, 4)
	return SVM{weights}
}

func (model *SVM) UpdateSVMGradients(ascend []float64, l_rate float64, numOfFeatures int) {
	for i := 0; i < numOfFeatures; i++ {
		model.weights[i] = model.weights[i] - (ascend[i] * l_rate)
	}
}

func (model *SVM) ComputeCostGradient(data [][]float64, target []float64, numOfFeatures int, regularizationStrength float64) []float64 {
	dw := make([]float64, 4) // create weights array of [data1, data2, data3, intercept]
	var distance []float64   // n # of training examples
	// distance = 1-(y_batch * (x dot weights))
	for i := 0; i < len(data); i++ {
		// distance[i] = 1 - (target[i] * (model.weights[0]*data1[i]) + (model.weights[1]*data2[i]) + (model.weights[2]*data3[i]) + (model.weights[3]*1))
		var weightsDotData float64 // for loop for all features
		for k := 0; k < numOfFeatures; k++ {
			weightsDotData += model.weights[k] * data[i][k]
		}
		distance[i] = 1 - (target[i] * weightsDotData)
	}

	for i, d := range distance {
		var di []float64
		if math.Max(0, float64(d)) == 0 {
			di = model.weights
		} else {
			offset := array.MulConstantArrayNew(target[i], data[i])            // ans = ybatch * xbatch
			offset = array.MulConstantArrayNew(regularizationStrength, offset) // ans = ans * regularization strength
			di = array.SubtArraysNew(model.weights, offset)
		}
		dw = array.AddArraysNew(dw, di)
	}

	dw = array.MulConstantArrayNew(float64(1/len(data)), dw) // average dw
	return dw

}

func (model *SVM) TrainSVM(data [][]float64, target []float64, epoch int, l_rate float64, numOfFeatures int, regularizationStrength float64) []float64 {
	weights := make([]float64, 4)
	model.weights = weights
	for i := 0; i < epoch; i++ {
		fmt.Printf("Start Training epoch number %d \n", i+1)
		ascent := model.ComputeCostGradient(data, target, numOfFeatures, regularizationStrength) // get dw
		model.UpdateSVMGradients(ascent, l_rate, numOfFeatures)
		fmt.Printf("The updated weights is %f \n\n", model.weights)
	}

	return model.weights
}
