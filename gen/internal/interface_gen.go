package internal

import (
	"context"

	"fnt/gen/internal/packone/packmain"

	packmain_1 "fnt/gen/internal/packtwo/packmain"

	packmain_2 "fnt/gen/internal/packtwo/packmain"
)

type (
	PerformRequest func(context.Context, Request, int, string) (Response, error)
	InterfaceParam func(context.Context, performer)
	Alias          func(context.Context, packmain.PackItem, packmain_1.PackItem) (packmain.PackItem, packmain_2.PackItem)
)
