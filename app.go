package main

import (
	"github.com/webook-project-go/webook-active/events"
	"github.com/webook-project-go/webook-active/grpc"
	"github.com/webook-project-go/webook-pkgs/grpcx"
)

type App struct {
	Server    *grpcx.GrpcxServer
	Service   *grpc.Service
	Consumers []events.Consumer
}
