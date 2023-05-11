package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/permadao/transbot/translator"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var transbot *translator.Translator

func StartServe() {
	router := gin.Default()
	router.Use(gin.Recovery())

	log.Info("Starting server...")
	openai_key := viper.GetString("openai.api_key")
	notion_auth := viper.GetString("notion.api_auth")
	transbot = translator.CreateTranslator(openai_key, notion_auth)

	// path
	group := router.Group("/v1/")
	group.GET("/translate/:pageuuid/:language", TranslatePage)

	port := fmt.Sprintf(":%s", viper.GetString("service.port"))
	router.Run(port)
}
