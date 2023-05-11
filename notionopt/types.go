package notionopt

import (
	"errors"

	"github.com/cryptowizard0/go-notion"
)

var (
	ErrBlockTypeUnsportected = errors.New("unsportected block type")
	ErrConvertDOTFailed      = errors.New("convert to notion.BlockDTO failed")
	ErrRichtextIsNull        = errors.New("the text is null")
)

// notion page
type NotionPage struct {
	PageInfo    notion.Page                  `json:"page_info"`
	PageContent notion.BlockChildrenResponse `json:"page_content"`
}
