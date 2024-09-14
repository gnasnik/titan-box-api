package main

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-box-api/api"
	"github.com/gnasnik/titan-box-api/config"
	"github.com/gnasnik/titan-box-api/core/dao"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	OsSignal := make(chan os.Signal, 1)

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("reading config file: %v\n", err)
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("unmarshaling config file: %v\n", err)
	}

	config.Cfg = cfg

	if err := dao.Init(&cfg); err != nil {
		log.Fatalf("initital: %v\n", err)
	}

	go api.ServerAPI(&cfg)

	ds := api.NewDataService()
	go ds.Run(context.Background())

	signal.Notify(OsSignal, syscall.SIGINT, syscall.SIGTERM)
	_ = <-OsSignal

	fmt.Printf("Exiting received OsSignal\n")
}
