package syncing

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-box-api/api"
	"github.com/gnasnik/titan-box-api/config"
	"github.com/gnasnik/titan-box-api/core/dao"
	"github.com/spf13/viper"
	"log"
)

func main() {
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
		if userKey.Status != 0 {
			continue
		}

		fmt.Printf("start syncing %s\n", userKey.Username)

		ds.StartSyncBoxHistory(userKey)

		ds.StartSyncBoxHistory(userKey)
	}
}
