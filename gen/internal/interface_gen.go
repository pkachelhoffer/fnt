package internal

import (
	"context"

	"github.com/pkachelhoffer/fnt/gen/internal/packone/packmain"

	packmain_1 "github.com/pkachelhoffer/fnt/gen/internal/packtwo/packmain"
)

type (
	PerformRequest func(context.Context, Request, int, string) (Response, error)
	InterfaceParam func(context.Context, performer)
	Alias          func(context.Context, packmain.PackItem, packmain_1.PackItem) (packmain.PackItem, packmain_1.PackItem)
	Pointers       func(*Request, *int) *Response
	Arrays         func([]*Request, []Request, []int) []int
	Maps           func(map[int]*Response, map[int]Response, map[packmain.PackItem]packmain_1.PackItem) map[string]int
)
