package main

import (
	"os"

	"github.com/hashicorp/go-tfe"
)

func Main(args map[string]interface{}) map[string]interface{} {
	token, success := os.LookupEnv("TFE_TOKEN")

	if !success {
		return map[string]interface{}{
			"body": map[string]interface{}{
				"error": "no env",
			},
		}
	}

	tfe.NewClient(&tfe.Config{
		// BasePath: "/api/v2",
		Token: token,
	})

	return map[string]interface{}{
		"body": map[string]interface{}{
			"token": token[:5],
		},
	}

	// if err != nil {
	// 	return map[string]interface{}{"error": err.Error()}
	// }

	// return map[string]interface{}{
	// 	"token":   token[:5],
	// 	"version": client.RemoteAPIVersion(),
	// }
}
