package crmapi

type ProductEntry struct {
	Title      string `json:"title"`
	PriceMinor int64  `json:"price_minor"`
	PriceUSD   *int64 `json:"price_usd,omitempty"`
}

type CategoryBucket struct {
	Title    string                  `json:"title"`
	Products map[string]ProductEntry `json:"products"`
}
