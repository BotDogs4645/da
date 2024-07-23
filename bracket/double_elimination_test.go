// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package bracket

import (
	"github.com/Team254/cheesy-arena-lite/game"
	"github.com/Team254/cheesy-arena-lite/tournament"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoubleEliminationInitial(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 8)
	bracket, err := NewDoubleEliminationBracket(8)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err := database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 4, len(matches)) {
		assertMatch(t, matches[0], "1", 1, 8)
		assertMatch(t, matches[1], "2", 4, 5)
		assertMatch(t, matches[2], "3", 2, 7)
		assertMatch(t, matches[3], "4", 3, 6)
	}
}

func TestDoubleEliminationErrors(t *testing.T) {
	_, err := NewDoubleEliminationBracket(7)
	if assert.NotNil(t, err) {
		assert.Equal(t, "Must have exactly 8 alliances", err.Error())
	}

	_, err = NewDoubleEliminationBracket(9)
	if assert.NotNil(t, err) {
		assert.Equal(t, "Must have exactly 8 alliances", err.Error())
	}
}

func TestDoubleEliminationProgression(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 8)
	bracket, err := NewDoubleEliminationBracket(8)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err := database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 4, len(matches))

	scoreMatch(database, "1", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 4, len(matches))

	scoreMatch(database, "2", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[4], "5", 1, 5)
		assertMatch(t, matches[5], "7", 8, 4)
	}

	scoreMatch(database, "3", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 6, len(matches))

	scoreMatch(database, "4", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 8, len(matches)) {
		assertMatch(t, matches[4], "5", 1, 5)
		assertMatch(t, matches[5], "6", 2, 6)
		assertMatch(t, matches[6], "7", 8, 4)
		assertMatch(t, matches[7], "8", 7, 3)
	}

	scoreMatch(database, "5", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 8, len(matches))

	scoreMatch(database, "6", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 8, len(matches))

	scoreMatch(database, "7", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 9, len(matches)) {
		assertMatch(t, matches[8], "9", 8, 2)
	}

	scoreMatch(database, "8", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 11, len(matches)) {
		assertMatch(t, matches[9], "10", 7, 5)
		assertMatch(t, matches[10], "11", 4, 3)
	}

	scoreMatch(database, "9", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 11, len(matches))

	scoreMatch(database, "10", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 12, len(matches)) {
		assertMatch(t, matches[10], "11", 4, 3)
		assertMatch(t, matches[11], "12", 7, 8)
	}

	scoreMatch(database, "11", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 12, len(matches))

	scoreMatch(database, "12", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 13, len(matches)) {
		assertMatch(t, matches[12], "13", 3, 7)
	}

	scoreMatch(database, "13", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 15, len(matches)) {
		assertMatch(t, matches[13], "F-1", 4, 7)
		assertMatch(t, matches[14], "F-2", 4, 7)
	}
	assert.False(t, bracket.IsComplete())
	assert.Equal(t, 0, bracket.Winner())
	assert.Equal(t, 0, bracket.Finalist())

	scoreMatch(database, "F-1", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 15, len(matches))
	assert.False(t, bracket.IsComplete())
	assert.Equal(t, 0, bracket.Winner())
	assert.Equal(t, 0, bracket.Finalist())

	scoreMatch(database, "F-2", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 16, len(matches)) {
		assertMatch(t, matches[15], "F-3", 4, 7)
	}
	assert.False(t, bracket.IsComplete())
	assert.Equal(t, 0, bracket.Winner())
	assert.Equal(t, 0, bracket.Finalist())

	scoreMatch(database, "F-3", game.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 17, len(matches)) {
		assertMatch(t, matches[16], "F-4", 4, 7)
	}
	assert.False(t, bracket.IsComplete())
	assert.Equal(t, 0, bracket.Winner())
	assert.Equal(t, 0, bracket.Finalist())

	scoreMatch(database, "F-4", game.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 18, len(matches)) {
		assertMatch(t, matches[17], "F-5", 4, 7)
	}
	assert.False(t, bracket.IsComplete())
	assert.Equal(t, 0, bracket.Winner())
	assert.Equal(t, 0, bracket.Finalist())

	scoreMatch(database, "F-5", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 18, len(matches))
	assert.True(t, bracket.IsComplete())
	assert.Equal(t, 7, bracket.Winner())
	assert.Equal(t, 4, bracket.Finalist())
}

func TestDoubleEliminationTie(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 8)
	bracket, err := NewDoubleEliminationBracket(8)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err := database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 4, len(matches))

	scoreMatch(database, "1", game.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 5, len(matches)) {
		assertMatch(t, matches[4], "1-2", 1, 8)
	}

	scoreMatch(database, "1-2", game.TieMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[5], "1-3", 1, 8)
	}

	scoreMatch(database, "1-3", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 6, len(matches))
}

func TestDoubleEliminationChangeResult(t *testing.T) {
	database := setupTestDb(t)

	tournament.CreateTestAlliances(database, 8)
	bracket, err := NewDoubleEliminationBracket(8)
	assert.Nil(t, err)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err := database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 4, len(matches))

	scoreMatch(database, "1", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	assert.Equal(t, 4, len(matches))

	scoreMatch(database, "2", game.RedWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[4], "5", 1, 5)
		assertMatch(t, matches[5], "7", 8, 4)
	}

	scoreMatch(database, "2", game.MatchNotPlayed)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Equal(t, 4, len(matches))

	scoreMatch(database, "2", game.BlueWonMatch)
	assert.Nil(t, bracket.Update(database, &dummyStartTime))
	matches, err = database.GetMatchesByType("elimination")
	assert.Nil(t, err)
	if assert.Equal(t, 6, len(matches)) {
		assertMatch(t, matches[4], "5", 1, 4)
		assertMatch(t, matches[5], "7", 8, 5)
	}
}
