package main

import (
	"flag"
	"github.com/dowenliu-xyz/Go-000/Week04/api/crm/customer_follow/v1"
	"github.com/dowenliu-xyz/Go-000/Week04/internal/customer_follow/di"
	xapp "github.com/dowenliu-xyz/Go-000/Week04/internal/pkg/app"
	"github.com/dowenliu-xyz/Go-000/Week04/internal/pkg/server/http"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/application.yaml", "config file")
}

func main() {
	flag.Parse()
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	customerFollowHttpService, cleanupFn, err := di.InitHttpService(config.Database)
	if err != nil {
		if cleanupFn != nil {
			cleanupFn()
		}
		log.Fatalf("组装失败L %v", err)
	}
	defer cleanupFn()
	conv := customer_follow.NewCustomerFollowServiceHTTPConverter(customerFollowHttpService)
	server := http.NewServer(config.Server.Addr)
	server.HandleServiceWithName(conv.WriteFollowWithName)
	server.HandleServiceWithName(conv.ListFollowWithName)
	app := xapp.New()
	app.Append(xapp.Hook{OnStart: server.Start, OnStop: server.Stop})
	if err = app.Run(); err != nil {
		log.Printf("%+v\n", err)
	}
}

func loadConfig() (*di.Config, error) {
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, errors.Wrapf(err, "读取配置文件 %s 失败: %v", configFile, err)
	}
	config := &di.Config{}
	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return nil, errors.Wrapf(err, "读取YAML配置失败: %v", err)
	}
	return config, nil
}
