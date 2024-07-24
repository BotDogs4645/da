package web

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/BotDogs4645/da/field"
	"github.com/BotDogs4645/da/game"
	"github.com/BotDogs4645/da/model"
	"github.com/BotDogs4645/da/websocket"
	"github.com/mitchellh/mapstructure"
)

func (web *Web) scoringPanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	practiceMatches, err := web.buildMatchPlayList("practice")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	qualificationMatches, err := web.buildMatchPlayList("qualification")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	eliminationMatches, err := web.buildMatchPlayList("elimination")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	template, err := web.parseFiles("templates/scoring_panel.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	matchesByType := map[string]MatchPlayList{"practice": practiceMatches,
		"qualification": qualificationMatches, "elimination": eliminationMatches}
	currentMatchType := web.arena.CurrentMatch.Type
	if currentMatchType == "test" {
		currentMatchType = "practice"
	}
	redOffFieldTeams, blueOffFieldTeams, err := web.arena.Database.GetOffFieldTeamIds(web.arena.CurrentMatch)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	matchResult, err := web.arena.Database.GetMatchResultForMatch(web.arena.CurrentMatch.Id)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	isReplay := matchResult != nil
	data := struct {
		*model.EventSettings
		PlcIsEnabled          bool
		MatchesByType         map[string]MatchPlayList
		CurrentMatchType      string
		Match                 *model.Match
		RedOffFieldTeams      []int
		BlueOffFieldTeams     []int
		RedScore              *game.Score
		BlueScore             *game.Score
		AllowSubstitution     bool
		IsReplay              bool
		SavedMatchType        string
		SavedMatch            *model.Match
		PlcArmorBlockStatuses map[string]bool
	}{
		web.arena.EventSettings,
		web.arena.Plc.IsEnabled(),
		matchesByType,
		currentMatchType,
		web.arena.CurrentMatch,
		redOffFieldTeams,
		blueOffFieldTeams,
		web.arena.RedScore,
		web.arena.BlueScore,
		web.arena.CurrentMatch.ShouldAllowSubstitution(),
		isReplay,
		web.arena.SavedMatch.CapitalizedType(),
		web.arena.SavedMatch,
		web.arena.Plc.GetArmorBlockStatuses(),
	}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

func (web *Web) scoringPanelWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.MatchTimingNotifier, web.arena.ArenaStatusNotifier, web.arena.MatchTimeNotifier,
		web.arena.RealtimeScoreNotifier, web.arena.AudienceDisplayModeNotifier,
		web.arena.AllianceStationDisplayModeNotifier, web.arena.EventStatusNotifier)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		switch messageType {
		case "substituteTeam":
			args := struct {
				Team     int
				Position string
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.arena.SubstituteTeam(args.Team, args.Position)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		case "toggleBypass":
			station, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			if _, ok := web.arena.AllianceStations[station]; !ok {
				ws.WriteError(fmt.Sprintf("Invalid alliance station '%s'.", station))
				continue
			}
			web.arena.AllianceStations[station].Bypass = !web.arena.AllianceStations[station].Bypass
		case "startMatch":
			args := struct {
				MuteMatchSounds bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			web.arena.MuteMatchSounds = args.MuteMatchSounds
			err = web.arena.StartMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		case "abortMatch":
			err = web.arena.AbortMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		case "signalVolunteers":
			if web.arena.MatchState != field.PostMatch {
				// Don't allow clearing the field until the match is over.
				continue
			}
			web.arena.FieldVolunteers = true
			continue // Don't reload.
		case "signalReset":
			if web.arena.MatchState != field.PostMatch {
				// Don't allow clearing the field until the match is over.
				continue
			}
			web.arena.FieldReset = true
			web.arena.AllianceStationDisplayMode = "fieldReset"
			web.arena.AllianceStationDisplayModeNotifier.Notify()
			continue // Don't reload.
		case "commitResults":
			err = web.commitCurrentMatchScore()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.arena.ResetMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.arena.LoadNextMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = ws.WriteNotifier(web.arena.ReloadDisplaysNotifier)
			if err != nil {
				log.Println(err)
				return
			}
			continue // Skip sending the status update, as the client is about to terminate and reload.
		case "discardResults":
			err = web.arena.ResetMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.arena.LoadNextMatch()
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = ws.WriteNotifier(web.arena.ReloadDisplaysNotifier)
			if err != nil {
				log.Println(err)
				return
			}
			continue // Skip sending the status update, as the client is about to terminate and reload.
		case "setAudienceDisplay":
			mode, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			web.arena.SetAudienceDisplayMode(mode)
			continue
		case "setAllianceStationDisplay":
			mode, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			web.arena.SetAllianceStationDisplayMode(mode)
			continue
		case "startTimeout":
			durationSec, ok := data.(float64)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			err = web.arena.StartTimeout(int(durationSec))
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		case "setTestMatchName":
			if web.arena.CurrentMatch.Type != "test" {
				// Don't allow changing the name of a non-test match.
				continue
			}
			name, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			web.arena.CurrentMatch.DisplayName = name
			web.arena.MatchLoadNotifier.Notify()
			continue
		case "updateRealtimeScore":
			args := data.(map[string]interface{})
			web.arena.BlueScore.AutoPoints = int(args["blueAuto"].(float64))
			web.arena.RedScore.AutoPoints = int(args["redAuto"].(float64))
			web.arena.BlueScore.TeleopPoints = int(args["blueTeleop"].(float64))
			web.arena.RedScore.TeleopPoints = int(args["redTeleop"].(float64))
			web.arena.BlueScore.EndgamePoints = int(args["blueEndgame"].(float64))
			web.arena.RedScore.EndgamePoints = int(args["redEndgame"].(float64))
			web.arena.RealtimeScoreNotifier.Notify()
		default:
			ws.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}

		// Send out the status again after handling the command, as it most likely changed as a result.
		err = ws.WriteNotifier(web.arena.ArenaStatusNotifier)
		if err != nil {
			log.Println(err)
			return
		}
	}
}
