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

	client, err := tfe.NewClient(&tfe.Config{
		// BasePath: "/api/v2",
		Token: token,
	})

	if err != nil {
		return map[string]interface{}{
			"body": map[string]interface{}{
				"error": err.Error(),
			},
		}
	}
	return map[string]interface{}{
		"body": map[string]interface{}{
			"version": client.RemoteAPIVersion(),
		},
	}
}
