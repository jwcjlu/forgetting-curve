package server

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewHTTPServer,
)
