package main

import (
	"context"
	"maps"
	"os"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	dummygo "github.com/jlemesh/dummy-go/v2"
)

func TestMain(t *testing.T) {

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
