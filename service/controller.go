package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/permadao/transbot/notionopt"
	log "github.com/sirupsen/logrus"
)

func TranslatePage(c *gin.Context) {
	uuid := c.Param("pageuuid")
	language := c.Param("language")
	log.Debugf("Get request <translate page> pageuuid: %s , target language: %s", uuid, language)

	// get notion page
	pageContent, err := transbot.NotionClient.FetchPage("d77601f7a3e649b7967f61a4462fad53")
	if err != nil {
		log.WithContext(WithGinContext(c)).Error("fetch page failed: ", err.Error())
		respondJSONError(c, http.StatusBadRequest, err)
		return
	}

	// convert string content to struct blocks
	page, err := notionopt.Content2NotionPage(pageContent)
	if err != nil {
		log.WithContext(WithGinContext(c)).Error("convert block error:: ", err.Error())
		respondJSONError(c, http.StatusBadRequest, err)
		return
	}

	// translate content
	for _, block := range page.PageContent.Results {
		toTrans, err := notionopt.GetBlockContent(block)
		if err != nil {
			log.WithContext(WithGinContext(c)).Error("get block content error: ", err.Error())
			respondJSONError(c, http.StatusBadRequest, err)
			return
		}
		if toTrans != "" {
			traned, err := transbot.Translate(toTrans, language)
			if err != nil {
				log.WithContext(WithGinContext(c)).Error("translate error: ", err.Error())
				respondJSONError(c, http.StatusBadRequest, err)
				return
			}

			err = notionopt.ReplaceBlockContent(block, traned)
			if err != nil {
				log.WithContext(WithGinContext(c)).Error("replace block content error: ", err.Error())
				respondJSONError(c, http.StatusBadRequest, err)
				return
			}
		}
	}

	// upload new page
	newPageuuid, err := transbot.NotionClient.UploadPage(page.PageInfo.ID, page)
	if err != nil {
		fmt.Println("UploadPage error: ", err)
		log.WithContext(WithGinContext(c)).Error("upload page error: ", err.Error())
		respondJSONError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": gin.H{
			"new_page": newPageuuid,
		},
	})
}
