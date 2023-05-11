# transbot
Automatic translation and typesetting of noiton articles using OpenAI.

## Config
- rename 'config_temp.toml' to 'config.toml' and field your api_key
- <openai.api_key> must be your OpenAI api key
- <notion.api_auth> must be your notion secret key
## Building and run 
### Using go cmd
- go mod tidy
- go run main.go

### Using makefile
- go mod tidy
- make build
- ./transbot

### Using docker
- make docker
- make run

## Frontend
Frontend is a sample frontend for transbot.

## Port
- Default transbot port is 8080. You can modify in 'config.toml' and 'Makefile': <TRANSBOT_PORT>
- Drfault frontend port is 8081. You can modify in 'Makefile': <FRONTEND_PORT>

## Using
### Restful api
```
GET: /v1/translate/:notion_page_url/:target_language
```
- **notion_page_url** is the notion page to be translated.
- **target_language** is the language to be translated. (english or chenese)

``` shell
# Example
curl --location 'http://127.0.0.1:8080/v1/translate/d77601f7a3e649b7967f61a4462fad53/english'
```

## Supported notion block types
- Paragraph
- Heading1
- Heading2
- Heading3
- NumberedListItem
- BulletedListItem
- ToDo
- Toggle
- Callout
- Quote
- Video
- Image

### 👉 Key references
- go-notion fork: https://github.com/cryptowizard0/go-notion 
- go-openai: https://github.com/sashabaranov/go-openai