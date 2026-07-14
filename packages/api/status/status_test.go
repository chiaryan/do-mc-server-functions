package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/mcstatus-io/mcutil/v4/status"
)

func TestGet(t *testing.T) {
	godotenv.Load()

	_, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	status, err := status.Modern(context.Background(), "mc.hypixel.net", 25565)

	if err != nil {
		t.Errorf("err %v", err)
	} else {
		_, err := json.Marshal(map[string]interface{}{
			"status":      "running",
			"motd":        status.MOTD.Raw,
			"players":     *status.Players.Online,
			"max_players": *status.Players.Max,
			"icon":        *status.Favicon,
		})
		if err != nil {
			t.Errorf("err %v", err.Error())
		} else {
			// t.Errorf("bytes %s", bytes)
			t.Errorf("bytes %s", *status.Favicon)
		}
	}

}
