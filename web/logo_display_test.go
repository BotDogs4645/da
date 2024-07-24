// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"testing"

	"github.com/BotDogs4645/da/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestLogoDisplay(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/displays/logo?displayId=1&message=")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Logo Display - Untitled Event - Cheesy Arena")
}

func TestLogoDisplayWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/displays/logo/websocket?displayId=123", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "displayConfiguration")
}
