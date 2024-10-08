// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the match play page.

var websocket;
var currentMatchId;
var lowBatteryThreshold = 8;
function capitalizeFirstLetter(string) {
    return string.charAt(0).toUpperCase() + string.slice(1).toLowerCase();
}


var scores = {
    red: {
        score: 0,
        auto: 0,
        teleop: 0,
        endgame: 0
    },
    blue: {
        score: 0,
        auto: 0,
        teleop: 0,
        endgame: 0
    }
};

var match = {
    state: "PRE_MATCH",
}

// Sends a websocket message to load a team into an alliance station.
var substituteTeam = function (team, position) {
    websocket.send("substituteTeam", { team: parseInt(team), position: position })
};

// Sends a websocket message to toggle the bypass status for an alliance station.
var toggleBypass = function (station) {
    websocket.send("toggleBypass", station);
};

// Sends a websocket message to start the match.
var startMatch = function () {
    websocket.send("startMatch",
        { muteMatchSounds: $("#muteMatchSounds").prop("checked") });
};

// Sends a websocket message to abort the match.
var abortMatch = function () {
    websocket.send("abortMatch");
};

// Sends a websocket message to signal to the volunteers that they may enter the field.
var signalVolunteers = function () {
    websocket.send("signalVolunteers");
};

// Sends a websocket message to signal to the teams that they may enter the field.
var signalReset = function () {
    websocket.send("signalReset");
};

// Sends a websocket message to commit the match score and load the next match.
var commitResults = function () {
    websocket.send("commitResults");
};

// Sends a websocket message to discard the match score and load the next match.
var discardResults = function () {
    websocket.send("discardResults");
};

// Sends a websocket message to change what the audience display is showing.
var setAudienceDisplay = function () {
    websocket.send("setAudienceDisplay", $("input[name=audienceDisplay]:checked").val());
};

// Sends a websocket message to change what the alliance station display is showing.
var setAllianceStationDisplay = function () {
    websocket.send("setAllianceStationDisplay", $("input[name=allianceStationDisplay]:checked").val());
};

// Sends a websocket message to start the timeout.
var startTimeout = function () {
    var duration = $("#timeoutDuration").val().split(":");
    var durationSec = parseFloat(duration[0]);
    if (duration.length > 1) {
        durationSec = durationSec * 60 + parseFloat(duration[1]);
    }
    websocket.send("startTimeout", durationSec);
};

// Sends a websocket message to update the realtime score
var updateRealtimeScore = function () {
    websocket.send("updateRealtimeScore", {
        blueAuto: parseInt($("#blueAutoScore").val()),
        redAuto: parseInt($("#redAutoScore").val()),
        blueTeleop: parseInt($("#blueTeleopScore").val()),
        redTeleop: parseInt($("#redTeleopScore").val()),
        blueEndgame: parseInt($("#blueEndgameScore").val()),
        redEndgame: parseInt($("#redEndgameScore").val())
    })
};

var increaseScore = function (color, amount) {
    var scoreObject = {
        blueAuto: scores.blue.auto,
        redAuto: scores.red.auto,
        blueTeleop: scores.blue.teleop,
        redTeleop: scores.red.teleop,
        blueEndgame: scores.blue.endgame,
        redEndgame: scores.red.endgame
    }
    if(color == "red") {
        if(match.state == "AUTO_PERIOD") {
            scoreObject.redAuto += amount;
        }
        else if(match.state == "TELEOP_PERIOD") {
            scoreObject.redTeleop += amount;
        }
        else if(match.state == "POST_MATCH") {
            scoreObject.redEndgame += amount;
        }
    }
    else if(color == "blue") {
        if(match.state == "AUTO_PERIOD") {
            scoreObject.blueAuto += amount;
        }
        else if(match.state == "TELEOP_PERIOD") {
            scoreObject.blueTeleop += amount;
        }
        else if(match.state == "POST_MATCH") {
            scoreObject.blueEndgame += amount;
        }
    }
    websocket.send("updateRealtimeScore", scoreObject);

}

