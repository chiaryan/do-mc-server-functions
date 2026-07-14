package main

import (
	"context"
	"maps"
	"os"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	dummygo "github.com/jlemesh/dummy-go/v2"
	"github.com/joho/godotenv"
)

func TestMain(t *testing.T) {
	godotenv.Load()

	t.Logf("%v", Main(context.Background(), map[string]interface{}{}))
}

func TestHelloName(t *testing.T) {

	if !maps.Equal(Main(context.Background(), map[string]interface{}{"name": "nam"}), map[string]interface{}{"body": "Hello nam!"}) {
		t.Errorf("failed %v", Main(context.Background(), map[string]interface{}{"name": "nam"}))
	}
}

func TestAdd(t *testing.T) {
	if dummygo.Add(1, 2, 3) != 6 {
		t.Errorf("failed")
	}
}

func TestApi(t *testing.T) {
	client, err := tfe.NewClient(&tfe.Config{
		// BasePath: "/api/v2",
		Token: os.Getenv("TFE_TOKEN"),
	})

	if err != nil {
		t.Errorf("error")
	}

	t.Errorf("%s", client.RemoteAPIVersion())
}

func TestGet(t *testing.T) {
	godotenv.Load()

	client, err := tfe.NewClient(&tfe.Config{
		// BasePath: "/api/v2",
		Token: os.Getenv("TFE_TOKEN"),
	})

	if err != nil {
		t.Errorf("error")
	}

	wsp, err := client.Workspaces.ReadByID(context.Background(), os.Getenv("WORKSPACE_ID"))
	run, err := client.Runs.Read(context.Background(), wsp.CurrentRun.ID)

	t.Errorf("%v", run.CreatedAt)

}
