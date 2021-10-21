package layers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sync"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/activations"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

//=================================================
//					DENSE LAYER
//=================================================

type Dense struct {
	utils          utility.Utils
	InputUnit      int
	OutputUnit     int
	Weights        [][]*ckks.Ciphertext
	Bias           []*ckks.Ciphertext
	Activation     *activations.Activation
	btspOutput     []bool
	btspActivation []bool
	batchSize      int
	weightLevel    int
	lr			   float64
}

func NewDense(utils utility.Utils, inputUnit int, outputUnit int, activation *activations.Activation, useBias bool, batchSize int, lr float64, weightLevel int) Dense {

	// Generate random weights and biases
	weights := make([][]*ckks.Ciphertext, outputUnit)
	bias := make([]*ckks.Ciphertext, outputUnit)

	// Determine the standard deviation of initial random weight distribution
	weightStdDev := 0.0

	if (*activation).GetType() == "relu" {
		weightStdDev = math.Sqrt(1.0 / float64(inputUnit))
	} else {
		weightStdDev = math.Sqrt(1.0 / float64(inputUnit+outputUnit))
	}

	counter := logger.NewOperationsCounter("Initializing weight", inputUnit*outputUnit + outputUnit)

	var wg sync.WaitGroup

	// Generate initial weight
	for node := 0; node < outputUnit; node++ {

		wg.Add(1)

		go func(nodeIndex int, u utility.Utils) {

			defer wg.Done()

			randomWeight := utils.GenerateRandomNormalArraySeed(inputUnit, weightStdDev, inputUnit + nodeIndex)
			weights[nodeIndex] = make([]*ckks.Ciphertext, inputUnit)

			if useBias {
				bias[nodeIndex] = u.EncryptToLevelScale(u.GenerateFilledArraySize(0, batchSize), weightLevel, math.Pow(2,40))
				counter.Increment()
			}

			var weightWg sync.WaitGroup

			for weight := 0; weight < inputUnit; weight++ {

				weightWg.Add(1)

				go func(weightIndex int, wUtils utility.Utils) {
					defer weightWg.Done()
					weights[nodeIndex][weightIndex] = wUtils.EncryptToLevelScale(u.GenerateFilledArraySize(randomWeight[weightIndex], batchSize), weightLevel, math.Pow(2,40))
					counter.Increment()
				}(weight, u.CopyWithClonedEncryptor())

			}

			weightWg.Wait()

		}(node, utils.CopyWithClonedEncryptor())

	}

	wg.Wait()

	return Dense{utils, inputUnit, outputUnit, weights, bias, activation, []bool{false, false}, []bool{false, false}, batchSize, weightLevel, lr}

}

