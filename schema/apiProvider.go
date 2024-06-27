package schema

import (
	"fmt"
	"sync"

	"github.com/bojanz/currency"
	openai "github.com/sashabaranov/go-openai"
)

type API_Provider struct {
	name               string
	model              string
	currencyUnit       string
	costPerInputToken  currency.Amount
	costPerOutputToken currency.Amount
	totalCost          currency.Amount
	apiRequests        []*ProxyRequest
	apiRequestBodies   []*openai.ChatCompletionRequest
	apiResponses       []*ProxyResponse
	apiResponseBodies  []*openai.ChatCompletionResponse
	rwMutex            sync.RWMutex
}

// newAPI_Provider creates a single object for a URL/Model combination
func newAPI_Provider(name, model, inputCost, outputCost, currencyUnit string) (*API_Provider, error) {
	iCost, err := currency.NewAmount(inputCost, currencyUnit)
	if err != nil {
		return nil, fmt.Errorf("failed to create input currency amount: %v", err)
	}

	oCost, err := currency.NewAmount(outputCost, currencyUnit)
	if err != nil {
		return nil, fmt.Errorf("failed to create output currency amount: %v", err)
	}

	total, _ := currency.NewAmount("0", currencyUnit)

	return &API_Provider{
		name:               name,
		model:              model,
		currencyUnit:       currencyUnit,
		costPerInputToken:  iCost,
		costPerOutputToken: oCost,
		totalCost:          total,
		apiRequests:        make([]*ProxyRequest, 0),
		apiRequestBodies:   make([]*openai.ChatCompletionRequest, 0),
		apiResponses:       make([]*ProxyResponse, 0),
		apiResponseBodies:  make([]*openai.ChatCompletionResponse, 0),
		rwMutex:            sync.RWMutex{},
	}, nil
}
func (cc *API_Provider) String() string {
	cc.rwMutex.RLock()
	defer cc.rwMutex.RUnlock()
	return cc.totalCost.Round().String()
}

func (cc *API_Provider) addRequest(req *ProxyRequest, chatCompReq *openai.ChatCompletionRequest) {
	cc.rwMutex.Lock()
	defer cc.rwMutex.Unlock()
	cc.apiRequests = append(cc.apiRequests, req)
	cc.apiRequestBodies = append(cc.apiRequestBodies, chatCompReq)
}

func (cc *API_Provider) addResponse(resp *ProxyResponse, chatCompResp *openai.ChatCompletionResponse) {
	cc.rwMutex.Lock()
	defer cc.rwMutex.Unlock()
	cc.apiResponses = append(cc.apiResponses, resp)
	cc.apiResponseBodies = append(cc.apiResponseBodies, chatCompResp)
}

func (cc *API_Provider) calculateCost(chatCompResp *openai.ChatCompletionResponse) (inputCost, outputCost currency.Amount, err error) {
	// extract token quant, and calculate cost of transaction
	inputTokens := fmt.Sprint(chatCompResp.Usage.PromptTokens)
	outputTokens := fmt.Sprint(chatCompResp.Usage.CompletionTokens)

	inputCost, err = cc.costPerInputToken.Mul(inputTokens)
	if err != nil {
		return currency.Amount{}, currency.Amount{}, fmt.Errorf("failed to calculate input cost: %v", err)
	}

	outputCost, err = cc.costPerOutputToken.Mul(outputTokens)
	if err != nil {
		return currency.Amount{}, currency.Amount{}, fmt.Errorf("failed to calculate output cost: %v", err)
	}

	cc.totalCost, err = cc.totalCost.Add(inputCost)
	if err != nil {
		return currency.Amount{}, currency.Amount{}, fmt.Errorf("failed to add input cost to totalCost: %v", err)
	}

	cc.totalCost, err = cc.totalCost.Add(outputCost)
	if err != nil {
		return currency.Amount{}, currency.Amount{}, fmt.Errorf("failed to add output cost  to totalCost: %v", err)
	}

	return inputCost, outputCost, nil
}
