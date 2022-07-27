package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("config", func(t *testing.T) {
		_, err := NewConfig("/very/wrong/path.conf")
		require.ErrorIs(t, err, ErrConfigRead, "Error must be: %q, actual: %q", ErrConfigRead, err)
	})
}