func (d Dense) Forward(input []*ckks.Ciphertext) Output1d {

	output := make([]*ckks.Ciphertext, d.OutputUnit)
	activatedOutput := make([]*ckks.Ciphertext, d.OutputUnit)

	var wg sync.WaitGroup

	// DEBUG
	timer := logger.StartTimer(fmt.Sprintf("Forward (%d) dot product", d.InputUnit))
	dotProductCounter := logger.NewOperationsCounter(fmt.Sprintf("Forward propagating (%d) multiplying", d.InputUnit), d.InputUnit * d.OutputUnit)

	for node := range d.Weights {

		wg.Add(1)

		go func(nodeIndex int, utils utility.Utils) {
			defer wg.Done()
			output[nodeIndex] = utils.InterDotProduct(input, d.Weights[nodeIndex], !d.btspOutput[0], true, &dotProductCounter)

			// DEBUG start
			if nodeIndex == 0{
				dutils := utils.CopyWithClonedDecryptor()
				fmt.Printf("Forward (%d) post dot sample: %f (L: %d | S: %f)\n", d.InputUnit, dutils.Decrypt(output[nodeIndex])[0:5], output[nodeIndex].Level(), output[nodeIndex].Scale)
			}
			// DEBUG end

			if len(d.Bias) != 0 {
				utils.Add(output[nodeIndex], d.Bias[nodeIndex], output[nodeIndex])
			}

			// DEBUG start
			if nodeIndex == 0{
				dutils := utils.CopyWithClonedDecryptor()
				fmt.Printf("Forward (%d) post bias sample: %f (L: %d | S: %f)\n", d.InputUnit, dutils.Decrypt(output[nodeIndex])[0:5], output[nodeIndex].Level(), output[nodeIndex].Scale)
			}
			// DEBUG end
		}(node, d.utils.CopyWithClonedEval())

	}

	wg.Wait()

	timer.LogTimeTakenSecond()

	if d.btspOutput[0] {

		timer = logger.StartTimer(fmt.Sprintf("Forward (%d) bootstrap", d.InputUnit))
		fmt.Printf("Bootstrapping node\n")

		d.utils.Bootstrap1dInPlace(output, true)

		// DEBUG start
		fmt.Printf("Forward (%d) post btp sample: %f (L: %d | S: %f)\n", d.InputUnit, d.utils.Decrypt(output[0])[0:5], output[0].Level(), output[0].Scale)
		timer.LogTimeTakenSecond()

	}

	if d.Activation != nil {

		timer = logger.StartTimer(fmt.Sprintf("Forward (%d) activation %s", d.InputUnit, (*d.Activation).GetType()))
		activatedOutput = (*d.Activation).Forward(output, d.batchSize)

		// DEBUG start
		fmt.Printf("Forward (%d) activated sample: %f (L: %d | S: %f)\n", d.InputUnit, d.utils.Decrypt(activatedOutput[0])[0:5], activatedOutput[0].Level(), activatedOutput[0].Scale)
		timer.LogTimeTakenSecond()

		if d.btspActivation[0] {
			timer = logger.StartTimer(fmt.Sprintf("Forward (%d) activation %s bootstrapping", d.InputUnit, (*d.Activation).GetType()))
			d.utils.Bootstrap1dInPlace(activatedOutput, true)
			timer.LogTimeTakenSecond()
		}

	}

	return Output1d{output, activatedOutput}

}

