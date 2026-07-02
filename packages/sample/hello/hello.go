package main

import (
	"os"

	tfe "github.com/hashicorp/go-tfe"
)

func Main(args map[string]interface{}) map[string]interface{} {
	client, err := tfe.NewClient(&tfe.Config{
		// BasePath: "/api/v2",
		Token: os.Getenv("TFE_TOKEN"),
	})

	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	return map[string]interface{}{
		"length":  len(os.Getenv("TFE_TOKEN")),
		"version": client.RemoteAPIVersion(),
	}

	// name, ok := args["name"].(string)
	// if !ok {
	// 	name = "stranger"
	// }
	// dummygo.Add(1, 2, 3)
	// msg := make(map[string]interface{})
	// msg["body"] = "Hello " + name + "!"
	// return msg
}
