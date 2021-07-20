package ml

import (
	"math"

	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

// This file is only for global varaiable declaration for ml package tests
// To create tests please go to or create a new file for that functions' category
// eg. test activation in activation_test.go

var log = logger.NewLogger(true)
var utils = utility.NewUtils(math.Pow(2, 35), 100, false, true)