// input is A(l-1) - activation of previous layer
// output is Z(l) - output of this layer
// gradient is ∂L/∂A(l) - influence that the activation of this layer has on the next layer
// hasPrevLayer determine whether this function calculates the gradient ∂L/∂A(l-1)
func (d *Dense) Backward(input []*ckks.Ciphertext, output []*ckks.Ciphertext, gradient []*ckks.Ciphertext, hasPrevLayer bool) Gradient1d {

	gradients := Gradient1d{}

	fmt.Printf("Backward gradient wrt output of layer %d: %d\n", d.InputUnit, gradient[0].Level())

	backwardTimer := logger.StartTimer(fmt.Sprintf("Backward (%d)", d.InputUnit))

	// Calculate gradients for last layer
	if d.Activation != nil {

		if (*d.Activation).GetType() != "softmax" {

			gradients.BiasGradient = make([]*ckks.Ciphertext, len(gradient))
			fmt.Printf("Backward (%d) output level: %d\n", d.InputUnit, output[0].Level())

			activationTimer := logger.StartTimer(fmt.Sprintf("Backward (%d) activation %s", d.InputUnit, (*d.Activation).GetType()))
			activationGradient := (*d.Activation).Backward(output, d.batchSize)
			activationTimer.LogTimeTakenSecond()
			fmt.Printf("Backward (%d) activation level: %d\n", d.InputUnit, activationGradient[0].Level())

			backwardTimer = logger.StartTimer(fmt.Sprintf("Backward (%d)", d.InputUnit))
			hasBootstrapped := false

			if activationGradient[0].Level() == 1 && d.btspActivation[1] {

				d.utils.Bootstrap1dInPlace(activationGradient, true)
				hasBootstrapped = true
				
			}

			var wg sync.WaitGroup

			for b := range d.Bias {

				wg.Add(1)

				go func(index int, utils utility.Utils) {
					defer wg.Done()
					gradients.BiasGradient[index] = utils.MultiplyNew(gradient[index], activationGradient[index], true, false)
				}(b, d.utils.CopyWithClonedEval())

			}

			wg.Wait()

			if d.btspActivation[1] && !hasBootstrapped {

				d.utils.Bootstrap1dInPlace(gradients.BiasGradient, true)

			}

		} else {
			gradients.BiasGradient = gradient
		}

	} else {
		gradients.BiasGradient = gradient
	}
	
	timer := logger.StartTimer(fmt.Sprintf("Backward (%d) outer", d.InputUnit))

	if d.InputUnit == 20 {
		b := d.utils.Decrypt(gradients.BiasGradient[7])
		inp := d.utils.Decrypt(input[10])
		fmt.Printf("Backward (%d) sample: %f x %f = ", d.InputUnit, b[0:21], inp[0:21])
	}

	gradients.WeightGradient = d.utils.InterOuter(gradients.BiasGradient, input, true)

	if d.InputUnit == 20 {
		fmt.Printf("%f\n", d.utils.Decrypt(gradients.WeightGradient[7][10])[0:21])
	}
	
	gradients.InputGradient = make([]*ckks.Ciphertext, d.InputUnit)

	if hasPrevLayer {

		timer = logger.StartTimer(fmt.Sprintf("Backward (%d) wrt input", d.InputUnit))

		var inputWg sync.WaitGroup

		// Calculate ∂L/∂A(l-1)
		transposedWeight := d.utils.InterTranspose(d.Weights)

		for xi := range transposedWeight {
			inputWg.Add(1)

			go func(xIndex int, utils utility.Utils){
				defer inputWg.Done()
				gradients.InputGradient[xIndex] = d.utils.InterDotProduct(transposedWeight[xIndex], gradients.BiasGradient, !d.btspOutput[1], true, nil)
			}(xi, d.utils.CopyWithClonedEval())
			
		}

		inputWg.Wait()
		backwardTimer.LogTimeTakenSecond()

		if d.btspOutput[1] {
			btpTimer := logger.StartTimer(fmt.Sprintf("Backward (%d) input gradient bootstrap", d.InputUnit))
			d.utils.Bootstrap1dInPlace(gradients.InputGradient, true)
			btpTimer.LogTimeTakenSecond()
		}

		timer.LogTimeTakenSecond()
		fmt.Printf("Backward (%d) input gradient: %f\n", d.InputUnit, d.utils.Decrypt(gradients.InputGradient[4])[0:5])

	} else {
		backwardTimer.LogTimeTakenSecond()
	}

	fmt.Printf("Backward (%d) weight level: %d \tbias level: %d\n", d.InputUnit, gradients.WeightGradient[0][0].Level(), gradients.BiasGradient[0].Level())

	return gradients

}

