package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalConfig(t *testing.T) {
	Config, err := UnmarshalConfig("/Users/lousong/Dev/src/bitxid/config")
	assert.Nil(t, err)
	DConfig, err := DefaultBitXIDConfig()
	assert.Equal(t, *DConfig, Config.BitXIDConfig)
}
