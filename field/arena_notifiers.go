// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains configuration of the publish-subscribe notifiers that allow the arena to push updates to websocket clients.

package field

import (
	"strconv"

	"github.com/BotDogs4645/da/bracket"
	"github.com/BotDogs4645/da/game"
	"github.com/BotDogs4645/da/model"
	"github.com/BotDogs4645/da/network"
	"github.com/BotDogs4645/da/websocket"
)

type ArenaNotifiers struct {
	AllianceSelectionNotifier          *websocket.Notifier
	AllianceStationDisplayModeNotifier *websocket.Notifier
	ArenaStatusNotifier                *websocket.Notifier
	AudienceDisplayModeNotifier        *websocket.Notifier
	DisplayConfigurationNotifier       *websocket.Notifier
	EventStatusNotifier                *websocket.Notifier
	LowerThirdNotifier                 *websocket.Notifier
	MatchLoadNotifier                  *websocket.Notifier
	MatchTimeNotifier                  *websocket.Notifier
	MatchTimingNotifier                *websocket.Notifier
	PlaySoundNotifier                  *websocket.Notifier
	RealtimeScoreNotifier              *websocket.Notifier
	ReloadDisplaysNotifier             *websocket.Notifier
	ScorePostedNotifier                *websocket.Notifier
}

type MatchTimeMessage struct {
	MatchState
	MatchTimeSec int
}

type audienceAllianceScoreFields struct {
	Score        *game.Score
	ScoreSummary *game.ScoreSummary
}

// Instantiates notifiers and configures their message producing methods.
func (arena *Arena) configureNotifiers() {
	arena.AllianceSelectionNotifier = websocket.NewNotifier("allianceSelection", arena.generateAllianceSelectionMessage)
	arena.AllianceStationDisplayModeNotifier = websocket.NewNotifier("allianceStationDisplayMode",
		arena.generateAllianceStationDisplayModeMessage)
	arena.ArenaStatusNotifier = websocket.NewNotifier("arenaStatus", arena.generateArenaStatusMessage)
	arena.AudienceDisplayModeNotifier = websocket.NewNotifier("audienceDisplayMode",
		arena.generateAudienceDisplayModeMessage)
	arena.DisplayConfigurationNotifier = websocket.NewNotifier("displayConfiguration",
		arena.generateDisplayConfigurationMessage)
	arena.EventStatusNotifier = websocket.NewNotifier("eventStatus", arena.generateEventStatusMessage)
	arena.LowerThirdNotifier = websocket.NewNotifier("lowerThird", arena.generateLowerThirdMessage)
	arena.MatchLoadNotifier = websocket.NewNotifier("matchLoad", arena.generateMatchLoadMessage)
	arena.MatchTimeNotifier = websocket.NewNotifier("matchTime", arena.generateMatchTimeMessage)
	arena.MatchTimingNotifier = websocket.NewNotifier("matchTiming", arena.generateMatchTimingMessage)
	arena.PlaySoundNotifier = websocket.NewNotifier("playSound", nil)
	arena.RealtimeScoreNotifier = websocket.NewNotifier("realtimeScore", arena.generateRealtimeScoreMessage)
	arena.ReloadDisplaysNotifier = websocket.NewNotifier("reload", nil)
	arena.ScorePostedNotifier = websocket.NewNotifier("scorePosted", arena.generateScorePostedMessage)
}

func (arena *Arena) generateAllianceSelectionMessage() interface{} {
	return &arena.AllianceSelectionAlliances
}

func (arena *Arena) generateAllianceStationDisplayModeMessage() interface{} {
	return arena.AllianceStationDisplayMode
}

func (arena *Arena) generateArenaStatusMessage() interface{} {
	// Convert AP team wifi network status array to a map by station for ease of client use.
	teamWifiStatuses := make(map[string]network.TeamWifiStatus)
	for i, station := range []string{"R1", "R2", "R3", "B1", "B2", "B3"} {
		if arena.EventSettings.Ap2TeamChannel == 0 || i < 3 {
			teamWifiStatuses[station] = arena.accessPoint.TeamWifiStatuses[i]
		} else {
			teamWifiStatuses[station] = arena.accessPoint2.TeamWifiStatuses[i]
		}
	}

	return &struct {
		MatchId          int
		AllianceStations map[string]*AllianceStation
		TeamWifiStatuses map[string]network.TeamWifiStatus
		MatchState
		CanStartMatch         bool
		PlcIsHealthy          bool
		FieldEstop            bool
		PlcArmorBlockStatuses map[string]bool
	}{arena.CurrentMatch.Id, arena.AllianceStations, teamWifiStatuses, arena.MatchState,
		arena.checkCanStartMatch() == nil, arena.Plc.IsHealthy, arena.Plc.GetFieldEstop(),
		arena.Plc.GetArmorBlockStatuses()}
}

func (arena *Arena) generateAudienceDisplayModeMessage() interface{} {
	return arena.AudienceDisplayMode
}

