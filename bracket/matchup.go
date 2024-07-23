// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Models and logic encapsulating a group of one or more matches between the same two alliances at a given point in a
// playoff tournament.

package bracket

import (
	"fmt"
	"github.com/Team254/cheesy-arena-lite/game"
	"github.com/Team254/cheesy-arena-lite/model"
	"strconv"
)

// Conveys how a given alliance should be populated -- either directly from alliance selection or based on the results
// of a prior matchup.
type allianceSource struct {
	allianceId int
	matchupKey matchupKey
	useWinner  bool
}

// Key for uniquely identifying a matchup. Round IDs are arbitrary and in descending order with "1" always representing
// the playoff finals. Group IDs are 1-indexed within a round and increasing in order of play.
type matchupKey struct {
	Round int
	Group int
}

// Conveys the complete generic information about a matchup required to construct it. In aggregate, the full list of
// match templates describing a bracket format can be used to construct an empty playoff bracket for a given number of
// alliances.
type matchupTemplate struct {
	matchupKey
	displayName        string
	NumWinsToAdvance   int
	redAllianceSource  allianceSource
	blueAllianceSource allianceSource
}

// Encapsulates the format and state of a group of one or more matches between the same two alliances at a given point
// in a playoff tournament.
type Matchup struct {
	matchupTemplate
	redAllianceSourceMatchup  *Matchup
	blueAllianceSourceMatchup *Matchup
	RedAllianceId             int
	BlueAllianceId            int
	RedAllianceWins           int
	BlueAllianceWins          int
}

// Convenience method to quickly create an alliance source that points to the winner of a different matchup.
func newWinnerAllianceSource(round, group int) allianceSource {
	return allianceSource{matchupKey: newMatchupKey(round, group), useWinner: true}
}

// Convenience method to quickly create an alliance source that points to the loser of a different matchup.
func newLoserAllianceSource(round, group int) allianceSource {
	return allianceSource{matchupKey: newMatchupKey(round, group), useWinner: false}
}

// Convenience method to quickly create a matchup key.
func newMatchupKey(round, group int) matchupKey {
	return matchupKey{Round: round, Group: group}
}

// Returns the display name for a specific match within a matchup.
func (matchupTemplate *matchupTemplate) matchDisplayName(instance int) string {
	displayName := matchupTemplate.displayName
	if matchupTemplate.NumWinsToAdvance > 1 || instance > 1 {
		// Append the instance if there is always more than one match in the series, or in exceptional circumstances
		// like a tie in double-elimination unresolved by tiebreakers.
		displayName += fmt.Sprintf("-%d", instance)
	}
	return displayName
}

// Returns the display name for the overall matchup.
func (matchup *Matchup) LongDisplayName() string {
	if matchup.isFinal() {
		return "Finals"
	}
	if _, err := strconv.Atoi(matchup.displayName); err == nil {
		return "Match " + matchup.displayName
	}
	return matchup.displayName
}

// Returns the display name for the linked matchup from which the red alliance is populated.
func (matchup *Matchup) RedAllianceSourceDisplayName() string {
	if matchup.redAllianceSourceMatchup == nil {
		return ""
	}
	if matchup.redAllianceSource.useWinner {
		return "W " + matchup.redAllianceSourceMatchup.displayName
	}
	return "L " + matchup.redAllianceSourceMatchup.displayName
}

// Returns the display name for the linked matchup from which the blue alliance is populated.
func (matchup *Matchup) BlueAllianceSourceDisplayName() string {
	if matchup.blueAllianceSourceMatchup == nil {
		return ""
	}
	if matchup.blueAllianceSource.useWinner {
		return "W " + matchup.blueAllianceSourceMatchup.displayName
	}
	return "L " + matchup.blueAllianceSourceMatchup.displayName
}

