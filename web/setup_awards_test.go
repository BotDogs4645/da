// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"testing"

	"github.com/BotDogs4645/da/model"
	"github.com/stretchr/testify/assert"
)

func TestSetupAwards(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.Database.CreateAward(&model.Award{Type: model.JudgedAward, AwardName: "Spirit Award"})
	web.arena.Database.CreateAward(&model.Award{Type: model.JudgedAward, AwardName: "Saftey Award"})

	recorder := web.getHttpResponse("/setup/awards")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Spirit Award")
	assert.Contains(t, recorder.Body.String(), "Saftey Award")

	recorder = web.postHttpResponse("/setup/awards", "action=delete&id=1")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/awards")
	assert.Equal(t, 200, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "Spirit Award")
	assert.Contains(t, recorder.Body.String(), "Saftey Award")

	recorder = web.postHttpResponse("/setup/awards", "awardId=2&awardName=Saftey+Award&personName=Englebert")
	assert.Equal(t, 303, recorder.Code)
	recorder = web.getHttpResponse("/setup/awards")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Englebert")
}

func TestSetupAwardsPublish(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.TbaClient.BaseUrl = "fakeurl"
	web.arena.EventSettings.TbaPublishingEnabled = true

	recorder := web.postHttpResponse("/setup/awards/publish", "")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Failed to publish awards")
}