func (arena *Arena) generateDisplayConfigurationMessage() interface{} {
	// Notify() for this notifier must always called from a method that has a lock on the display mutex.
	// Make a copy of the map to avoid potential data races; otherwise the same map would get iterated through as it is
	// serialized to JSON, outside the mutex lock.
	displaysCopy := make(map[string]Display)
	for displayId, display := range arena.Displays {
		displaysCopy[displayId] = *display
	}
	return displaysCopy
}

func (arena *Arena) generateEventStatusMessage() interface{} {
	return arena.EventStatus
}

func (arena *Arena) generateLowerThirdMessage() interface{} {
	return &struct {
		LowerThird     *model.LowerThird
		ShowLowerThird bool
	}{arena.LowerThird, arena.ShowLowerThird}
}

func (arena *Arena) generateMatchLoadMessage() interface{} {
	teams := make(map[string]*model.Team)
	for station, allianceStation := range arena.AllianceStations {
		teams[station] = allianceStation.Team
	}

	rankings := make(map[string]*game.Ranking)
	for _, allianceStation := range arena.AllianceStations {
		if allianceStation.Team != nil {
			rankings[strconv.Itoa(allianceStation.Team.Id)], _ =
				arena.Database.GetRankingForTeam(allianceStation.Team.Id)
		}
	}

	var matchup *bracket.Matchup
	var redOffFieldTeams []*model.Team
	var blueOffFieldTeams []*model.Team
	if arena.CurrentMatch.Type == "elimination" {
		matchup, _ = arena.PlayoffBracket.GetMatchup(arena.CurrentMatch.ElimRound, arena.CurrentMatch.ElimGroup)
		redOffFieldTeamIds, blueOffFieldTeamIds, _ := arena.Database.GetOffFieldTeamIds(arena.CurrentMatch)
		for _, teamId := range redOffFieldTeamIds {
			team, _ := arena.Database.GetTeamById(teamId)
			redOffFieldTeams = append(redOffFieldTeams, team)
		}
		for _, teamId := range blueOffFieldTeamIds {
			team, _ := arena.Database.GetTeamById(teamId)
			blueOffFieldTeams = append(blueOffFieldTeams, team)
		}
	}

	return &struct {
		MatchType         string
		Match             *model.Match
		Teams             map[string]*model.Team
		Rankings          map[string]*game.Ranking
		Matchup           *bracket.Matchup
		RedOffFieldTeams  []*model.Team
		BlueOffFieldTeams []*model.Team
	}{
		arena.CurrentMatch.CapitalizedType(),
		arena.CurrentMatch,
		teams,
		rankings,
		matchup,
		redOffFieldTeams,
		blueOffFieldTeams,
	}
}

func (arena *Arena) generateMatchTimeMessage() interface{} {
	return MatchTimeMessage{arena.MatchState, int(arena.MatchTimeSec())}
}

func (arena *Arena) generateMatchTimingMessage() interface{} {
	return &game.MatchTiming
}

func (arena *Arena) generateRealtimeScoreMessage() interface{} {
	fields := struct {
		Red  *audienceAllianceScoreFields
		Blue *audienceAllianceScoreFields
		MatchState
	}{}
	fields.Red = getAudienceAllianceScoreFields(arena.RedScore, arena.RedScoreSummary())
	fields.Blue = getAudienceAllianceScoreFields(arena.BlueScore, arena.BlueScoreSummary())
	fields.MatchState = arena.MatchState
	return &fields
}

func (arena *Arena) generateScorePostedMessage() interface{} {
	// For elimination matches, summarize the state of the series.
	var seriesStatus, seriesLeader string
	var matchup *bracket.Matchup
	if arena.SavedMatch.Type == "elimination" {
		matchup, _ = arena.PlayoffBracket.GetMatchup(arena.SavedMatch.ElimRound, arena.SavedMatch.ElimGroup)
		seriesLeader, seriesStatus = matchup.StatusText()
	}

	rankings := make(map[int]game.Ranking, len(arena.SavedRankings))
	for _, ranking := range arena.SavedRankings {
		rankings[ranking.TeamId] = ranking
	}

	return &struct {
		MatchType        string
		Match            *model.Match
		RedScoreSummary  *game.ScoreSummary
		BlueScoreSummary *game.ScoreSummary
		Rankings         map[int]game.Ranking
		SeriesStatus     string
		SeriesLeader     string
	}{
		arena.SavedMatch.CapitalizedType(),
		arena.SavedMatch,
		arena.SavedMatchResult.RedScoreSummary(),
		arena.SavedMatchResult.BlueScoreSummary(),
		rankings,
		seriesStatus,
		seriesLeader,
	}
}

// Constructs the data object for one alliance sent to the audience display for the realtime scoring overlay.
func getAudienceAllianceScoreFields(allianceScore *game.Score,
	allianceScoreSummary *game.ScoreSummary) *audienceAllianceScoreFields {
	fields := new(audienceAllianceScoreFields)
	fields.Score = allianceScore
	fields.ScoreSummary = allianceScoreSummary
	return fields
}