// Returns a pair of strings indicating the leading alliance and a readable status of the matchup.
func (matchup *Matchup) StatusText() (string, string) {
	var leader, status string
	winText := "Advances"
	if matchup.isFinal() {
		winText = "Wins"
	}
	if matchup.RedAllianceWins >= matchup.NumWinsToAdvance {
		leader = "red"
		status = fmt.Sprintf("Red %s %d-%d", winText, matchup.RedAllianceWins, matchup.BlueAllianceWins)
	} else if matchup.BlueAllianceWins >= matchup.NumWinsToAdvance {
		leader = "blue"
		status = fmt.Sprintf("Blue %s %d-%d", winText, matchup.BlueAllianceWins, matchup.RedAllianceWins)
	} else if matchup.RedAllianceWins > matchup.BlueAllianceWins {
		leader = "red"
		status = fmt.Sprintf("Red Leads %d-%d", matchup.RedAllianceWins, matchup.BlueAllianceWins)
	} else if matchup.BlueAllianceWins > matchup.RedAllianceWins {
		leader = "blue"
		status = fmt.Sprintf("Blue Leads %d-%d", matchup.BlueAllianceWins, matchup.RedAllianceWins)
	} else if matchup.RedAllianceWins > 0 {
		status = fmt.Sprintf("Series Tied %d-%d", matchup.RedAllianceWins, matchup.BlueAllianceWins)
	}
	return leader, status
}

// Returns the winning alliance ID of the matchup, or 0 if it is not yet known.
func (matchup *Matchup) Winner() int {
	if matchup.RedAllianceWins >= matchup.NumWinsToAdvance {
		return matchup.RedAllianceId
	}
	if matchup.BlueAllianceWins >= matchup.NumWinsToAdvance {
		return matchup.BlueAllianceId
	}
	return 0
}

// Returns the losing alliance ID of the matchup, or 0 if it is not yet known.
func (matchup *Matchup) Loser() int {
	if matchup.RedAllianceWins >= matchup.NumWinsToAdvance {
		return matchup.BlueAllianceId
	}
	if matchup.BlueAllianceWins >= matchup.NumWinsToAdvance {
		return matchup.RedAllianceId
	}
	return 0
}

// Returns true if the matchup has been won, and false if it is still to be determined.
func (matchup *Matchup) IsComplete() bool {
	return matchup.Winner() > 0
}

// Returns true if the matchup represents the final matchup in the bracket.
func (matchup *Matchup) isFinal() bool {
	return matchup.displayName == "F"
}

