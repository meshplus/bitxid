package repo

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
	Addr           string `toml:"addr" json:"addr"`
	GenesisAdmin   string `toml:"genesis_admin" json:"genesis_admin"`
	GenesisAccount string `toml:"genesis_account" json:"genesis_account"`
	GenesisDoc     string `toml:"genesis_doc" json:"genesis_doc"`
}

// MethodConfig .
type MethodConfig struct {
	Addr          string `toml:"addr" json:"addr"`
	IsRoot        bool   `toml:"is_root" json:"is_root"`
	GenesisMetohd string `toml:"genesis_metohd" json:"genesis_metohd"`
	GenesisAdmin  string `toml:"genesis_admin" json:"genesis_admin"`
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
			Addr:           ".",
			GenesisAdmin:   "did:bitxhub:relayroot:0x12348848",
			GenesisAccount: "0x12348848",
			GenesisDoc:     "",
		},
		MethodConfig: MethodConfig{
			Addr:          ".",
			IsRoot:        true,
			GenesisMetohd: "did:bitxhub:relayroot:.",
			GenesisAdmin:  "did:bitxhub:relayroot:0x01",
		},
	}, nil
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
