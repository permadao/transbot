package translator

import (
	"context"
	"fmt"

	"github.com/permadao/transbot/notionopt"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Translator struct {
	APIKey       string
	AiClient     *openai.Client
	NotionClient *notionopt.NotionOperator
	Temperature  float64
	Model        string
}

func CreateTranslator(apiKey, notionAuth string) *Translator {
	return &Translator{
		APIKey:       apiKey,
		AiClient:     openai.NewClient(apiKey),
		NotionClient: notionopt.CreateNotionOperator(notionAuth),
		Temperature:  viper.GetFloat64("openai.temperature"),
		Model:        viper.GetString("openai.model"),
	}
}

func (a *Translator) Translate(content, targetLanguage string) (string, error) {
	c := fmt.Sprintf("Translate to %s: %s", targetLanguage, content)
	return a.OpenAIRequest(c)
}

func (a *Translator) OpenAIRequest(content string) (string, error) {
	log.Info("chat completion: ", content)
	resp, err := a.AiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: a.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
		},
	)
	if err != nil {
		log.Error("chat completion content error:", err.Error())
		return "", err
	}

	log.Info("openai: ", resp.Choices[0].Message.Content)
	return resp.Choices[0].Message.Content, nil
}
