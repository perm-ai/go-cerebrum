package main

<<<<<<< Updated upstream
import "github.com/perm-ai/go-cerebrum/models"

func main(){

	models.ModelCreationExample()

}
=======
import (
	"fmt"
	"math"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

func main() {

	var keyChain = key.GenerateKeys(0, true, true)
	var utils = utility.NewUtils(keyChain, math.Pow(2, 35), 100, true)

	a1 := utils.GenerateFilledArray(1)
	a2 := utils.GenerateFilledArray(2)
	a3 := utils.GenerateFilledArray(3)
	a4 := utils.GenerateFilledArray(4)

	a1ct := utils.Encrypt(a1)
	a2ct := utils.Encrypt(a2)
	a3ct := utils.Encrypt(a3)
	a4ct := utils.Encrypt(a4)

	a := make([]ckks.Ciphertext, 4)
	a[0] = a1ct
	a[1] = a2ct
	a[2] = a3ct
	a[3] = a4ct

	b1 := utils.GenerateFilledArray(1)
	b2 := utils.GenerateFilledArray(2)
	b3 := utils.GenerateFilledArray(3)
	b4 := utils.GenerateFilledArray(4)

	b1ct := utils.Encrypt(b1)
	b2ct := utils.Encrypt(b2)
	b3ct := utils.Encrypt(b3)
	b4ct := utils.Encrypt(b4)

	b := make([]ckks.Ciphertext, 4)
	b[0] = b1ct
	b[1] = b2ct
	b[2] = b3ct
	b[3] = b4ct

	timer1 := logger.StartTimer("nonConcurrency")
	nonConcur := utils.InterDotProduct(a, b, true, false, false)
	timer1.LogTimeTaken()

	timer2 := logger.StartTimer("Concurrency")
	Concur := utils.InterDotProduct(a, b, true, false, true)
	timer2.LogTimeTaken()

	plainNonConcur := utils.Decrypt(&nonConcur)
	plainConcur := utils.Decrypt(&Concur)

	fmt.Printf("The result for Non concurrency is %f\n", plainNonConcur[0:3])
	fmt.Printf("The result for concurrency is %f\n", plainConcur[0:3])
}
>>>>>>> Stashed changes
