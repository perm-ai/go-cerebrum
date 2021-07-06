package ml

import (
	// "fmt"
	// "strconv"

	"github.com/ldsec/lattigo/v2/ckks"
	// "github.com/perm-ai/GO-HEML-prototype/src/logger"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)

type LogisticRegression struct {
	utils utility.Utils
	b0     ckks.Ciphertext
	b1     ckks.Ciphertext
	b2     ckks.Ciphertext
}

type LogisticRegressionGradient struct {
	Db0 ckks.Ciphertext
	Db1 ckks.Ciphertext
	Db2 ckks.Ciphertext
}

func NewLogisticRegression(u utility.Utils) LogisticRegression {

	zeros := u.GenerateFilledArray(0.0)
	b0 := u.Encrypt(zeros)
	b1 := u.Encrypt(zeros)
	b2 := u.Encrypt(zeros)

	return LogisticRegression{u, b0, b1, b2}

}

// func Evaluate(b0 ckks.Ciphertext, b1 ckks.Ciphertext, b2 ckks.Ciphertext, ) {

// }

func Sigmoid(x ckks.Ciphertext) ckks.Ciphertext {


	// 0.5 + 0.197x + 0.004x^3 (0.004 * x) * (x * X)


	ans := utils.AddNew(utils.MultiplyNew(x, x, true, false), (utils.MultiplyConstNew(x, utils.GenerateFilledArraySize(0.004, 32768),true ,false))) // ((x * x) * (x * 0.004))
	Add(MultiplyConst(&x, l.utils.GenerateFilledArraySize(0.197, size), &x, true, false), &ans, &ans) // + (x * 0.197)
	AddConst(&ans, 0.5, &ans) // + 0.5

	return ans

}

func (model *LogisticRegression) Train(x *ckks.Ciphertext, y *ckks.Ciphertext, target int, learningRate float64, size int, epoch int) {

}

