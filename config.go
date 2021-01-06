package bitxid

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const configName = "bitxid.toml"

// Config .
type Config struct {
	Title        string `json:"title"`
	BitXIDConfig `json:"bitxid"`
}

// BitXIDConfig .
type BitXIDConfig struct {
	DIDConfig    `json:"did"`
	MethodConfig `json:"method"`
}

// DIDConfig .
type DIDConfig struct {
	Admin    DID     `toml:"admin" json:"admin"`
	AdminDoc *DIDDoc `toml:"admin_doc" json:"admin_doc"`
	Addr     string  `toml:"addr" json:"addr"`
}

// MethodConfig .
type MethodConfig struct {
	Mode          RegistryMode
	Admin         DID        `toml:"admin" json:"admin"`
	IsRoot        bool       `toml:"is_root" json:"is_root"`
	GenesisMetohd DID        `toml:"genesis_metohd" json:"genesis_metohd"`
	GenesisDoc    *MethodDoc `toml:"genesis_doc" json:"genesis_doc"`
}

// DefaultConfig .
func DefaultConfig() (*Config, error) {
	BConfig, _ := DefaultBitXIDConfig()
	return &Config{
		Title:        "",
		BitXIDConfig: *BConfig,
	}, nil
}

// DefaultBitXIDConfig .
func DefaultBitXIDConfig() (*BitXIDConfig, error) {
	return &BitXIDConfig{
		DIDConfig: DIDConfig{
			Addr:     ".",
			Admin:    "did:bitxhub:appchain001:0x00000001",
			AdminDoc: getAdminDoc(),
		},
		MethodConfig: MethodConfig{
			Admin: "did:bitxhub:relayroot:0x00000001",
			// AdminDoc:      getAdminDoc(),
			// Addr:          ".",
			IsRoot:        true,
			GenesisMetohd: "did:bitxhub:relayroot:.",
			GenesisDoc:    genesisMetohdDoc(),
		},
	}, nil
}

func getAdminDoc() *DIDDoc {
	doc := &DIDDoc{}
	doc.ID = "did:bitxhub:appchain001:0x00000001"
	doc.Type = "user"
	pk := PubKey{
		ID:           "KEY#1",
		Type:         "Ed25519",
		PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
	}
	doc.PublicKey = []PubKey{pk}
	auth := Auth{
		PublicKey: []string{"KEY#1"},
	}
	doc.Authentication = []Auth{auth}
	return doc
}

func genesisDIDDoc() *DIDDoc {
	return &DIDDoc{
		BasicDoc: BasicDoc{
			ID:   "did:bitxhub:appchain001:0x00000001",
			Type: "user",
			PublicKey: []PubKey{
				{
					ID:           "KEY#1",
					Type:         "Ed25519",
					PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
				},
			},
			Authentication: []Auth{
				{PublicKey: []string{"KEY#1"}},
			},
		},
	}
}

func genesisMetohdDoc() *MethodDoc {
	return &MethodDoc{
		BasicDoc: BasicDoc{
			ID:   "did:bitxhub:relayroot:.",
			Type: "method",
			PublicKey: []PubKey{
				{
					ID:           "KEY#1",
					Type:         "Secp256k1",
					PublicKeyPem: "02b97c30de767f084ce3080168ee293053ba33b235d7116a3263d29f1450936b71",
				},
			},
			Controller: DID("did:bitxhub:relayroot:0x00000001"),
			Authentication: []Auth{
				{PublicKey: []string{"KEY#1"}},
			},
		},
	}
}

// UnmarshalConfig .
func UnmarshalConfig(repoRoot string) (*Config, error) {
	viper.SetConfigFile(filepath.Join(repoRoot, configName))
	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("BITXID")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config, err := DefaultConfig()
	if err != nil {
		fmt.Println("config DefaultConfig err", err)
		return nil, err
	}

	if err := viper.Unmarshal(config); err != nil {
		fmt.Println("config Unmarshal err", err)
		return nil, err
	}

	return config, nil
}
