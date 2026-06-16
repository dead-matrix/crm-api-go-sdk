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

// ProxyBindingsResult — сводка привязок прокси к аккаунтам пользователя
// (основной бот), GET /api/proxy/bindings. Аккаунт хранит прокси строкой;
// CRM матчит аккаунты с таблицей прокси по ip:port и считает агрегаты.
type ProxyBindingsResult struct {
	TotalAccounts          int64   `json:"total_accounts"`
	AccountsWithProxy      int64   `json:"accounts_with_proxy"`
	AccountsWithoutProxy   int64   `json:"accounts_without_proxy"`
	TotalProxies           int64   `json:"total_proxies"`
	ProxiesWithAccounts    int64   `json:"proxies_with_accounts"`
	ProxiesWithoutAccounts int64   `json:"proxies_without_accounts"`
	AvgAccountsPerProxy    float64 `json:"avg_accounts_per_proxy"`
}
