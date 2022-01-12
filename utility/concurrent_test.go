package utility

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/logger"
)

func TestConcurrentSumElement(t *testing.T) {

	// iter := 7840

	// ct := utils.EncryptToPointer([]float64{1.2, 2.5, 0.9, 5.3})
	// utils.Evaluator.DropLevel(ct, ct.Level()-2)
	// cts := make([]*ckks.Ciphertext, iter)

	// var ctWg sync.WaitGroup
	// for i := range cts{
	// 	ctWg.Add(1)
	// 	go func(index int){
	// 		defer ctWg.Done()
	// 		cts[index] = ct.CopyNew()
	// 	}(i)
	// }
	// ctWg.Wait()

	// var wg sync.WaitGroup

	// fmt.Println("Starting sum")
	// timer := logger.StartTimer("Sum")

	// for i := 0; i < iter; i++ {

	// 	wg.Add(1)
	// 	go func(index int, u Utils){
	// 		defer wg.Done()
	// 		u.SumElementsInPlace(cts[index])
	// 	}(i, utils.CopyWithClonedEval())

	// }

	// wg.Wait()

	// timer.LogTimeTakenSecond()

	// if !ValidateResult(utils.Decrypt(cts[5]), utils.GenerateFilledArray(9.9), false, 1, log) {
	// 	t.Error("Incorrect Sum index 5")
	// }

	// if !ValidateResult(utils.Decrypt(cts[104]), utils.GenerateFilledArray(9.9), false, 1, log) {
	// 	t.Error("Incorrect Sum index 104")
	// }

	// if !ValidateResult(utils.Decrypt(cts[792]), utils.GenerateFilledArray(9.9), false, 1, log) {
	// 	t.Error("Incorrect Sum index 792")
	// }

}

func TestConcurrentBootstrap(t *testing.T) {

	total := 10
	plain := make([][]float64, total)
	cts := make([]*ckks.Ciphertext, total)

	for i := range cts {
		plain[i] = utils.GenerateRandomArray(-10, 10, 1000)
		cts[i] = utils.EncryptToLevel(plain[i], 4)
	}

	utils.Bootstrap1dInPlace(cts, true)

	for i := range cts {

		if !ValidateResult(utils.Decrypt(cts[i]), plain[i], false, 1, log) {
			t.Error(fmt.Sprintf("Incorrect bootstrapping [%d]", i))
		}

		if cts[i].Level() != 9 {
			t.Error(fmt.Sprintf("Incorrect level [%d]", i))
		}

	}

}

func TestConcurrentBenchmark(t *testing.T){

	testSizes := []int{50,100,150,200}

	for _, total := range testSizes{
		plain := make([][]float64, total)
		cts := make([]*ckks.Ciphertext, total)

		for i := range cts {
			plain[i] = utils.GenerateRandomArray(-10, 10, 1000)
			cts[i] = utils.EncryptToLevel(plain[i], 2)
		}

		fmt.Printf("Addition %d\n", total)
		var wg sync.WaitGroup

		timer := logger.StartTimer(fmt.Sprintf("Addition %d", total))
		for i := range cts{

			wg.Add(1)
			go func(index int, u Utils){
				defer wg.Done()
				u.Add(cts[index], cts[index], cts[index])
			}(i, utils.CopyWithClonedEval())

		}

		wg.Wait()
		
		timer.LogTimeTakenSecond()
	}

	for _, total := range testSizes{
		plain := make([][]float64, total)
		cts := make([]*ckks.Ciphertext, total)

		for i := range cts {
			plain[i] = utils.GenerateRandomArray(-10, 10, 1000)
			cts[i] = utils.EncryptToLevel(plain[i], 2)
		}

		fmt.Printf("Multiplication %d\n", total)
		var wg sync.WaitGroup

		timer := logger.StartTimer(fmt.Sprintf("Multiplication %d", total))
		for i := range cts{

			wg.Add(1)
			go func(index int, u Utils){
				defer wg.Done()
				u.Multiply(cts[index], cts[index], cts[index], true, false)
			}(i, utils.CopyWithClonedEval())

		}

		wg.Wait()
		
		timer.LogTimeTakenSecond()
	}

	for _, total := range testSizes{
		plain := make([][]float64, total)
		cts := make([]*ckks.Ciphertext, total)

		for i := range cts {
			plain[i] = utils.GenerateRandomArray(-10, 10, 1000)
			cts[i] = utils.EncryptToLevel(plain[i], 2)
		}

		fmt.Printf("Bootstrapping %d\n", total)
		timer := logger.StartTimer(fmt.Sprintf("Bootstrap %d", total))
		utils.Bootstrap1dInPlace(cts, true)
		timer.LogTimeTakenSecond()

		for i := range cts {

			if !ValidateResult(utils.Decrypt(cts[i]), plain[i], false, 1, log) {
				t.Error(fmt.Sprintf("Incorrect bootstrapping [%d]", i))
			}

			if cts[i].Level() != 9 {
				t.Error(fmt.Sprintf("Incorrect level [%d]", i))
			}

		}
	}
	

}