func (d *Dense) UpdateGradient(gradient Gradient1d, lr float64) {

	avgScale := lr/float64(d.batchSize)
	batchAverager := d.utils.EncodePlaintextFromArray(d.utils.GenerateFilledArraySize(avgScale, d.batchSize))

	bootstrapGradient := gradient.WeightGradient[0][0].Level() - 1 < d.weightLevel 
	ciphertextNeeded := int(math.Ceil(float64(d.InputUnit * d.OutputUnit) / float64(d.utils.Params.Slots())))
	weightToBootstrap := make([]utility.SafeSum, ciphertextNeeded)
	biasToBootstrap := utility.SafeSum{}

	// create weight group
	var wg sync.WaitGroup

	averageTimer := logger.StartTimer(fmt.Sprintf("SGD gradient filtering and averaging (%d)", d.InputUnit))
	counter := logger.NewOperationsCounter(fmt.Sprintf("SGD (%d)", d.InputUnit), (d.InputUnit * d.OutputUnit) + d.OutputUnit)

	for node := range d.Weights {

		wg.Add(1)

		go func(nodeIndex int, utils utility.Utils) {

			defer wg.Done()

			var biasWg sync.WaitGroup

			if len(d.Bias) != 0 {

				biasWg.Add(1)

				// Calculate updated bias concurrently
				go func(biasUtils utility.Utils) {

					defer biasWg.Done()

					if bootstrapGradient {

						if gradient.BiasGradient[nodeIndex].Level() > 2{
							utils.Evaluator.DropLevel(gradient.BiasGradient[nodeIndex], gradient.BiasGradient[nodeIndex].Level() - 2)
						}

						biasUtils.SumElementsInPlace(gradient.BiasGradient[nodeIndex])
						
						// Generate averager
						encoder := ckks.NewEncoder(biasUtils.Params)
						plain := make([]complex128, biasUtils.Params.Slots())
						plain[nodeIndex] = complex(avgScale, 0)
						avg := encoder.EncodeNTTAtLvlNew(gradient.BiasGradient[nodeIndex].Level(), plain, biasUtils.Params.LogSlots())

						// Multiply with averager filter and sum for parallel bootstrapping
						biasUtils.MultiplyPlain(gradient.BiasGradient[nodeIndex], avg, gradient.BiasGradient[nodeIndex], true, false)
						biasToBootstrap.Add(gradient.BiasGradient[nodeIndex], biasUtils)
						counter.Increment()

					} else {

						if gradient.BiasGradient[nodeIndex].Level() > d.weightLevel + 1 {
							biasUtils.Evaluator.DropLevel(gradient.BiasGradient[nodeIndex], gradient.BiasGradient[nodeIndex].Level() - (d.weightLevel + 1))
						}

						biasUtils.SumElementsInPlace(gradient.BiasGradient[nodeIndex])
						averagedLrBias := biasUtils.MultiplyPlainNew(gradient.BiasGradient[nodeIndex], batchAverager, false, false)
						biasUtils.Sub(d.Bias[nodeIndex], averagedLrBias, d.Bias[nodeIndex])
						counter.Increment()
					}

				}(utils.CopyWithClonedEval().CopyWithClonedEncoder())

			}

			var weightWg sync.WaitGroup

			for w := range d.Weights[nodeIndex] {

				weightWg.Add(1)

				// Update weight concurrently
				go func(weightIndex int, weightUtils utility.Utils) {

					defer weightWg.Done()
					
					// Bootstrap if gradient's level is lower than it's suppose to be
					if bootstrapGradient {

						rescale := true
						if gradient.WeightGradient[nodeIndex][weightIndex].Level() < 2 {
							rescale = false
						}

						// Drop level to make computation faster if possible
						if gradient.WeightGradient[nodeIndex][weightIndex].Level() > 2{
							utils.Evaluator.DropLevel(gradient.WeightGradient[nodeIndex][weightIndex], gradient.WeightGradient[nodeIndex][weightIndex].Level() - 2)
						}

						weightUtils.SumElementsInPlace(gradient.WeightGradient[nodeIndex][weightIndex])

						// get the index of ciphertext in weight parallel bootstrapper, and get the position that this should be in
						ctIndex := (nodeIndex * d.InputUnit) + weightIndex
						ct := int(math.Floor(float64(ctIndex) / float64(d.utils.Params.Slots())))
						ctIndex %= d.utils.Params.Slots()

						// Generate avg filter
						encoder := ckks.NewEncoder(weightUtils.Params)
						plain := make([]complex128, weightUtils.Params.Slots())
						plain[ctIndex] = complex(avgScale, 0)
						avg := encoder.EncodeNTTAtLvlNew(gradient.WeightGradient[nodeIndex][weightIndex].Level(), plain, weightUtils.Params.LogSlots())

						weightUtils.MultiplyPlain(gradient.WeightGradient[nodeIndex][weightIndex], avg, gradient.WeightGradient[nodeIndex][weightIndex], rescale, false)
						weightToBootstrap[ct].Add(gradient.WeightGradient[nodeIndex][weightIndex], weightUtils)
						counter.Increment()

					} else {

						// Drop level to minimum requirement + 1 for faster evaluation
						if gradient.WeightGradient[nodeIndex][weightIndex].Level() > d.weightLevel + 1{
							weightUtils.Evaluator.DropLevel(gradient.WeightGradient[nodeIndex][weightIndex], gradient.WeightGradient[nodeIndex][weightIndex].Level() - (d.weightLevel + 1))
						}

						// DEBUG start
						if nodeIndex == 7 && weightIndex == 10 && d.InputUnit == 20{
							weightUtils = weightUtils.CopyWithClonedDecryptor()
							fmt.Printf("\nSGD Sample pre-sum: %f\n", weightUtils.Decrypt(gradient.WeightGradient[nodeIndex][weightIndex])[0:25])
						}
						// DEBUG end

						// Perform Ciphertext inner sum and average
						weightUtils.SumElementsInPlace(gradient.WeightGradient[nodeIndex][weightIndex])

						// DEBUG start
						if nodeIndex == 7 && weightIndex == 10 && d.InputUnit == 20{
							weightUtils = weightUtils.CopyWithClonedDecryptor()
							fmt.Printf("\nSGD Sample post-sum: %f\n", weightUtils.Decrypt(gradient.WeightGradient[nodeIndex][weightIndex])[0:5])
						}
						if nodeIndex == 0 && weightIndex == 500{
							fmt.Printf("\nSGD weight gradient scale: %f\n", gradient.WeightGradient[nodeIndex][weightIndex].Scale)
						}
						// DEBUG end

						// Multiply with average scale
						weightUtils.MultiplyPlain(gradient.WeightGradient[nodeIndex][weightIndex], batchAverager, gradient.WeightGradient[nodeIndex][weightIndex], false, false)

						// Ensure same scale
						if gradient.WeightGradient[nodeIndex][weightIndex].Level() > d.Weights[nodeIndex][weightIndex].Level() && gradient.WeightGradient[nodeIndex][weightIndex].Scale > d.utils.Scale{

							idealRescaleScale := (d.utils.Scale * math.Pow(2, 40))
							if gradient.WeightGradient[nodeIndex][weightIndex].Scale > idealRescaleScale{
								utils.Evaluator.Rescale(gradient.WeightGradient[nodeIndex][weightIndex], d.utils.Scale, gradient.WeightGradient[nodeIndex][weightIndex])
							} else {
								scaleUpBy := idealRescaleScale / gradient.WeightGradient[nodeIndex][weightIndex].Scale
								utils.Evaluator.ScaleUp(gradient.WeightGradient[nodeIndex][weightIndex], scaleUpBy, gradient.WeightGradient[nodeIndex][weightIndex])
								utils.Evaluator.Rescale(gradient.WeightGradient[nodeIndex][weightIndex], d.utils.Scale, gradient.WeightGradient[nodeIndex][weightIndex])
							}

						}

						// DEBUG start
						weightData := []float64{0, 0, 0}
						gradData := []float64{0, 0, 0}
						if nodeIndex == 7 && weightIndex == 10 && d.InputUnit == 20{
							weightData = []float64{weightUtils.Decrypt(d.Weights[nodeIndex][weightIndex])[0], float64(d.Weights[nodeIndex][weightIndex].Level()), d.Weights[nodeIndex][weightIndex].Scale}
							gradData = []float64{weightUtils.Decrypt(gradient.WeightGradient[nodeIndex][weightIndex])[0], float64(gradient.WeightGradient[nodeIndex][weightIndex].Level()), gradient.WeightGradient[nodeIndex][weightIndex].Scale}
						}
						// DEBUG end

						// Perform SGD
						weightUtils.Sub(d.Weights[nodeIndex][weightIndex], gradient.WeightGradient[nodeIndex][weightIndex], d.Weights[nodeIndex][weightIndex])

						// DEBUG start
						if nodeIndex == 7 && weightIndex == 10 && d.InputUnit == 20{
							res := weightUtils.Decrypt(d.Weights[nodeIndex][weightIndex])[0]
							fmt.Printf("SGD Sample: %f (L: %f, S: %f) - %f (L: %f, S: %f) = %f (L: %f, S: %f)\n", weightData[0], weightData[1], weightData[2], gradData[0], gradData[1], gradData[2], res, float64(d.Weights[nodeIndex][weightIndex].Level()), d.Weights[nodeIndex][weightIndex].Scale)
						}
						// DEBUG end

						counter.Increment()

					}

				}(w, utils.CopyWithClonedEval().CopyWithClonedEncoder())

			}

			weightWg.Wait()
			biasWg.Wait()

		}(node, d.utils.CopyWithClonedEval())

	}

	wg.Wait()
	averageTimer.LogTimeTakenSecond()

	// Check if bootstrap needs to be performed
	if bootstrapGradient {

		fmt.Printf("Bootstrapping SGD gradient (%d)", d.InputUnit)
		btpTimer := logger.StartTimer(fmt.Sprintf("Bootstrapping SGD gradient (%d)", d.InputUnit))

		// Combine all ciphertext into one array of ciphertexts
		biasLen := 0

		if len(d.Bias) != 0{
			biasLen = 1
		}

		cts := make([]*ckks.Ciphertext, len(weightToBootstrap) + biasLen)

		for i := range weightToBootstrap{
			cts[i] = weightToBootstrap[i].Ct
		}

		if len(d.Bias) != 0{
			cts[len(cts)-1] = biasToBootstrap.Ct
		}

		d.utils.Bootstrap1dInPlace(cts, true)

		btpTimer.LogTimeTakenSecond()

		btpTimer = logger.StartTimer("Unpack bootstrapped")

		// Update weight and bias with bootstrapped gradients
		var updateGradWg sync.WaitGroup

		for node := range d.Weights{

			updateGradWg.Add(1)

			go func(nodeIndex int, utils utility.Utils){

				var biasWg sync.WaitGroup

				if len(d.Bias) != 0{
					biasWg.Add(1)

					go func(biasUtils utility.Utils){

						defer biasWg.Done()

						// Generate filter
						encoder := ckks.NewEncoder(biasUtils.Params)
						plain := make([]complex128, biasUtils.Params.Slots())
						plain[nodeIndex] = complex(1, 0)
						filter := encoder.EncodeNTTAtLvlNew(cts[len(cts) - 1].Level(), plain, biasUtils.Params.LogSlots())

						rescale := true
						if cts[len(cts) - 1].Level() > d.weightLevel + 1{
							biasUtils.Evaluator.DropLevel(cts[len(cts) - 1], cts[len(cts) - 1].Level() - (d.weightLevel + 1))
						} else if cts[len(cts) - 1].Level() == d.weightLevel{
							rescale = false
						}

						biasGradient := biasUtils.MultiplyPlainNew(cts[len(cts) - 1], filter, rescale, true)
						utils.Rotate(biasGradient, nodeIndex)
						biasUtils.FillCiphertextInPlace(biasGradient, d.batchSize)
						biasUtils.Sub(d.Bias[nodeIndex], biasGradient, d.Bias[nodeIndex])

					}(utils.CopyWithClonedEval())

				}

				var weightWg sync.WaitGroup

				for w := range d.Weights[nodeIndex]{

					weightWg.Add(1)
					go func(weightIndex int, weightUtils utility.Utils){

						defer weightWg.Done()

						// get the index of ciphertext in weight parallel bootstrapper, and get the position that this should be in
						ctIndex := (nodeIndex * d.InputUnit) + weightIndex
						ct := int(math.Floor(float64(ctIndex) / float64(d.utils.Params.Slots())))
						ctIndex %= d.utils.Params.Slots()

						rescale := true
						if cts[ct].Level() > d.weightLevel + 1{
							weightUtils.Evaluator.DropLevel(cts[ct], cts[ct].Level() - (d.weightLevel + 1))
						} else if cts[ct].Level() == d.weightLevel || cts[ct].Scale > math.Pow(2,50){
							rescale = false
						}

						// Generate filter
						encoder := ckks.NewEncoder(weightUtils.Params)
						plain := make([]complex128, weightUtils.Params.Slots())
						plain[ctIndex] = complex(1, 0)
						filter := encoder.EncodeNTTAtLvlNew(cts[ct].Level(), plain, weightUtils.Params.LogSlots())

						// Isolate weight grqadient
						weightGradient := weightUtils.MultiplyPlainNew(cts[ct], filter, rescale, false)
						weightUtils.SumElementsInPlace(weightGradient)
						weightUtils.Sub(d.Weights[nodeIndex][weightIndex], weightGradient, d.Weights[nodeIndex][weightIndex])

					}(w, utils.CopyWithClonedEval())

				}

				biasWg.Wait()
				weightWg.Wait()

			}(node, d.utils.CopyWithClonedEval())

		}

		updateGradWg.Wait()
		btpTimer.LogTimeTakenSecond()
		
	}

}

