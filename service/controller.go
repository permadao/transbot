package service

import (
	"fmt"
	"net/http"

	"github.com/cryptowizard0/go-notion"
	"github.com/gin-gonic/gin"
	"github.com/permadao/transbot/notionopt"
	log "github.com/sirupsen/logrus"
)

func TranslatePage(c *gin.Context) {
	uuid := c.Param("pageuuid")
	language := c.Param("language")
	log.Debugf("Get request <translate page> pageuuid: %s , target language: %s", uuid, language)

	go translate_segmentation(uuid, language)

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "OK",
	})
}

// Divide the text into segments and write each segment
// into Notion after completing its translation,
// and write the translated segments separately.
func translate_segmentation(uuid, language string) {
	// get notion page
	pageContent, err := transbot.NotionClient.FetchPage(uuid)
	if err != nil {
		log.WithField("page_id", uuid).Error("fetch page failed: ", err.Error())
		return
	}

	// convert string content to struct blocks
	page, err := transbot.NotionClient.Content2NotionPage(pageContent)
	if err != nil {
		log.Error("convert block error: ", err.Error())
		return
	}
	var richTitle *[]notion.RichText
	// traslate title
	pageProp, ok := page.PageInfo.Properties.(notion.PageProperties)
	if !ok {
		// try to covert to (notion.DatabasePageProperty)
		dbProp, ok := page.PageInfo.Properties.(notion.DatabasePageProperties)
		if !ok {
			log.Error("get page properties failed!")
			return
		}
		tmpRichTitle := dbProp["Name"].Title
		richTitle = &tmpRichTitle
	} else {
		richTitle = &pageProp.Title.Title
	}
	title := notionopt.GetFullRichtext(*richTitle)
	tranedTitle, err := transbot.Translate(title, language)
	if err != nil {
		log.Error("translate title error: ", err.Error())
		return
	}
	notionopt.ReplaceRichtext(richTitle, tranedTitle)

	// create new page
	newPageuuid, err := transbot.NotionClient.CreateNewPage(page.PageInfo.ID, page)
	if err != nil {
		log.Error("create new page error: ", err.Error())
		return
	}

	// translate content
	for _, block := range page.PageContent.Results {
		toTrans, err := notionopt.GetBlockContent(block)
		if err != nil {
			log.Error("get block content error: ", err.Error())
			return
		}
		if toTrans != "" {
			traned, err := transbot.Translate(toTrans, language)
			if err != nil {
				log.Error("translate block content error: ", err.Error())
				return
			}

			err = notionopt.ReplaceBlockContent(block, traned)
			if err != nil {
				log.Error("replace block content error: ", err.Error())
				return
			}

			err = transbot.NotionClient.AppendBlockChildren(newPageuuid, block)
			if err != nil {
				log.Error("append child block error: ", err.Error())
				return
			}
		} else { // contentless block, like image
			err = transbot.NotionClient.AppendBlockChildren(newPageuuid, block)
			if err != nil {
				log.Error("append child block error: ", err.Error())
				return
			}
		}
	}
}

// Concurrent translation, aggregate the results after translation,
// and generate the complete page in one go.
func translate_concurrent(c *gin.Context) {
	uuid := c.Param("pageuuid")
	language := c.Param("language")
	log.Debugf("Get request <translate page> pageuuid: %s , target language: %s", uuid, language)

	// get notion page
	pageContent, err := transbot.NotionClient.FetchPage(uuid)
	if err != nil {
		log.WithContext(WithGinContext(c)).Error("fetch page failed: ", err.Error())
		respondJSONError(c, http.StatusBadRequest, err)
		return
	}

	// convert string content to struct blocks
	page, err := transbot.NotionClient.Content2NotionPage(pageContent)
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
