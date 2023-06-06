package notionopt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cryptowizard0/go-notion"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/cryptowizard0/notion2arweave/utils"
	"github.com/go-resty/resty/v2"
)

// NotionOperator Implementation of INotionOperator
// Not considering concurrency
type NotionOperator struct {
	authToken    string
	httpClient   *resty.Client
	notionClient *notion.Client
	s3Client     *s3.Client
	s3Endpoint   string
	s3Key        string
	s3Secret     string
	s3Bucket     string
}

// CreateNotionOperator
func CreateNotionOperator(auth string) *NotionOperator {
	// http client
	client := resty.New()
	client.SetHeader("Accept", "application/json").
		SetHeader("Notion-Version", viper.GetString("notion.version")).
		SetAuthToken(auth).
		SetBaseURL(viper.GetString("notion.base_url"))

	// s3 client
	key := viper.GetString("4everland.key")
	secret := viper.GetString("4everland.secret")
	endpoint := viper.GetString("4everland.endpoint")
	bucket := viper.GetString("4everland.bucket_name")
	token := ""
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(key, secret, token)),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: endpoint,
			}, nil
		})),
	)
	if err != nil {
		log.Fatalf("unable to load s3 config, %v", err)
	}
	s3client := s3.NewFromConfig(cfg)

	return &NotionOperator{
		authToken:    auth,
		httpClient:   client,
		notionClient: notion.NewClient(auth),
		s3Client:     s3client,
		s3Key:        key,
		s3Secret:     secret,
		s3Endpoint:   endpoint,
		s3Bucket:     bucket,
	}
}

// Fetch page from notion
// @Pararm uuid, page uuid
// @Return txId, txId return by Arweave
func (n *NotionOperator) FetchPage(uuid string) (content string, err error) {
	log.WithField("uuid", uuid).Info("notion operator: fetch page")

	// 1. get page info
	strPageInfo, err := n.fetchPageInfo(uuid)
	if err != nil {
		log.Error("fetch page info error:", err.Error())
		return "", err
	}

	// 2. get child blocks
	strPageContent, err := n.fetchPageContent(uuid, "")
	if err != nil {
		log.Error("fetch content error:", err.Error())
		return "", err
	}

	// 3. make full content
	content = fmt.Sprintf(`{"page_info":%s,"page_content":%s}`, strPageInfo, strPageContent)
	log.WithField("uuid", uuid).Info("notion operator: fetch page done")
	return
}

// UploadPage uploading page content to notion
// @Pararm parentId, parent page uuid, where new page to be loaded
// @Pararm page, page content
// @Return uuid, uuid of new page
func (n *NotionOperator) UploadPage(parentId string, page *NotionPage) (uuid string, err error) {
	log.WithField("parent", parentId).Info("notion operator: upload page")

	var title []notion.RichText
	pageProp, ok := page.PageInfo.Properties.(notion.PageProperties)
	if ok {
		title = pageProp.Title.Title
	} else {
		// try to covert to (notion.DatabasePageProperty)
		dbProp, ok := page.PageInfo.Properties.(notion.DatabasePageProperties)
		if !ok {
			return "", fmt.Errorf("convert page preperites error")
		}
		title = dbProp["Name"].Title
	}
	newPageParams := notion.CreatePageParams{
		ParentType: notion.ParentTypePage,
		ParentID:   parentId,
		Title:      title,
		Children:   nil,
		Icon:       page.PageInfo.Icon,
		Cover:      page.PageInfo.Cover,
	}

	newPage, err := n.notionClient.CreatePage(context.Background(), newPageParams)
	if err != nil {
		return "", err
	}

	// upload content blocks, Max 100 per request
	blocks := page.PageContent.Results
	for i := 0; (i * 100) < len(blocks); i++ {
		starindex := i * 100
		endindex := i*100 + 100
		if endindex > len(blocks) {
			endindex = len(blocks)
		}
		_, err = n.notionClient.AppendBlockChildren(context.Background(), newPage.ID, blocks[starindex:endindex])
		if err != nil {
			return "", err
		}
	}

	return newPage.ID, nil
}

// Content2NotionPage converting string content to NotionPage struct
func (n *NotionOperator) Content2NotionPage(srcContent string) (*NotionPage, error) {
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
		//if dto.Image != nil && dto.Image.Type == notion.FileTypeFile {
		if dto.Type == notion.BlockTypeImage {
			block = n.ConvertImageBlock(&dto)
		}
		tmpBlocks = append(tmpBlocks, block)
	}
	page.PageContent.Results = tmpBlocks

	return &page, nil
}

