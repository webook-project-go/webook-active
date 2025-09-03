package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	cfg := pflag.String("config", "", "")
	pflag.Parse()
	if *cfg == "etcd" {
		viper.SetConfigFile("config/dev.yaml")
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
		etcdSub := viper.Sub("etcd")

		type ETCDConfig struct {
			Endpoint string `yaml:"endpoint"`
			Path     string `yaml:"path"`
		}
		etcd := ETCDConfig{
			Endpoint: "http://localhost:13317",
			Path:     "/webook",
		}
		if etcdSub != nil {
			err := etcdSub.Unmarshal(&etcd)
			if err != nil {
				panic(err)
			}
		}
		err := viper.AddRemoteProvider("etcd3", etcd.Endpoint, etcd.Path)
		if err != nil {
			panic(err)
		}
		err = viper.ReadRemoteConfig()
		if err != nil {
			panic(err)
		}
	} else {
		viper.SetConfigFile("config/dev.yaml")
		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}
	}
}
