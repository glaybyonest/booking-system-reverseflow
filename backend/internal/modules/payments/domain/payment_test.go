package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidForceStatus(t *testing.T) {
	require.True(t, ValidForceStatus(""))
	require.True(t, ValidForceStatus(StatusSucceeded))
	require.True(t, ValidForceStatus(StatusFailed))
	require.False(t, ValidForceStatus("processing"))
}
