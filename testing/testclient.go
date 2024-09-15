package testing

import (
	"context"
	"fnt/testing/packone/packmain"
	p2 "fnt/testing/packone/packmain"
)

//go:generate main

type TestInterface interface {
	PerformRequest(ctx context.Context, req Request, val1 int, val2 string) (Response, error)
	InterfaceParam(ctx context.Context, perf performer)
	Alias(ctx context.Context, pack1 packmain.PackItem, pack2 p2.PackItem) (packmain.PackItem, p2.PackItem)
}

type Request struct {
}

type Response struct {
}

type performer interface {
	DoSomething()
}

type TestInterfaceImplementation struct {
}

func (t TestInterfaceImplementation) PerformRequest(ctx context.Context, req Request, val1 int, val2 string) (Response, error) {
	panic("implement me")
}

func (t TestInterfaceImplementation) InterfaceParam(ctx context.Context, perf performer) {
	panic("implement me")
}

func (t TestInterfaceImplementation) Alias(ctx context.Context, pack1 packmain.PackItem, pack2 p2.PackItem) (packmain.PackItem, p2.PackItem) {
	panic("implement me")
}

func DoSomething() {
	imp := TestInterfaceImplementation{}
	processTestInterface(imp.PerformRequest, imp.InterfaceParam, imp.Alias)
}

func processTestInterface(fnPerformRequest PerformRequest, fnInterfaceParam InterfaceParam, fnAlias Alias) {

}
