package bosh

type Password struct {
	Secret string `json:"secret"`
}

type PasswordMetadata struct {
	Password Password `json:"password"`
	Label    string   `json:"label"`
	Primary  bool     `json:"primary"`
}

type Encryption struct {
	Enabled   bool               `json:"enabled"`
	Passwords []PasswordMetadata `json:"passwords"`
}

type CA struct {
	Cert string `json:"cert"`
}

type DBBlock struct {
	Host       string     `json:"host"`
	Name       string     `json:"name"`
	User       string     `json:"user"`
	Password   string     `json:"password"`
	Port       int        `json:"port"`
	Encryption Encryption `json:"encryption"`
	CA         CA         `json:"ca"`
}
