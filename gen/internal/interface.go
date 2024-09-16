package internal

import (
	"context"
	"fnt/gen/internal/packone/packmain"
	p2 "fnt/gen/internal/packtwo/packmain"
)

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
