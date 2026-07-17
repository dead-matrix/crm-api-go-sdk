package crmapi

type ProductEntry struct {
	Title      string `json:"title"`
	PriceMinor int64  `json:"price_minor"`
	PriceUSD   *int64 `json:"price_usd,omitempty"`
	// Токен-пак (access.tokens > 0): при выставлении платежа CRM требует
	// для таких товаров функцию AI-леджера (ai_function).
	IsTokenPack bool `json:"is_token_pack,omitempty"`
}

type CategoryBucket struct {
	Title    string                  `json:"title"`
	Products map[string]ProductEntry `json:"products"`
}
