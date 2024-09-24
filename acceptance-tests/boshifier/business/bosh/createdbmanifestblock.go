package bosh

import (
	"boshifier/business/capi"
	"strings"
)

func CreateDBManifestBlock(sk capi.ServiceKey, dbEncryptionSecret string) DBBlock {

	dbb := DBBlock{
		Host:       sk.Credentials.Hostname,
		Name:       sk.Credentials.Name,
		User:       sk.Credentials.Username,
		Password:   sk.Credentials.Password,
		Port:       sk.Credentials.Port,
		Encryption: Encryption{Enabled: false, Passwords: []PasswordMetadata{}},
		CA: CA{
			Cert: strings.ReplaceAll(sk.Credentials.TLS.Cert.CA, "\n", "\\n"),
		},
	}

	if dbEncryptionSecret != "" {
		dbb.Encryption.Enabled = true
		dbb.Encryption.Passwords = []PasswordMetadata{
			{
				Password: Password{
					Secret: dbEncryptionSecret,
				},
				Label:   "first-encryption",
				Primary: true,
			},
		}
	}

	return dbb
}
