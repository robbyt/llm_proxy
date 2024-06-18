package schema

import (
	"fmt"
	"sync"

	"github.com/bojanz/currency"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"

	"github.com/proxati/llm_proxy/schema/providers/openai_com"
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

func (cc *API_Provider) addResponse(resp *ProxyResponse, chatCompResp *openai.ChatCompletionResponse) error {
	cc.rwMutex.Lock()
	defer cc.rwMutex.Unlock()
	cc.apiResponses = append(cc.apiResponses, resp)
	cc.apiResponseBodies = append(cc.apiResponseBodies, chatCompResp)

	// extract token quant, and calculate cost of transaction
	inputTokens := fmt.Sprint(chatCompResp.Usage.PromptTokens)
	outputTokens := fmt.Sprint(chatCompResp.Usage.CompletionTokens)

	inputCost, err := cc.costPerInputToken.Mul(inputTokens)
	if err != nil {
		return fmt.Errorf("failed to calculate input cost: %v", err)
	}

	outputCost, err := cc.costPerOutputToken.Mul(outputTokens)
	if err != nil {
		return fmt.Errorf("failed to calculate output cost: %v", err)
	}

	cc.totalCost, err = cc.totalCost.Add(inputCost)
	if err != nil {
		return fmt.Errorf("failed to add input cost to totalCost: %v", err)
	}

	cc.totalCost, err = cc.totalCost.Add(outputCost)
	if err != nil {
		return fmt.Errorf("failed to add output cost  to totalCost: %v", err)
	}

	return nil
}

type CostCounter struct {
	grandTotal  currency.Amount
	providers   map[string][]*API_Provider // key: provider URL, value: slide of models/products
	lookupCache map[string]*API_Provider   // key: provider URL + model, value: API_Provider
}

// NewCostCounter creates an object that _should_ be a singleton, in the Addon layer
func NewCostCounter() *CostCounter {
	cc := &CostCounter{
		providers:   make(map[string][]*API_Provider),
		lookupCache: make(map[string]*API_Provider),
	}

	// iterate over the pricing data and populate this struct, data loaded from json in the openai_com package
	for _, provider := range openai_com.API_Endpoint_Pricing {
		for _, product := range provider.Products {
			apiProvider, err := newAPI_Provider(provider.URL, product.Name, product.InputTokenCost, product.OutputTokenCost, "USD")
			if err != nil {
				panic(fmt.Sprintf("Error creating API_Provider: %v", err))
			}
			cc.providers[provider.URL] = append(cc.providers[provider.URL], apiProvider)
		}
	}
	return cc
}

func (cc *CostCounter) String() string {
	return cc.grandTotal.String()
}

func (cc *CostCounter) providerLookup(url, product string) *API_Provider {
	cacheKey := fmt.Sprintf("%s|%s", url, product)

	// Check if the result is already in the cache
	if provider, found := cc.lookupCache[cacheKey]; found {
		return provider // Return the cached result
	}

	providers := cc.providers[url]
	for _, provider := range providers {
		if provider.model == product {
			cc.lookupCache[cacheKey] = provider // Cache the result
			return provider
		}
	}
	return nil
}

func (cc *CostCounter) Add(req ProxyRequest, resp ProxyResponse) error {
	// load the request
	chatCompReq, err := openai_com.NewOpenAIChatCompletionRequest(&req.Body)
	if err != nil {
		return fmt.Errorf("failed to create OpenAI completion request: %v", err)
	}

	// find the provider, which is a combination of the URL and the requested model
	provider := cc.providerLookup(req.URL.String(), chatCompReq.Model)
	if provider == nil {
		return fmt.Errorf("provider not found for: %s|%s", req.URL.String(), chatCompReq.Model)
	}

	// store the request
	provider.addRequest(&req, chatCompReq)

	chatCompResp, err := openai_com.NewOpenAIChatCompletionResponse(&resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create OpenAI completion response: %v", err)
	}
	if chatCompResp == nil {
		return fmt.Errorf("chat completion response is nil")
	}

	// store the response (which adds currency totals to the provider)
	err = provider.addResponse(&resp, chatCompResp)
	if err != nil {
		return fmt.Errorf("failed to add response to provider: %v", err)
	}

	cc.grandTotal, err = cc.grandTotal.Add(provider.totalCost)
	if err != nil {
		return fmt.Errorf("failed to add provider total cost to grand total: %v", err)
	}
	log.Infof("Request for: %s using: %s cost: %s", req.URL.String(), chatCompReq.Model, provider.totalCost.String())

	return nil
}
