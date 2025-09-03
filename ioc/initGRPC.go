package ioc

import (
	"context"
	"github.com/go-kratos/aegis/circuitbreaker/sre"
	"github.com/spf13/viper"
	"github.com/webook-project-go/webook-pkgs/grpcx"
	"github.com/webook-project-go/webook-pkgs/grpcx/interceptor"
	"github.com/webook-project-go/webook-pkgs/grpcx/prometheus"
	"github.com/webook-project-go/webook-pkgs/grpcx/trace"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
)

func InitGrpcServer(client *clientv3.Client) *grpcx.GrpcxServer {
	type Config struct {
		Addr     string `yaml:"addr"`
		Protocol string `yaml:"protocol"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}

	breaker := sre.NewBreaker()
	bk := interceptor.NewSreInterceptor(breaker)

	pro := prometheus.NewPrometheusBuilder()

	tracer := trace.NewTracerBuilder(otel.GetTracerProvider().
		Tracer("webook-service-active"), otel.GetTextMapPropagator())

	gc := grpc.NewServer(
		grpc.UnaryInterceptor(ChainUnaryServer(pro.BuildUnaryInterceptor(),
			bk.UnaryInterceptor(),
			tracer.BuildServer())),
	)
	server := grpcx.NewGrpcxServer(client, gc, grpcx.Config{
		TTL:      30,
		Name:     "active",
		Addr:     cfg.Addr,
		Protocol: cfg.Protocol,
	})
	return server
}
func ChainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		chainedHandler := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			current := interceptors[i]
			next := chainedHandler
			chainedHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
				return current(ctx, req, info, next)
			}
		}
		return chainedHandler(ctx, req)
	}
}
