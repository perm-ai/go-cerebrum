package utility

import (
	"math"

	// "github.com/ldsec/lattigo/v2/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"

)

//=================================================
//		CIPHERTEXT DATA STRUCT FOR GROUPING
//		  CIPHERTEXTS INTO ONE CIPHERTEXT
//=================================================

type CiphertextData struct {

	Ciphertext 		rlwe.Ciphertext
	Length			int
	start			int
	end 			int
	groupIndex		int

}

func NewCiphertextData (ct rlwe.Ciphertext, length int) CiphertextData {

	return CiphertextData{Ciphertext: ct, Length: length}

}

func (c *CiphertextData) setStart(start int){
	c.start = start
}

func (c *CiphertextData) setEnd(end int){
	c.end = end
}

func (c *CiphertextData) setGroupIndex(group int){
	c.groupIndex = group
}


//=================================================
//				CIPHERTEXT GROUP
//=================================================

type CiphertextGroup struct {

	CiphertextGroups 	[]rlwe.Ciphertext
	ciphertexts			[]CiphertextData
	utils				Utils

}

func NewCiphertextGroup(ciphertexts []CiphertextData, utils Utils) CiphertextGroup {

	combinedCiphertext := []rlwe.Ciphertext{}
	availableRotations := make([]int, utils.Params.LogSlots() + 1)

	for i := 0; i <= utils.Params.LogSlots(); i++ {
		availableRotations[i] = int(math.Pow(2, float64(i)))
	}

	groupIndex := 0
	start := 0

	for ctIndex := range ciphertexts{

		// Find the power of 2 ceiling of ciphertext length
		var pow2Ceil int

		if ciphertexts[ctIndex].Length == 1 {
			pow2Ceil = 1
		} else if ciphertexts[ctIndex].Length > utils.Params.Slots(){
			panic("Invalid length")
		} else {
			for i, rotation := range availableRotations[1:]{
			
				if rotation == ciphertexts[ctIndex].Length || (availableRotations[i - 1] < ciphertexts[ctIndex].Length && ciphertexts[ctIndex].Length < availableRotations[i]) {
					pow2Ceil = rotation
					break
				}
	
			}
		}

		if start + pow2Ceil > utils.Params.Slots() {
			groupIndex++
			start = 0
			combinedCiphertext = append(combinedCiphertext, ciphertexts[ctIndex].Ciphertext)
		} else {
			if start == 0{
				combinedCiphertext = append(combinedCiphertext, ciphertexts[ctIndex].Ciphertext)
			} else {
				rotated := utils.RotateNew(&ciphertexts[ctIndex].Ciphertext, (-1 * pow2Ceil))
				utils.Add(&combinedCiphertext[groupIndex], &rotated, &combinedCiphertext[groupIndex])
			}
		}

		ciphertexts[ctIndex].setStart(start)
		ciphertexts[ctIndex].setEnd(start + pow2Ceil)
		ciphertexts[ctIndex].setGroupIndex(groupIndex)

		start += pow2Ceil

	}

	return CiphertextGroup{combinedCiphertext, ciphertexts, utils}

}

func (c *CiphertextGroup) Bootstrap(){

	for i := range c.CiphertextGroups{
		c.utils.BootstrapInPlace(&c.CiphertextGroups[i])
	}

}

func (c CiphertextGroup) BreakGroup(rescale bool) []*rlwe.Ciphertext {

	brokenCiphertexts := make([]*rlwe.Ciphertext, len(c.ciphertexts))

	for i, ciphertext := range c.ciphertexts{

		filter := make([]float64, c.utils.Params.Slots())

		for f := ciphertext.start; f < ciphertext.end; f++ {
			filter[f] = 1
		}
		
		encodedFilter := c.utils.Encoder.EncodeNew(c.utils.Float64ToComplex128(filter), c.utils.Params.MaxLevel(), c.utils.Params.DefaultScale(), c.utils.Params.LogSlots())
		brokenCiphertexts[i] = c.utils.MultiplyPlainNew(&c.CiphertextGroups[ciphertext.groupIndex], encodedFilter, rescale, false)

	}

	return brokenCiphertexts

}