var scoreKeyHandler = function (e) {
    var keycode = (event.keyCode ? event.keyCode : event.which);
    if (keycode == 13) {
        $('#' + $(event.target).attr("data-next")).focus().select();
    }
};

var confirmCommit = function (isReplay) {
    if (isReplay) {
        // Show the appropriate message(s) in the confirmation dialog.
        $("#confirmCommitReplay").css("display", isReplay ? "block" : "none");
        $("#confirmCommitResults").modal("show");
    } else {
        commitResults();
    }
};

// Sends a websocket message to specify a custom name for the current test match.
var setTestMatchName = function () {
    websocket.send("setTestMatchName", $("#testMatchName").val());
};

// Handles a websocket message to update the team connection status.
var handleArenaStatus = function (data) {
    // If getting data for the wrong match (e.g. after a server restart), reload the page.
    if (currentMatchId == null) {
        currentMatchId = data.MatchId;
    } else if (currentMatchId !== data.MatchId) {
        location.reload();
    }

    // Update the team status view.
    $.each(data.AllianceStations, function (station, stationStatus) {
        var wifiStatus = data.TeamWifiStatuses[station];
        $("#status" + station + " .radio-status").text(wifiStatus.TeamId);

        if (stationStatus.DsConn) {
            // Format the driver station status box.
            var dsConn = stationStatus.DsConn;
            $("#status" + station + " .ds-status").attr("data-status-ok", dsConn.DsLinked);

            // Format the radio status box according to the connection status of the robot radio.
            var radioOkay = stationStatus.Team && stationStatus.Team.Id === wifiStatus.TeamId && wifiStatus.RadioLinked;
            $("#status" + station + " .radio-status").attr("data-status-ok", radioOkay);

            // Format the robot status box.
            var robotOkay = dsConn.BatteryVoltage > lowBatteryThreshold && dsConn.RobotLinked;
            $("#status" + station + " .robot-status").attr("data-status-ok", robotOkay);
            if (stationStatus.DsConn.SecondsSinceLastRobotLink > 1 && stationStatus.DsConn.SecondsSinceLastRobotLink < 1000) {
                $("#status" + station + " .robot-status").text(stationStatus.DsConn.SecondsSinceLastRobotLink.toFixed());
            } else {
                $("#status" + station + " .robot-status").text(dsConn.BatteryVoltage.toFixed(1) + "V");
            }
        } else {
            $("#status" + station + " .ds-status").attr("data-status-ok", "");
            $("#status" + station + " .robot-status").attr("data-status-ok", "");
            $("#status" + station + " .robot-status").text("");

            // Format the robot status box according to whether the AP is configured with the correct SSID.
            var expectedTeamId = stationStatus.Team ? stationStatus.Team.Id : 0;
            if (wifiStatus.TeamId === expectedTeamId) {
                if (wifiStatus.RadioLinked) {
                    $("#status" + station + " .radio-status").attr("data-status-ok", true);
                } else {
                    $("#status" + station + " .radio-status").attr("data-status-ok", "");
                }
            } else {
                $("#status" + station + " .radio-status").attr("data-status-ok", false);
            }
        }

        if (stationStatus.Estop) {
            $("#status" + station + " .bypass-status").attr("data-status-ok", false);
            $("#status" + station + " .bypass-status").text("ES");
        } else if (stationStatus.Bypass) {
            $("#status" + station + " .bypass-status").attr("data-status-ok", false);
            $("#status" + station + " .bypass-status").text("B");
        } else {
            $("#status" + station + " .bypass-status").attr("data-status-ok", true);
            $("#status" + station + " .bypass-status").text("");
        }
    });

    // Enable/disable the buttons based on the current match state.
    switch (matchStates[data.MatchState]) {
        case "PRE_MATCH":
            $('.score-button').prop("disabled", true);
            $('.score-button').addClass('disabled-score-button');
            $("#startMatch").prop("disabled", !data.CanStartMatch);
            $("#abortMatch").prop("disabled", true);
            $("#signalVolunteers").prop("disabled", true);
            $("#signalReset").prop("disabled", true);
            $("#commitResults").prop("disabled", true);
            $("#discardResults").prop("disabled", true);
            $("#editResults").prop("disabled", true);
            $("#startTimeout").prop("disabled", false);
            $("#blueAutoScore").val("0");
            $("#redAutoScore").val("0");
            $("#blueTeleopScore").val("0");
            $("#redTeleopScore").val("0");
            $("#blueEndgameScore").val("0");
            $("#redEndgameScore").val("0");
            $("#blueAutoScore").prop("disabled", true);
            $("#redAutoScore").prop("disabled", true);
            $("#blueTeleopScore").prop("disabled", true);
            $("#redTeleopScore").prop("disabled", true);
            $("#blueEndgameScore").prop("disabled", true);
            $("#redEndgameScore").prop("disabled", true);
            break;
        case "START_MATCH":
        case "WARMUP_PERIOD":
        case "AUTO_PERIOD":
            $('.score-button').prop("disabled", true);
            $('.score-button').removeClass('disabled-score-button');
        case "PAUSE_PERIOD":
        case "TELEOP_PERIOD":
            $('.score-button').prop("disabled", false);
            $('.score-button').removeClass('disabled-score-button');
            $('#blueIncrease').prep("disabled", false);
            $("#startMatch").prop("disabled", true);
            $("#abortMatch").prop("disabled", false);
            $("#signalVolunteers").prop("disabled", true);
            $("#signalReset").prop("disabled", true);
            $("#commitResults").prop("disabled", true);
            $("#discardResults").prop("disabled", true);
            $("#editResults").prop("disabled", true);
            $("#startTimeout").prop("disabled", true);
            $("#blueAutoScore").prop("disabled", false);
            $("#redAutoScore").prop("disabled", false);
            $("#blueTeleopScore").prop("disabled", false);
            $("#redTeleopScore").prop("disabled", false);
            $("#blueEndgameScore").prop("disabled", false);
            $("#redEndgameScore").prop("disabled", false);
            break;
        case "POST_MATCH":
            $('.score-button').prop("disabled", false);
            $('.score-button').removeClass('disabled-score-button');
            $("#startMatch").prop("disabled", true);
            $("#abortMatch").prop("disabled", true);
            $("#signalVolunteers").prop("disabled", false);
            $("#signalReset").prop("disabled", false);
            $("#commitResults").prop("disabled", false);
            $("#discardResults").prop("disabled", false);
            $("#editResults").prop("disabled", false);
            $("#startTimeout").prop("disabled", true);
            $("#blueAutoScore").prop("disabled", false);
            $("#redAutoScore").prop("disabled", false);
            $("#blueTeleopScore").prop("disabled", false);
            $("#redTeleopScore").prop("disabled", false);
            $("#blueEndgameScore").prop("disabled", false);
            $("#redEndgameScore").prop("disabled", false);
            break;
        case "TIMEOUT_ACTIVE":
            $("#startMatch").prop("disabled", true);
            $("#abortMatch").prop("disabled", false);
            $("#signalVolunteers").prop("disabled", true);
            $("#signalReset").prop("disabled", true);
            $("#commitResults").prop("disabled", true);
            $("#discardResults").prop("disabled", true);
            $("#editResults").prop("disabled", true);
            $("#startTimeout").prop("disabled", true);
            $("#blueAutoScore").prop("disabled", false);
            $("#redAutoScore").prop("disabled", false);
            $("#blueTeleopScore").prop("disabled", false);
            $("#redTeleopScore").prop("disabled", false);
            $("#blueEndgameScore").prop("disabled", false);
            $("#redEndgameScore").prop("disabled", false);
            break;
        case "POST_TIMEOUT":
            $("#startMatch").prop("disabled", true);
            $("#abortMatch").prop("disabled", true);
            $("#signalVolunteers").prop("disabled", true);
            $("#signalReset").prop("disabled", true);
            $("#commitResults").prop("disabled", true);
            $("#discardResults").prop("disabled", true);
            $("#editResults").prop("disabled", true);
            $("#startTimeout").prop("disabled", true);
            $("#blueAutoScore").prop("disabled", false);
            $("#redAutoScore").prop("disabled", false);
            $("#blueTeleopScore").prop("disabled", false);
            $("#redTeleopScore").prop("disabled", false);
            $("#blueEndgameScore").prop("disabled", false);
            $("#redEndgameScore").prop("disabled", false);
            break;
    }

    if (data.PlcIsHealthy) {
        $("#plcStatus").text("Connected");
        $("#plcStatus").attr("data-ready", true);
    } else {
        $("#plcStatus").text("Not Connected");
        $("#plcStatus").attr("data-ready", false);
    }
    $("#fieldEstop").attr("data-ready", !data.FieldEstop);
    $.each(data.PlcArmorBlockStatuses, function (name, status) {
        $("#plc" + name + "Status").attr("data-ready", status);
    });
};

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function (data) {
    translateMatchTime(data, function (matchState, matchStateText, countdownSec) {
        $("#currentStage").text(capitalizeFirstLetter(matchStateText));
        $("#matchTime").text(countdownSec);
        match.state = matchState;
    });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function (data) {
    scores.red.score = data.Red.ScoreSummary.Score;
    scores.blue.score = data.Blue.ScoreSummary.Score;
    scores.red.score = data.Red.ScoreSummary.Score;
    scores.blue.score = data.Blue.ScoreSummary.Score;
    scores.red.auto = data.Red.ScoreSummary.AutoPoints;
    scores.blue.auto = data.Blue.ScoreSummary.AutoPoints;
    scores.red.teleop = data.Red.ScoreSummary.TeleopPoints;
    scores.blue.teleop = data.Blue.ScoreSummary.TeleopPoints;
    scores.red.endgame = data.Red.ScoreSummary.EndgamePoints;
    scores.blue.endgame = data.Blue.ScoreSummary.EndgamePoints;

    if (parseInt($("#blueTotalScore").val()) != data.Blue.ScoreSummary.Score) {
        $("#blueTotalScore").text(data.Blue.ScoreSummary.Score);
    }
    if (parseInt($("#redTotalScore").val()) != data.Red.ScoreSummary.Score) {
        $("#redTotalScore").text(data.Red.ScoreSummary.Score);
    }
    
}

// Handles a websocket message to update the audience display screen selector.
var handleAudienceDisplayMode = function (data) {
    $("input[name=audienceDisplay]:checked").prop("checked", false);
    $("input[name=audienceDisplay][value=" + data + "]").prop("checked", true);
};

// Handles a websocket message to update the alliance station display screen selector.
var handleAllianceStationDisplayMode = function (data) {
    $("input[name=allianceStationDisplay]:checked").prop("checked", false);
    $("input[name=allianceStationDisplay][value=" + data + "]").prop("checked", true);
};

// Handles a websocket message to update the event status message.
var handleEventStatus = function (data) {
    if (data.CycleTime === "") {
        $("#cycleTimeMessage").text("Last cycle time: Unknown");
    } else {
        $("#cycleTimeMessage").text("Last cycle time: " + data.CycleTime);
    }
    $("#earlyLateMessage").text(data.EarlyLateMessage);
};

$(function () {
    // Activate tooltips above the status headers.
    $("[data-toggle=tooltip]").tooltip({ "placement": "top" });

    // Set up the websocket back to the server.
    websocket = new CheesyWebsocket("/match_play/websocket", {
        allianceStationDisplayMode: function (event) { handleAllianceStationDisplayMode(event.data); },
        arenaStatus: function (event) { handleArenaStatus(event.data); },
        audienceDisplayMode: function (event) { handleAudienceDisplayMode(event.data); },
        eventStatus: function (event) { handleEventStatus(event.data); },
        matchTime: function (event) { handleMatchTime(event.data); },
        matchTiming: function (event) { handleMatchTiming(event.data); },
        realtimeScore: function (event) { handleRealtimeScore(event.data); },
    });
});
