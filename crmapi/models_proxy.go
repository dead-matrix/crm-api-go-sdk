package crmapi

type ProxyCheckItem struct {
	Proxy    string  `json:"proxy"`
	Valid    bool    `json:"valid"`
	RUError  *string `json:"ru_error,omitempty"`
	Location *string `json:"location,omitempty"`
}

type ProxyCheckResult struct {
	Checked int64            `json:"checked"`
	Valid   int64            `json:"valid"`
	Invalid int64            `json:"invalid"`
	Results []ProxyCheckItem `json:"results"`
}

type ProxyItem struct {
	Type     *string `json:"type,omitempty"`
	IP       *string `json:"ip,omitempty"`
	Port     *int64  `json:"port,omitempty"`
	Login    *string `json:"login,omitempty"`
	Password *string `json:"password,omitempty"`
	Valid    bool    `json:"valid"`
	Location *string `json:"location,omitempty"`
}
