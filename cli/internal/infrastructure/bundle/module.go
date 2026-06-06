package bundle

import "go.uber.org/fx"

var Module = fx.Module(
	"bundle",
	fx.Provide(NewClient),
)
