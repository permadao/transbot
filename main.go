package main

import (
	"fmt"
	"os"

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
	logfile, err := os.OpenFile("transbot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("open log file error: ", err)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(logfile)
	log.SetLevel(log.DebugLevel)

	// serve
	fmt.Println("=======================")
	fmt.Println("Hello Transbot!")
	fmt.Println("=======================")
	service.StartServe()

}
