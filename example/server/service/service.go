package service

import (
	"context"
)

// @rpcxService arithSev
type Arith struct{}

// @rpcxMethod
func (t *Arith) Mul(a1 int, b1 int) int {
	return a1 + b1
}

// @rpcxMethod
func (t *Arith) One(d int) []int {
	return []int{d, d}
}

// @rpcxMethod
func (t *Arith) TestContext(ctx context.Context) {
}

// @rpcxMethod
func ArithMulFun(a1 int, b1 int) int {
	return a1 + b1
}
