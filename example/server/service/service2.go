package service

import (
	"context"
	"time"
)

// @rpcxMethod
func (t *Arith) MulDouble(d int, ctx context.Context) []int {
	return []int{d, d}
}

// @rpcxMethod
func NoResult() {
	println("call fun NoResult")
}

// @rpcxMethod.go
func GoFunc() {
	time.Sleep(10 * time.Second)
	println("call fun GoFunc")
}

// @rpcxMethod
func ContextFunc(ctx context.Context) {
	println("context.Context")
}

// @rpcxMethod
func OneParamFunc(x bool) {
	if x {
		println("OneParamFunc true")
	} else {
		println("OneParamFunc false")
	}
}
