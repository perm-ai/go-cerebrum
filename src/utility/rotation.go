package utility

import (
	"math"
	"sort"

	"github.com/ldsec/lattigo/v2/ckks"
)

func (u Utils) Rotate (ct *ckks.Ciphertext, k int){

	evaluator := u.Get2PowRotationEvaluator()

	availableSteps := []int{}

	for i := 0; i <= u.Params.LogSlots(); i++{
		positive := int(math.Pow(2, float64(i)))
		availableSteps = append(availableSteps, positive)
		availableSteps = append(availableSteps, (-1 * positive))
	}

	sort.Ints(availableSteps[:])

	steps := findStep(k, 0, []int{}, availableSteps)

	evaluator.RotateHoisted(ct, steps)

}

func (u Utils) RotateNew (ct *ckks.Ciphertext, k int) ckks.Ciphertext {

	newCt := ct.CopyNew()
	u.Rotate(newCt, k)

	return *newCt

}

func findStep(target int, stepSum int, steps []int, availableSteps []int) []int {

	dif := target - stepSum

	distanceToDif := make([]int, len(availableSteps))

	for i, n := range availableSteps{

		distanceToDif[i] = int(math.Abs(float64(n - dif)))

		if i != 0{
			if distanceToDif[i] > distanceToDif[i - 1]{
				steps = append(steps, distanceToDif[i - 1])
				break
			} else if distanceToDif[i] == distanceToDif[i - 1] {
				steps = append(steps, distanceToDif[i])
				break
			}
		}

	}

	newSum := 0

	for _, n := range steps{
		newSum += n
	}

	if(newSum == target){
		return steps;
	} else {
		return findStep(target, newSum, steps, availableSteps)
	}

}