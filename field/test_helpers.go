// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package field

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/BotDogs4645/da/game"
	"github.com/BotDogs4645/da/model"
	"github.com/stretchr/testify/assert"
)

func SetupTestArena(t *testing.T, uniqueName string) *Arena {
	model.BaseDir = ".."
	dbPath := filepath.Join(model.BaseDir, fmt.Sprintf("%s_test.db", uniqueName))
	os.Remove(dbPath)
	arena, err := NewArena(dbPath)
	assert.Nil(t, err)
	return arena
}

func setupTestArena(t *testing.T) *Arena {
	game.MatchTiming.WarmupDurationSec = 3
	game.MatchTiming.PauseDurationSec = 2
	return SetupTestArena(t, "field")
}
