package ioc

import (
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func InitEtcd() *clientv3.Client {
	addrs := viper.GetStringSlice("etcd.addrs")
	client, err := clientv3.New(clientv3.Config{
		Endpoints: addrs,
	})
	if err != nil {
		panic(err)
	}
	return client
}
