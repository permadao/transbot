package notionopt

import "github.com/cryptowizard0/go-notion"

// notion page
type NotionPage struct {
	PageInfo    notion.Page                  `json:"page_info"`
	PageContent notion.BlockChildrenResponse `json:"page_content"`
}
