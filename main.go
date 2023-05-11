package main

import (
	"fmt"

	"github.com/permadao/transbot/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	// Read configs
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("read config failed: %s", err.Error()))
	}

	// Init log
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.DebugLevel)

	// serve
	fmt.Println("=======================")
	fmt.Println("Hello Transbot!")
	fmt.Println("=======================")
	service.StartServe()

}
