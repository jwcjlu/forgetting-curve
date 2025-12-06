//go:build wireinject
// +build wireinject

package main

import (
	"forgetting-curve/backend/internal/biz"
	"forgetting-curve/backend/internal/conf"
	"forgetting-curve/backend/internal/data"
	"forgetting-curve/backend/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		newApp,
	))
}
