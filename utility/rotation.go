package utility

import (
	"math"
	"sort"

	"github.com/ldsec/lattigo/v2/ckks"
)

func (u Utils) Rotate(ct *ckks.Ciphertext, k int) {

	evaluator := u.Get2PowRotationEvaluator()

	availableSteps := []int{}

	for i := 0; i <= u.Params.LogSlots(); i++ {
		positive := int(math.Pow(2, float64(i)))
		availableSteps = append(availableSteps, positive)
		availableSteps = append(availableSteps, (-1 * positive))
	}

	sort.Ints(availableSteps[:])

	steps := findStep(k, 0, []int{}, availableSteps)

	for _, step := range steps {
		evaluator.Rotate(ct, step, ct)
	}

}

func (u Utils) RotateNew(ct *ckks.Ciphertext, k int) ckks.Ciphertext {

	newCt := ct.CopyNew()
	u.Rotate(newCt, k)

	return *newCt

}

func findStep(target int, stepSum int, steps []int, availableSteps []int) []int {

	dif := target - stepSum

	distanceToDif := make([]int, len(availableSteps))

	for i, n := range availableSteps {
		distanceToDif[i] = int(math.Abs(float64(n - dif)))

		if i != 0 {

			if distanceToDif[i] > distanceToDif[i-1] {
				steps = append(steps, availableSteps[i-1])
				break
			} else if distanceToDif[i] == 0 || distanceToDif[i] == distanceToDif[i-1] || i == len(availableSteps)-1 {
				steps = append(steps, n)
				break
			}

		}

	}

	newSum := 0

	for _, n := range steps {
		newSum += n
	}

	if newSum == target {
		return steps
	} else {
		return findStep(target, newSum, steps, availableSteps)
	}

}

// This function will rotate and add ciphertext to fill the ciphertext to a certain slot with number at index 1
func (u Utils) FillCiphertextInPlace(ct *ckks.Ciphertext, slots int) {

	evaluator := u.Get2PowRotationEvaluator()
	current := 0
	cache := make([]*ckks.Ciphertext, u.Params.LogSlots())
	cache[0] = ct
	var result *ckks.Ciphertext

	for current != slots {
		for rot := 0; int(math.Pow(2, float64(rot)))+current <= slots; rot++ {
			if cache[rot] == nil && rot != 0 {

				tmp := evaluator.RotateNew(cache[rot-1], -1*int(math.Pow(2, float64(rot-1))))
				u.Add(tmp, cache[rot-1], tmp)
				cache[rot] = tmp

			}
			if int(math.Pow(2, float64(rot+1)))+current > slots {
				if result == nil {
					result = cache[rot]
				} else {
					tmp := u.RotateNew(cache[rot], -1*current)
					u.Add(&tmp, result, result)
				}
				current += int(math.Pow(2, float64(rot)))
			}
		}
	}

	*ct = *result

}
