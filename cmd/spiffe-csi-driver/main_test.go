package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewZapLogger(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		log, err := newZapLogger("text")
		require.NoError(t, err)
		require.NotNil(t, log)
	})

	t.Run("json", func(t *testing.T) {
		log, err := newZapLogger("json")
		require.NoError(t, err)
		require.NotNil(t, log)
	})

	t.Run("invalid", func(t *testing.T) {
		log, err := newZapLogger("invalid")
		require.Error(t, err)
		require.Nil(t, log)
	})
}
