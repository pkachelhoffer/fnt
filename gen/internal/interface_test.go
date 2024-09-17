package internal

import (
	"context"
	"github.com/pkachelhoffer/fnt/gen"
	"github.com/pkachelhoffer/fnt/gen/internal/packone/packmain"
	p2 "github.com/pkachelhoffer/fnt/gen/internal/packtwo/packmain"
	"testing"
)

type TestInterfaceImplementation struct {
}

func (t TestInterfaceImplementation) Arrays(reqsPoint []*Request, reqs []Request, numbers []int) []int {
	panic("implement me")
}

func (t TestInterfaceImplementation) Pointers(req *Request, id *int) *Response {
	panic("implement me")
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

var _ = TestInterface(TestInterfaceImplementation{})

func DoSomething() {
	imp := TestInterfaceImplementation{}
	processTestInterface(imp.PerformRequest, imp.InterfaceParam, imp.Alias, imp.Pointers, imp.Arrays)
}

func processTestInterface(fnPerformRequest PerformRequest, fnInterfaceParam InterfaceParam, fnAlias Alias, fnPointers Pointers, fnArrays Arrays) {

}

func TestGetInterface(t *testing.T) {
	err := gen.PerformTypeGeneration("", "TestInterface", "", "")
	if err != nil {
		t.Fatalf("err: %e", err)
	}
}
