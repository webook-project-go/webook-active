//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/webook-project-go/webook-active/grpc"
	"github.com/webook-project-go/webook-active/ioc"
	"github.com/webook-project-go/webook-active/repository/redis"
	"github.com/webook-project-go/webook-active/service"
)

var activeServiceProvider = wire.NewSet(
	service.New,
	redis.New,
)

var thirdPartyProvider = wire.NewSet(
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.InitLogger,
	ioc.InitEtcd,
)

func InitApp() *App {
	wire.Build(
		wire.Struct(new(App), "*"),
		thirdPartyProvider,
		grpc.New,
		activeServiceProvider,
		ioc.InitGrpcServer,
		ioc.InitConsumer,
	)
	return new(App)
}
