package contract

type Config struct {
	Network         string `json:"network"`
	Ingress         string `json:"ingress"`
	Infura          string `json:"infura"`
	Etherscan       string `json:"etherscan"`
	EthNetwork      string `json:"ethNetwork"`
	EthNetworkLabel string `json:"ethNetworkLabel"`
	LedgerNetworkID string `json:"ledgerNetworkId"`
	Contracts       []ConfigContracts
	Tokens          ConfigTokens
}

type ConfigContracts struct {
	DarknodeRegistry string `json:"darknodeRegistry"`
	Orderbook        string `json:"orderbook"`
	RenExTokens      string `json:"renExTokens"`
	RenExBalances    string `json:"renExBalances"`
	RenExSettlement  string `json:"renExSettlement"`
	Wyre             string `json:"wyre"`
}

type ConfigTokens struct {
	TUSD string `json:"TUSD"`
	DGX  string `json:"DGX"`
	REN  string `json:"REN"`
	OMG  string `json:"OMG"`
	ZRX  string `json:"ZRX"`
}
