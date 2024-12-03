package solanchor

type Account struct {
	Name     string `json:"name"`
	Writable bool   `json:"writable"`
	Signer   bool   `json:"signer"`
	Address  string `json:"address"`
}
