package notionopt

import (
	"encoding/json"
	"fmt"

	"github.com/cryptowizard0/go-notion"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// MergeChildBlocks
// merge the content into a single JSON string
// content format is notion.BlockChildrenResponse
func MergeChildBlocks(content1, content2 string) (merged string, err error) {
	value1 := gjson.Get(content1, "results")
	value2 := gjson.Get(content2, "results")

	array1 := value1.Array()
	array2 := value2.Array()
	array1 = append(array1, array2...)
	strArr := "["
	for i, block := range array1 {
		strArr += block.String()
		if i < len(array1)-1 {
			strArr += ","
		}
	}
	strArr += "]"

	merged = fmt.Sprintf("{\"object\":\"%s\",\"results\":%s,\"type\":\"%s\",\"block\":\"%s\"}",
		gjson.Get(content1, "object").String(),
		strArr,
		gjson.Get(content1, "type").String(),
		gjson.Get(content1, "obblockject").String())

	log.Debug("merged: ", merged)

	return merged, nil
}

// ConvertImageBlock convert image block to richtext link.
func ConvertImageBlock(blockDTO *notion.BlockDTO) notion.Block {
	imageBlock := blockDTO.Image

	var url string
	if imageBlock.Type == notion.FileTypeExternal {
		url = imageBlock.External.URL
	} else {
		url = imageBlock.File.URL
	}
	textContent := fmt.Sprintf("Image: %s", url)
	richTextBlock := notion.ParagraphBlock{
		RichText: []notion.RichText{
			{
				Type: notion.RichTextTypeText,
				Text: &notion.Text{
					Content: textContent,
					Link: &notion.Link{
						URL: url,
					},
				},
				PlainText: textContent,
				HRef:      &url,
			},
		},
	}

	blockDTO.Type = notion.BlockTypeParagraph
	blockDTO.Paragraph = &richTextBlock
	blockDTO.Image = nil

	return blockDTO
}

// content2NotionPage converting string content to NotionPage struct
func Content2NotionPage(srcContent string) (*NotionPage, error) {
	var page NotionPage
	err := json.Unmarshal([]byte(srcContent), &page)
	if err != nil {
		return nil, err
	}
	log.Info("Blocks count: ", len(page.PageContent.Results))
	var tmpBlocks []notion.Block

	for _, block := range page.PageContent.Results {
		dto, ok := block.(notion.BlockDTO)
		if !ok {
			return nil, fmt.Errorf("convert to notion.BlockDTO failed")
		}
		if !IsSupported(&dto) {
			continue
		}
		if dto.Image != nil {
			block = ConvertImageBlock(&dto)
		}
		tmpBlocks = append(tmpBlocks, block)
	}
	page.PageContent.Results = tmpBlocks

	return &page, nil
}

//
func IsSupported(dto *notion.BlockDTO) bool {
	if dto.Paragraph != nil ||
		dto.Heading1 != nil ||
		dto.Heading2 != nil ||
		dto.Heading3 != nil ||
		dto.BulletedListItem != nil ||
		dto.NumberedListItem != nil ||
		dto.ToDo != nil ||
		dto.Toggle != nil ||
		dto.Callout != nil ||
		dto.Divider != nil ||
		dto.Video != nil ||
		dto.Quote != nil ||
		dto.Image != nil {
		return true
	}

	return false
}

func GetBlockContent(block *notion.Block) (string, error) {
	return "", nil
}

func ReplaceBlockContent(block *notion.Block, newContent string) error {
	return nil
}
