package testing

import (
	"context"

	"fnt/testing/packone/packmain"
)

type (
	PerformRequest func(context.Context, Request, int, string) (Response, error)
	InterfaceParam func(context.Context, performer)
	Alias          func(context.Context, packmain.PackItem, packmain.PackItem) (packmain.PackItem, packmain.PackItem)
)
