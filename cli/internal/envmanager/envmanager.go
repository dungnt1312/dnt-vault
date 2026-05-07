package envmanager

import (
	"github.com/dnt/vault-cli/internal/crypto"
)

func EncryptVariables(vars map[string]string, password string) (map[string]string, error) {
	encrypted := make(map[string]string, len(vars))
	for k, v := range vars {
		enc, err := crypto.Encrypt(v, password)
		if err != nil {
			return nil, err
		}
		encrypted[k] = enc
	}
	return encrypted, nil
}

func DecryptVariables(vars map[string]string, password string) (map[string]string, error) {
	decrypted := make(map[string]string, len(vars))
	for k, v := range vars {
		dec, err := crypto.Decrypt(v, password)
		if err != nil {
			return nil, err
		}
		decrypted[k] = dec
	}
	return decrypted, nil
}

func MergeVariables(existing, incoming map[string]string) map[string]string {
	out := make(map[string]string, len(existing)+len(incoming))
	for k, v := range existing {
		out[k] = v
	}
	for k, v := range incoming {
		out[k] = v
	}
	return out
}