func (n *NotionOperator) CreateNewPage(parentId string, page *NotionPage) (uuid string, err error) {
	log.WithField("parent", parentId).Info("notion operator: create page")

	var title []notion.RichText
	pageProp, ok := page.PageInfo.Properties.(notion.PageProperties)
	if ok {
		title = pageProp.Title.Title
	} else {
		// try to covert to (notion.DatabasePageProperty)
		dbProp, ok := page.PageInfo.Properties.(notion.DatabasePageProperties)
		if !ok {
			return "", fmt.Errorf("convert page preperites error")
		}
		title = dbProp["Name"].Title
	}
	newPageParams := notion.CreatePageParams{
		ParentType: notion.ParentTypePage,
		ParentID:   parentId,
		Title:      title,
		Children:   nil,
		Icon:       page.PageInfo.Icon,
		Cover:      page.PageInfo.Cover,
	}

	newPage, err := n.notionClient.CreatePage(context.Background(), newPageParams)
	if err != nil {
		return "", err
	}
	return newPage.ID, nil
}

func (n *NotionOperator) AppendBlockChildren(parentId string, block notion.Block) error {
	blocks := []notion.Block{block}
	_, err := n.notionClient.AppendBlockChildren(context.Background(), parentId, blocks)
	return err
}

// ========================================================================
// fetchPageInfo
func (n *NotionOperator) fetchPageInfo(uuid string) (content string, err error) {
	url := fmt.Sprintf("/v1/pages/%s", uuid)
	resp, err := n.httpClient.R().Get(url)
	if err != nil {
		log.Error("get request error: ", err.Error())
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		utils.LogResp_Error(resp)
		return "", fmt.Errorf(resp.String())
	}

	return string(resp.Body()), nil
}

// fetchPageInfoByNotionSdk
func (n *NotionOperator) fetchPageInfoByNotionSdk(uuid string) (content string, err error) {
	page, err := n.notionClient.FindPageByID(context.Background(), uuid)
	if err != nil {
		log.Error("get page error: ", err.Error())
		return "", err
	}

	jsonPage, err := json.Marshal(page)
	if err != nil {
		return "", err
	}

	return string(jsonPage), nil
}

// fetchPageContent
func (n *NotionOperator) fetchPageContent(uuid, startCursor string) (content string, err error) {
	var url string
	if startCursor == "" {
		url = fmt.Sprintf("/v1/blocks/%s/children", uuid)
	} else {
		url = fmt.Sprintf("/v1/blocks/%s/children?start_cursor=%s", uuid, startCursor)
	}

	resp, err := n.httpClient.R().Get(url)
	if err != nil {
		log.Error("get request error:", err.Error())
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		utils.LogResp_Error(resp)
		return "", fmt.Errorf(resp.String())
	}

	type hasMore struct {
		HasMore    bool    `json:"has_more"`
		NextCursor *string `json:"next_cursor"`
	}
	var morePage hasMore
	err = json.Unmarshal(resp.Body(), &morePage)
	if err != nil {
		return "", err
	}
	fullContent := string(resp.Body())

	if morePage.HasMore {
		log.WithField("id", morePage.NextCursor).Info("more page")
		moreContent, err := n.fetchPageContent(uuid, *morePage.NextCursor)
		if err != nil {
			return "", err
		}
		fullContent, err = MergeChildBlocks(string(resp.Body()), moreContent)
		if err != nil {
			return "", err
		}
	}

	return fullContent, nil
}

// ConvertImageBlock
// 1. Down image
// 2. Upload image content to 4everland buckets
// 3. Replace url with the new image path
func (n *NotionOperator) ConvertImageBlock(blockDTO *notion.BlockDTO) notion.Block {
	if blockDTO.Image.Type == notion.FileTypeFile {
		objectKey := blockDTO.ID() + ".jpg"
		newurl, err := n.uploadImageTo4Everland(blockDTO.Image.File.URL, objectKey)
		if err != nil {
			log.Error("upload to 4everland error: ", err) // Just log the error
		} else {
			blockDTO.Image.File = nil
			blockDTO.Image.Type = notion.FileTypeExternal
			blockDTO.Image.External = &notion.FileExternal{
				URL: newurl,
			}
		}
	}
	return *blockDTO
}

func (n *NotionOperator) uploadImageTo4Everland(imgPath, objectKey string) (url string, err error) {
	log.Info("Download image: ", imgPath)
	c := resty.New()
	resp, err := c.R().Get(imgPath)
	if err != nil {
		log.Error("get image error: ", err)
		return "", err
	}
	// upload image to object store
	input := &s3.PutObjectInput{
		Bucket: aws.String(n.s3Bucket),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(resp.Body()),
	}
	_, err = n.s3Client.PutObject(context.TODO(), input)
	if err != nil {
		log.Error("upload image error: ", err)
		return "", err
	}

	url = fmt.Sprintf("https://%s.4everland.store/%s", n.s3Bucket, objectKey)
	return
}
