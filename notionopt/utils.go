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

	return merged, nil
}

// ConvertImageBlock convert image block to richtext link.
func ConvertImageBlock(blockDTO *notion.BlockDTO) notion.Block {
	if blockDTO.Image.Type == notion.FileTypeFile {
		url := blockDTO.Image.File.URL
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
	}
	return blockDTO
	// return notion.BlockDTO{
	// 	Type:      notion.BlockTypeParagraph,
	// 	Paragraph: blockDTO.Paragraph,
	// }
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
			return nil, fmt.Errorf("convert to notion.BlockDTO failed1")
		}
		if !IsSupported(&dto) {
			continue
		}
		if dto.Image != nil && dto.Image.Type == notion.FileTypeFile {
			block = ConvertImageBlock(&dto)
			continue
		}
		tmpBlocks = append(tmpBlocks, block)
	}
	page.PageContent.Results = tmpBlocks

	return &page, nil
}

// Get string block content
func GetBlockContent(block notion.Block) (string, error) {
	dto, ok := block.(notion.BlockDTO)
	if !ok {
		return "", ErrConvertDOTFailed
	}
	if !IsSupported(&dto) {
		return "", ErrBlockTypeUnsportected
	}
	richtext, err := GetRichtext(block)
	if err != nil {
		return "", nil
	}
	if richtext == nil {
		return "", nil
	}
	tmpString := GetFullRichtext(*richtext)
	return tmpString, nil
}

// Replace block content by a new string content.
// May cause loss of some formatting properties, such as color attributes.
func ReplaceBlockContent(block notion.Block, newContent string) error {
	dto, ok := block.(notion.BlockDTO)
	if !ok {
		return ErrConvertDOTFailed
	}
	if !IsSupported(&dto) {
		return ErrBlockTypeUnsportected
	}

	richtext, err := GetRichtext(block)
	if err != nil {
		return err
	}

	return ReplaceRichtext(richtext, newContent)
}

func ReplaceRichtext(richtext *[]notion.RichText, newContent string) error {
	if len(*richtext) == 0 {
		return ErrRichtextIsNull
	}
	*richtext = (*richtext)[0:1]
	(*richtext)[0].Text.Content = newContent
	return nil
}

func GetRichtext(block notion.Block) (*[]notion.RichText, error) {
	dto, ok := block.(notion.BlockDTO)
	if !ok {
		return nil, ErrConvertDOTFailed
	}
	if !IsSupported(&dto) {
		return nil, ErrBlockTypeUnsportected
	}

	switch dto.Type {
	case notion.BlockTypeParagraph:
		return &dto.Paragraph.RichText, nil
	case notion.BlockTypeHeading1:
		return &dto.Heading1.RichText, nil
	case notion.BlockTypeHeading2:
		return &dto.Heading2.RichText, nil
	case notion.BlockTypeHeading3:
		return &dto.Heading3.RichText, nil
	case notion.BlockTypeNumberedListItem:
		return &dto.NumberedListItem.RichText, nil
	case notion.BlockTypeBulletedListItem:
		return &dto.BulletedListItem.RichText, nil
	case notion.BlockTypeToDo:
		return &dto.ToDo.RichText, nil
	case notion.BlockTypeToggle:
		return &dto.Toggle.RichText, nil
	case notion.BlockTypeCallout:
		return &dto.Callout.RichText, nil
	case notion.BlockTypeQuote:
		return &dto.Quote.RichText, nil
	default:
		return nil, nil
	}
}

// GetFullRickText
// Merging all richtext into a single string.
// May cause loss of some formatting properties, such as color attributes.
func GetFullRichtext(richText []notion.RichText) string {
	fullContent := ""
	for _, rt := range richText {
		fullContent += rt.Text.Content
	}
	return fullContent
}

// Supported block types
func IsSupported(dto *notion.BlockDTO) bool {
	switch dto.Type {
	case notion.BlockTypeParagraph,
		notion.BlockTypeHeading1,
		notion.BlockTypeHeading2,
		notion.BlockTypeHeading3,
		notion.BlockTypeBulletedListItem,
		notion.BlockTypeNumberedListItem,
		notion.BlockTypeToDo,
		notion.BlockTypeToggle,
		notion.BlockTypeCallout,
		notion.BlockTypeVideo,
		notion.BlockTypeQuote,
		notion.BlockTypeImage:
		return true
	default:
		return false
	}
}
