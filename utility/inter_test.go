package utility

import (
	"fmt"
	"testing"

	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/logger"
)

func TestInter(t *testing.T) {

	a1 := utils.GenerateFilledArray(1)
	a2 := utils.GenerateFilledArray(2)
	a3 := utils.GenerateFilledArray(3)
	a4 := utils.GenerateFilledArray(4)

	a1ct := utils.Encrypt(a1)
	a2ct := utils.Encrypt(a2)
	a3ct := utils.Encrypt(a3)
	a4ct := utils.Encrypt(a4)

	a := make([]*rlwe.Ciphertext, 4)
	*a[0] = a1ct
	*a[1] = a2ct
	*a[2] = a3ct
	*a[3] = a4ct

	b1 := utils.GenerateFilledArray(1)
	b2 := utils.GenerateFilledArray(2)
	b3 := utils.GenerateFilledArray(3)
	b4 := utils.GenerateFilledArray(4)

	b1ct := utils.Encrypt(b1)
	b2ct := utils.Encrypt(b2)
	b3ct := utils.Encrypt(b3)
	b4ct := utils.Encrypt(b4)

	b := make([]*rlwe.Ciphertext, 4)
	*b[0] = b1ct
	*b[1] = b2ct
	*b[2] = b3ct
	*b[3] = b4ct

	timer1 := logger.StartTimer("nonConcurrency")
	nonConcur := utils.InterDotProduct(a, b, true, false, nil)
	timer1.LogTimeTaken()

	timer2 := logger.StartTimer("Concurrency")
	Concur := utils.InterDotProduct(a, b, true, true, nil)
	timer2.LogTimeTaken()

	plainNonConcur := utils.Decrypt(nonConcur)
	plainConcur := utils.Decrypt(Concur)

	fmt.Printf("The result for Non concurrency is %f\n", plainNonConcur)
	fmt.Printf("The result for concurrency is %f\n", plainConcur)

}
