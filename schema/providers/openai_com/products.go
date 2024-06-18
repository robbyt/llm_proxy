package openai_com

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
)

//go:embed data.json
var pricingDataJSON embed.FS

var API_Endpoint_Pricing []APIEndpoint

type Product struct {
	Name            string `json:"name"`
	InputTokenCost  string `json:"inputTokenCost"`
	OutputTokenCost string `json:"outputTokenCost"`
	Currency        string `json:"currency"`
}

type APIEndpoint struct {
	URL      string    `json:"url"`
	Products []Product `json:"products"`
}

func loadEmbeddedDataJSON() error {
	data, err := fs.ReadFile(pricingDataJSON, "data.json")
	if err != nil {
		return fmt.Errorf("failed to read embedded data.json: %w", err)
	}
	return json.Unmarshal(data, &API_Endpoint_Pricing) // Fixed: Pass a pointer to API_Endpoint_Pricing
}

func init() {
	err := loadEmbeddedDataJSON()
	if err != nil {
		panic(fmt.Sprintf("Error loading openai pricing data: %v\n", err))
	}
}
