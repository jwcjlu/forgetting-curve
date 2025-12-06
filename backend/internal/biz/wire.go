package biz

import (
	"forgetting-curve/backend/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewStudentUsecase,
	NewWordUsecase,
	NewWechatServiceProvider,
)
