// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package model

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/BotDogs4645/da/game"
	"github.com/stretchr/testify/assert"
)

func SetupTestDb(t *testing.T, uniqueName string) *Database {
	BaseDir = ".."
	dbPath := filepath.Join(BaseDir, fmt.Sprintf("%s_test.db", uniqueName))
	os.Remove(dbPath)
	database, err := OpenDatabase(dbPath)
	assert.Nil(t, err)
	return database
}

func BuildTestMatchResult(matchId int, playNumber int) *MatchResult {
	matchResult := &MatchResult{MatchId: matchId, PlayNumber: playNumber, MatchType: "qualification"}
	matchResult.RedScore = game.TestScore1()
	matchResult.BlueScore = game.TestScore2()
	return matchResult
}

func BuildTestAlliances(database *Database) {
	database.CreateAlliance(&Alliance{Id: 2, TeamIds: []int{1718, 2451, 1619}, Lineup: [3]int{2451, 1718, 1619}})
	database.CreateAlliance(&Alliance{Id: 1, TeamIds: []int{254, 469, 2848, 74, 3175}, Lineup: [3]int{469, 254, 2848}})
}
