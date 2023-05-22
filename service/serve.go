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
	log.Info("Starting server...")
	openai_key := viper.GetString("openai.api_key")
	notion_auth := viper.GetString("notion.api_auth")
	transbot = translator.CreateTranslator(openai_key, notion_auth)

	// ruter
	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// path
	group := router.Group("/v1/")
	group.GET("/translate/:pageuuid/:language", TranslatePage)

	port := fmt.Sprintf(":%s", viper.GetString("service.port"))
	if viper.GetBool("service.tls") {
		key_file := viper.GetString("service.tls_key")
		cert_file := viper.GetString("service.tls_cert")
		err := router.RunTLS(port, cert_file, key_file)
		if err != nil {
			log.Fatal("run tls service error: ", err.Error())
		}
	} else {
		err := router.Run(port)
		if err != nil {
			log.Fatal("run service error: ", err.Error())
		}
	}

}
