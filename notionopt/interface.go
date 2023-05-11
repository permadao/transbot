package notionopt

// INotionOperator.
// Get Notion page content, support for recursion. Object serialization.
type INotionOperator interface {
	// Fetch notion content for a given page.
	// !MUST support recursive
	// @Pararm uuid,
	// @Return string content
	FetchPage(uuid string) (content string, err error)

	// Upload content to notion
	// @Pararm parentId, uuid of parent page
	// @Pararm content, json format page content
	// @Return uuid, new page's uuid
	UploadPage(parentId string, page *NotionPage) (uuid string, err error)

	// Content2NotionPage converting string content to NotionPage struct
	// @Pararm srcContent, JSON string page content
	// @Return *NotionPage, converted page
	Content2NotionPage(srcContent string) (*NotionPage, error)

	// GetBlockContent(block *notion.Block) (string, error)
}
