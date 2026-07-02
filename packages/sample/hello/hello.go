package main

import (
	"os"

	tfe "github.com/hashicorp/go-tfe"
)

func Main(args map[string]interface{}) map[string]interface{} {
	token, success := os.LookupEnv("TFE_TOKEN")

	if !success {
		return map[string]interface{}{"error": "no env"}
	}

	client, err := tfe.NewClient(&tfe.Config{
		// BasePath: "/api/v2",
		Token: token,
	})

	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	return map[string]interface{}{
		"token":   token[:5],
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
