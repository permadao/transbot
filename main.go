package main

import (
	"fmt"

	"github.com/permadao/transbot/notionopt"
	"github.com/permadao/transbot/service"
	"github.com/permadao/transbot/translator"
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

func TranslateTest() {
	t := translator.CreateTranslator(viper.GetString("openai.api_key"), viper.GetString("notion.api_auth"))
	n := t.NotionClient

	// get notion page
	pageContent, err := n.FetchPage("d77601f7a3e649b7967f61a4462fad53")
	if err != nil {
		fmt.Println("fetch page error: ", err)
		return
	}
	//fmt.Println("JSON page: ", pageContent)

	// convert string content to struct blocks
	page, err := notionopt.Content2NotionPage(pageContent)
	if err != nil {
		fmt.Println("convert block error: ", err)
		return
	}

	// translate content
	for _, block := range page.PageContent.Results {
		toTrans, err := notionopt.GetBlockContent(block)
		if err != nil {
			fmt.Println("GetBlockContent error: ", err)
			return
		}
		if toTrans != "" {
			traned, err := t.Translate(toTrans, "english")
			if err != nil {
				fmt.Println("Translate error: ", err)
				return
			}
			//traned := "Replaced: " + toTrans
			err = notionopt.ReplaceBlockContent(block, traned)
			if err != nil {
				fmt.Println("ReplaceBlockContent error: ", err)
				return
			}
		}
	}

	// upload new page
	uuid, err := n.UploadPage(page.PageInfo.ID, page)
	if err != nil {
		fmt.Println("UploadPage error: ", err)
		return
	}
	fmt.Println("Sucess! uuid: ", uuid)
}