// Recursively traverses the matchup graph to update the state of this matchup and all of its children based on match
// results, counting wins and creating or deleting matches as required.
func (matchup *Matchup) update(database *model.Database) error {
	// Update child matchups first. Only recurse down winner links to avoid visiting a node twice.
	if matchup.redAllianceSourceMatchup != nil && matchup.redAllianceSource.useWinner {
		if err := matchup.redAllianceSourceMatchup.update(database); err != nil {
			return err
		}
	}
	if matchup.blueAllianceSourceMatchup != nil && matchup.blueAllianceSource.useWinner {
		if err := matchup.blueAllianceSourceMatchup.update(database); err != nil {
			return err
		}
	}

	// Populate the alliance IDs from the lower matchups (or with a zero value if they are not yet complete).
	if matchup.redAllianceSourceMatchup != nil {
		if matchup.redAllianceSource.useWinner {
			matchup.RedAllianceId = matchup.redAllianceSourceMatchup.Winner()
		} else {
			matchup.RedAllianceId = matchup.redAllianceSourceMatchup.Loser()
		}
	}
	if matchup.blueAllianceSourceMatchup != nil {
		if matchup.blueAllianceSource.useWinner {
			matchup.BlueAllianceId = matchup.blueAllianceSourceMatchup.Winner()
		} else {
			matchup.BlueAllianceId = matchup.blueAllianceSourceMatchup.Loser()
		}
	}

	matches, err := database.GetMatchesByElimRoundGroup(matchup.Round, matchup.Group)
	if err != nil {
		return err
	}

	// Bail if we do not yet know both alliances.
	if matchup.RedAllianceId == 0 || matchup.BlueAllianceId == 0 {
		// Ensure the current state is reset; it may have previously been populated if a match result was edited.
		matchup.RedAllianceWins = 0
		matchup.BlueAllianceWins = 0

		// Delete any previously created matches.
		for _, match := range matches {
			if err = database.DeleteMatch(match.Id); err != nil {
				return err
			}
		}

		return nil
	}

	// Create, update, and/or delete unplayed matches as required.
	redAlliance, err := database.GetAllianceById(matchup.RedAllianceId)
	if err != nil {
		return err
	}
	if redAlliance == nil {
		return fmt.Errorf("alliance %d does not exist in the database", matchup.RedAllianceId)
	}
	blueAlliance, err := database.GetAllianceById(matchup.BlueAllianceId)
	if err != nil {
		return err
	}
	if blueAlliance == nil {
		return fmt.Errorf("alliance %d does not exist in the database", matchup.BlueAllianceId)
	}
	matchup.RedAllianceWins = 0
	matchup.BlueAllianceWins = 0
	var unplayedMatches []model.Match
	for _, match := range matches {
		if !match.IsComplete() {
			// Update the teams in the match if they are not yet set or are incorrect.
			changed := false
			if match.Red1 != redAlliance.Lineup[0] || match.Red2 != redAlliance.Lineup[1] ||
				match.Red3 != redAlliance.Lineup[2] {
				positionRedTeams(&match, redAlliance)
				match.ElimRedAlliance = redAlliance.Id
				changed = true
				if err = database.UpdateMatch(&match); err != nil {
					return err
				}
			}
			if match.Blue1 != blueAlliance.Lineup[0] || match.Blue2 != blueAlliance.Lineup[1] ||
				match.Blue3 != blueAlliance.Lineup[2] {
				positionBlueTeams(&match, blueAlliance)
				match.ElimBlueAlliance = blueAlliance.Id
				changed = true
			}
			if changed {
				if err = database.UpdateMatch(&match); err != nil {
					return err
				}
			}

			unplayedMatches = append(unplayedMatches, match)
			continue
		}

		// Check who won.
		if match.Status == game.RedWonMatch {
			matchup.RedAllianceWins++
		} else if match.Status == game.BlueWonMatch {
			matchup.BlueAllianceWins++
		}
	}

	maxWins := matchup.RedAllianceWins
	if matchup.BlueAllianceWins > maxWins {
		maxWins = matchup.BlueAllianceWins
	}
	numUnplayedMatchesNeeded := matchup.NumWinsToAdvance - maxWins
	if len(unplayedMatches) > numUnplayedMatchesNeeded {
		// Delete any superfluous matches off the end of the list.
		for i := 0; i < len(unplayedMatches)-numUnplayedMatchesNeeded; i++ {
			if err = database.DeleteMatch(unplayedMatches[len(unplayedMatches)-i-1].Id); err != nil {
				return err
			}
		}
	} else if len(unplayedMatches) < numUnplayedMatchesNeeded {
		// Create initial set of matches or any additional required matches due to tie matches or ties in the round.
		for i := 0; i < numUnplayedMatchesNeeded-len(unplayedMatches); i++ {
			instance := len(matches) + i + 1
			match := model.Match{
				Type:             "elimination",
				DisplayName:      matchup.matchDisplayName(instance),
				ElimRound:        matchup.Round,
				ElimGroup:        matchup.Group,
				ElimInstance:     instance,
				ElimRedAlliance:  redAlliance.Id,
				ElimBlueAlliance: blueAlliance.Id,
			}
			positionRedTeams(&match, redAlliance)
			positionBlueTeams(&match, blueAlliance)
			if err = database.CreateMatch(&match); err != nil {
				return err
			}
		}
	}

	return nil
}

// Assigns the lineup from the alliance into the red team slots for the match.
func positionRedTeams(match *model.Match, alliance *model.Alliance) {
	match.Red1 = alliance.Lineup[0]
	match.Red2 = alliance.Lineup[1]
	match.Red3 = alliance.Lineup[2]
}

// Assigns the lineup from the alliance into the blue team slots for the match.
func positionBlueTeams(match *model.Match, alliance *model.Alliance) {
	match.Blue1 = alliance.Lineup[0]
	match.Blue2 = alliance.Lineup[1]
	match.Blue3 = alliance.Lineup[2]
}
