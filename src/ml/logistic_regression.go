package ml

// import (
// 	// "fmt"
// 	// "strconv"

// 	"github.com/ldsec/lattigo/v2/ckks"
// 	// "github.com/perm-ai/GO-HEML-prototype/src/logger"
// 	// "github.com/perm-ai/GO-HEML-prototype/src/utility"
// )

// type LogisticRegression struct {
// 	utils utility.Utils
// 	b0     ckks.Ciphertext
// 	b1     ckks.Ciphertext
// 	b2     ckks.Ciphertext
// }

// type LogisticRegressionGradient struct {
// 	Db0 ckks.Ciphertext
// 	Db1 ckks.Ciphertext
// 	Db2 ckks.Ciphertext
// }

// func NewLogisticRegression(u utility.Utils) LogisticRegression {

// 	zeros := u.GenerateFilledArray(0.0)
// 	b0 := u.Encrypt(zeros)
// 	b1 := u.Encrypt(zeros)
// 	b2 := u.Encrypt(zeros)

// 	return LogisticRegression{u, b0, b1, b2}

// }

// func Evaluate(b0 ckks.Ciphertext, b1 ckks.Ciphertext, b2 ckks.Ciphertext, ) {

// }

// func Sigmoid(x ckks.Ciphertext) ckks.Ciphertext {


// }

// func (model *LogisticRegression) Train(x *ckks.Ciphertext, y *ckks.Ciphertext, target int, learningRate float64, size int, epoch int) {

// }

