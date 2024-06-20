package schema

import (
	"fmt"
	"sync"

	"github.com/bojanz/currency"
	log "github.com/sirupsen/logrus"

	"github.com/proxati/llm_proxy/schema/providers/openai_com"
)

type CostCounter struct {
	grandTotal  currency.Amount
	providers   map[string][]*API_Provider // key: provider URL, value: slice of models/products
	lookupCache map[string]*API_Provider   // key: provider URL + model, value: API_Provider
	formatter   *currency.Formatter
	rwMutex     sync.RWMutex
}

// NewCostCounter creates an object that _should_ be a singleton, in the Addon layer
func NewCostCounter() *CostCounter {
	loc := currency.NewLocale("en-US")

	cc := &CostCounter{
		providers:   make(map[string][]*API_Provider),
		lookupCache: make(map[string]*API_Provider),
		formatter:   currency.NewFormatter(loc),
		rwMutex:     sync.RWMutex{},
	}

	// iterate over the pricing data and populate this struct, data loaded from json in the openai_com package
	for _, provider := range openai_com.API_Endpoint_Data {
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
	cc.rwMutex.RLock()
	defer cc.rwMutex.RUnlock()
	return cc.grandTotal.String()
}

func (cc *CostCounter) providerCacheLookup(cacheKey string) *API_Provider {
	cc.rwMutex.RLock()
	defer cc.rwMutex.RUnlock()

	if provider, found := cc.lookupCache[cacheKey]; found {
		return provider // Return the cached result
	}
	return nil
}

func (cc *CostCounter) providerLookup(url, product string) *API_Provider {
	cacheKey := fmt.Sprintf("%s|%s", url, product)

	// Check if the result is already in the cache
	if provider := cc.providerCacheLookup(cacheKey); provider != nil {
		// found it in the cache, return it
		return provider
	}

	providers := cc.providers[url]
	for _, provider := range providers {
		if provider.model == product {
			cc.rwMutex.Lock()
			defer cc.rwMutex.Unlock()

			cc.lookupCache[cacheKey] = provider // Cache the result
			return provider
		}
	}
	return nil
}

func (cc *CostCounter) Add(req ProxyRequest, resp ProxyResponse) error {
	// parse the request
	chatCompReq, err := openai_com.NewOpenAIChatCompletionRequest(&req.Body)
	if err != nil || chatCompReq == nil {
		return fmt.Errorf("failed to create OpenAI completion request: %v", err)
	}

	// find the provider, which is a combination of the URL and the requested model
	provider := cc.providerLookup(req.URL.String(), chatCompReq.Model)
	if provider == nil {
		return fmt.Errorf("provider not found for: %s|%s", req.URL.String(), chatCompReq.Model)
	}

	// store the request objects
	provider.addRequest(&req, chatCompReq)

	// parse the response object
	chatCompResp, err := openai_com.NewOpenAIChatCompletionResponse(&resp.Body)
	if err != nil || chatCompResp == nil {
		return fmt.Errorf("failed to create OpenAI completion response: %v", err)
	}

	// store the response objects
	provider.addResponse(&resp, chatCompResp)

	// calculate the cost
	inputCost, outputCost, err := provider.calculateCost(chatCompResp)
	if err != nil {
		return fmt.Errorf("failed to calculate cost: %v", err)
	}

	cc.rwMutex.Lock()
	defer cc.rwMutex.Unlock()

	cc.grandTotal, err = cc.grandTotal.Add(inputCost)
	if err != nil {
		return fmt.Errorf("failed to add provider total cost to grand total: %v", err)
	}

	cc.grandTotal, err = cc.grandTotal.Add(outputCost)
	if err != nil {
		return fmt.Errorf("failed to add provider total cost to grand total: %v", err)
	}
	totalReqCost, _ := inputCost.Add(outputCost)

	log.Infof("Request for: %s using: %s inputCost: %s outputCost %s = Request Cost: %s",
		req.URL.String(), chatCompReq.Model, cc.formatter.Format(inputCost), cc.formatter.Format(outputCost), cc.formatter.Format(totalReqCost))
	log.Infof("Total cost for this session: %s", cc.formatter.Format(cc.grandTotal))

	return nil
}
