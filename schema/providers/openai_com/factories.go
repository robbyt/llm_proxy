package openai_com

import (
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func NewOpenAIChatCompletionRequest(body *string) (completion *openai.ChatCompletionRequest, err error) {
	bodyBytes := []byte(*body)

	err = json.Unmarshal(bodyBytes, &completion)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal OpenAI completion response body: %v", err)
	}

	return completion, nil
}

func NewOpenAIChatCompletionResponse(body *string) (completion *openai.ChatCompletionResponse, err error) {
	bodyBytes := []byte(*body)

	err = json.Unmarshal(bodyBytes, &completion)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal OpenAI completion response body: %v", err)
	}

	return completion, nil
}