func (d Dense) GetOutputSize() int {
	return d.OutputUnit
}

func (d Dense) IsTrainable() bool {
	return true
}

func (d Dense) HasActivation() bool {
	return d.Activation != nil
}

func (d Dense) GetForwardLevelConsumption() int {
	return 1
}

func (d Dense) GetBackwardLevelConsumption() int {
	return 1
}

func (d Dense) GetForwardActivationLevelConsumption() int {
	if d.HasActivation() {
		return (*d.Activation).GetForwardLevelConsumption()
	} else {
		return 0
	}
}

func (d Dense) GetBackwardActivationLevelConsumption() int {
	if d.HasActivation() {
		return (*d.Activation).GetBackwardLevelConsumption()
	} else {
		return 0
	}
}

func (d *Dense) SetBootstrapOutput(set bool, direction string) {
	switch direction {
	case "forward":
		d.btspOutput[0] = set
	case "backward":
		d.btspOutput[1] = set
	}
}

func (d *Dense) SetBootstrapActivation(set bool, direction string) {
	switch direction {
	case "forward":
		d.btspActivation[0] = set
	case "backward":
		d.btspActivation[1] = set
	}
}

func (d *Dense) SetWeightLevel(lvl int) {
	d.weightLevel = lvl
}

type denseWeight struct {
	Weight [][]float64
	Bias   []float64
}

