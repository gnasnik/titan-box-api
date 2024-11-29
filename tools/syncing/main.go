package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gnasnik/titan-box-api/api"
	"github.com/gnasnik/titan-box-api/config"
	"github.com/gnasnik/titan-box-api/core/dao"
	"github.com/spf13/viper"
	"log"
)

func main() {

	var (
		dataType, username, from string
	)

	flag.StringVar(&dataType, "type", "", "syncing the box data, including all, bandwidth, income, quality")
	flag.StringVar(&username, "user", "", "specify the user")
	flag.StringVar(&from, "from", "", "syncing from this day")

	flag.Parse()

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

	ds := api.NewDataService()

	apiKeys, err := dao.GetUserKeys(context.Background())
	if err != nil {
		log.Fatalf("get user keys: %v", err)
	}

	for _, userKey := range apiKeys {
		if userKey.Status == 1 {
			continue
		}

		fmt.Printf("start syncing %s\n", userKey.PaiUsername)

		switch dataType {
		case "all":
			ds.StartSyncBoxList(userKey)
			ds.StartSyncBoxIncomeHistoryFrom(userKey, from)
			ds.StartSyncBoxDayBandwidthHistoryFrom(userKey, from)
			ds.StartSyncBoxDayQualitiesHistoryFrom(userKey, from)
		case "box":
			ds.StartSyncBoxList(userKey)
		case "income":
			ds.StartSyncBoxIncomeHistoryFrom(userKey, from)
		case "bandwidth":
			ds.StartSyncBoxDayBandwidthHistoryFrom(userKey, from)
		case "qualities":
			ds.StartSyncBoxDayQualitiesHistoryFrom(userKey, from)
		default:
			log.Fatalf("unsupport data type")
		}
	}
}
