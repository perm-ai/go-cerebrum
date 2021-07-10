package ml

import (
	"math"

	"github.com/perm-ai/GO-HEML-prototype/src/logger"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)

var log = logger.NewLogger(true)
var utils = utility.NewUtils(math.Pow(2, 35), 100, false, true)