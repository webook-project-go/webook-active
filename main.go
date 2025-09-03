package main

import (
	"context"
	_ "github.com/webook-project-go/webook-active/config"
	"github.com/webook-project-go/webook-active/ioc"
	v1 "github.com/webook-project-go/webook-apis/gen/go/apis/active/v1"
)

func main() {
	app := InitApp()
	for _, c := range app.Consumers {
		_ = c.Start()
	}
	shutdwon := ioc.InitOTEL()
	defer shutdwon(context.Background())
	v1.RegisterActiveServiceServer(app.Server, app.Service)
	err := app.Server.Serve()
	if err != nil {
		panic(err)
	}
}
