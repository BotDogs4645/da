// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for the results (score and fouls) from a match at an event.

package model

import (
	"github.com/Team254/cheesy-arena-lite/game"
)

type MatchResult struct {
	Id         int `db:"id"`
	MatchId    int
	PlayNumber int
	MatchType  string
	RedScore   *game.Score
	BlueScore  *game.Score
}

// Returns a new match result object with empty slices instead of nil.
func NewMatchResult() *MatchResult {
	matchResult := new(MatchResult)
	matchResult.RedScore = new(game.Score)
	matchResult.BlueScore = new(game.Score)
	return matchResult
}

func (database *Database) CreateMatchResult(matchResult *MatchResult) error {
	return database.matchResultTable.create(matchResult)
}

func (database *Database) GetMatchResultForMatch(matchId int) (*MatchResult, error) {
	matchResults, err := database.matchResultTable.getAll()
	if err != nil {
		return nil, err
	}

	var mostRecentMatchResult *MatchResult
	for i, matchResult := range matchResults {
		if matchResult.MatchId == matchId &&
			(mostRecentMatchResult == nil || matchResult.PlayNumber > mostRecentMatchResult.PlayNumber) {
			mostRecentMatchResult = &matchResults[i]
		}
	}
	return mostRecentMatchResult, nil
}

func (database *Database) UpdateMatchResult(matchResult *MatchResult) error {
	return database.matchResultTable.update(matchResult)
}

func (database *Database) DeleteMatchResult(id int) error {
	return database.matchResultTable.delete(id)
}

func (database *Database) TruncateMatchResults() error {
	return database.matchResultTable.truncate()
}

// Calculates and returns the summary fields used for ranking and display for the red alliance.
func (matchResult *MatchResult) RedScoreSummary() *game.ScoreSummary {
	return matchResult.RedScore.Summarize()
}

// Calculates and returns the summary fields used for ranking and display for the blue alliance.
func (matchResult *MatchResult) BlueScoreSummary() *game.ScoreSummary {
	return matchResult.BlueScore.Summarize()
}