func (d *Dense) LoadWeights(filename string, weightLevel int){
	
	jsonFile, _ := os.Open(filename)
	defer jsonFile.Close()
	file, _ := ioutil.ReadAll(jsonFile)

	var data denseWeight
	json.Unmarshal([]byte(file), &data)

	counter := logger.NewOperationsCounter("Load weight", (d.InputUnit * d.OutputUnit) + d.OutputUnit)

	var wg sync.WaitGroup

	for node := range data.Weight {

		wg.Add(1)

		go func(nodeIndex int, nodeUtils utility.Utils){

			defer wg.Done()

			var weightWg sync.WaitGroup

			for w := range data.Weight[nodeIndex] {

				weightWg.Add(1)

				go func(weightIndex int, weightUtils utility.Utils){
					defer weightWg.Done()
					d.Weights[nodeIndex][weightIndex] = weightUtils.EncryptToLevel(weightUtils.GenerateFilledArraySize(data.Weight[nodeIndex][weightIndex], d.batchSize), weightLevel)
					counter.Increment()
				}(w, nodeUtils.CopyWithClonedEncryptor())
	
			}
	
			d.Bias[nodeIndex] = nodeUtils.EncryptToLevel(nodeUtils.GenerateFilledArraySize(data.Bias[nodeIndex], d.batchSize), weightLevel)
			counter.Increment()

			weightWg.Wait()

		}(node, d.utils.CopyWithClonedEncryptor())

	}

	wg.Wait()

}

func (d *Dense) ExportWeights(filename string) {

	plainWeights := make([][]float64, len(d.Weights))

	var wg sync.WaitGroup

	for node := range d.Weights {
		plainWeights[node] = make([]float64, len(d.Weights[node]))
		for i := range plainWeights[node] {
			wg.Add(1)
			go func(nodeIdx int, weightIdx int, utils utility.Utils) {
				defer wg.Done()
				plainWeights[nodeIdx][weightIdx] = utils.Decrypt(d.Weights[nodeIdx][weightIdx])[0]
			}(node, i, d.utils.CopyWithClonedDecryptor())
		}
	}

	wg.Wait()

	bias := make([]float64, len(d.Bias))

	for node := range bias {
		bias[node] = d.utils.Decrypt(d.Bias[node])[0]
	}

	weight := denseWeight{plainWeights, bias}

	file, _ := json.MarshalIndent(weight, "", " ")

	_ = ioutil.WriteFile(filename, file, 0644)

}
