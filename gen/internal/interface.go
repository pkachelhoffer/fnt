package internal

import (
	"context"
	"github.com/pkachelhoffer/fnt/gen/internal/packone/packmain"
	p2 "github.com/pkachelhoffer/fnt/gen/internal/packtwo/packmain"
)

type TestInterface interface {
	PerformRequest(ctx context.Context, req Request, val1 int, val2 string) (Response, error)
	InterfaceParam(ctx context.Context, perf performer)
	Alias(ctx context.Context, pack1 packmain.PackItem, pack2 p2.PackItem) (packmain.PackItem, p2.PackItem)
	Pointers(req *Request, id *int) *Response
	Arrays(reqsPoint []*Request, reqs []Request, numbers []int) []int
}

type Request struct {
}

type Response struct {
}

type performer interface {
	DoSomething()
}
