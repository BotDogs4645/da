// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"testing"
	"time"

	"github.com/BotDogs4645/da/model"
	"github.com/BotDogs4645/da/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestSetupLowerThirds(t *testing.T) {
	web := setupTestWeb(t)

	web.arena.Database.CreateLowerThird(&model.LowerThird{TopText: "Top Text 1", BottomText: "Bottom Text 1"})
	web.arena.Database.CreateLowerThird(&model.LowerThird{TopText: "Top Text 2", BottomText: "Bottom Text 2", DisplayOrder: 1})
	web.arena.Database.CreateLowerThird(&model.LowerThird{TopText: "Top Text 3", BottomText: "Bottom Text 3", DisplayOrder: 2})

	recorder := web.getHttpResponse("/setup/lower_thirds")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Top Text 1")
	assert.Contains(t, recorder.Body.String(), "Bottom Text 2")

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/setup/lower_thirds/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	ws.Write("saveLowerThird", model.LowerThird{Id: 1, TopText: "Top Text 4", BottomText: "Bottom Text 1"})
	time.Sleep(time.Millisecond * 10) // Allow some time for the command to be processed.
	lowerThird, _ := web.arena.Database.GetLowerThirdById(1)
	assert.Equal(t, "Top Text 4", lowerThird.TopText)

	ws.Write("deleteLowerThird", model.LowerThird{Id: 1, TopText: "Top Text 4", BottomText: "Bottom Text 1"})
	time.Sleep(time.Millisecond * 10)
	lowerThird, _ = web.arena.Database.GetLowerThirdById(1)
	assert.Nil(t, lowerThird)

	assert.Equal(t, "blank", web.arena.AudienceDisplayMode)
	ws.Write("showLowerThird", model.LowerThird{Id: 2, TopText: "Top Text 5", BottomText: "Bottom Text 1"})
	time.Sleep(time.Millisecond * 10)
	lowerThird, _ = web.arena.Database.GetLowerThirdById(2)
	assert.Equal(t, "Top Text 5", lowerThird.TopText)
	assert.Equal(t, true, web.arena.ShowLowerThird)

	ws.Write("hideLowerThird", model.LowerThird{Id: 2, TopText: "Top Text 6", BottomText: "Bottom Text 1"})
	time.Sleep(time.Millisecond * 10)
	lowerThird, _ = web.arena.Database.GetLowerThirdById(2)
	assert.Equal(t, "Top Text 6", lowerThird.TopText)
	assert.Equal(t, false, web.arena.ShowLowerThird)

	ws.Write("reorderLowerThird", map[string]interface{}{"Id": 2, "moveUp": false})
	time.Sleep(time.Millisecond * 100)
	lowerThirds, _ := web.arena.Database.GetAllLowerThirds()
	assert.Equal(t, 3, lowerThirds[0].Id)
	assert.Equal(t, 2, lowerThirds[1].Id)
}
