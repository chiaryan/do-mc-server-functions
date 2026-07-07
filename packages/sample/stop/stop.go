package main

import (
	"context"
	"os"

	"github.com/hashicorp/go-tfe"
)

func Main(ctx context.Context, args map[string]interface{}) map[string]interface{} {
	tfe_token, success := os.LookupEnv("TFE_TOKEN")
	if !success {
		return map[string]interface{}{
			"body": map[string]interface{}{
				"error": "no tfe token",
			},
		}
	}
	workspace_id, success := os.LookupEnv("WORKSPACE_ID")

	if !success {
		return map[string]interface{}{
			"body": map[string]interface{}{
				"error": "no workspace id",
			},
		}
	}

	client, err := tfe.NewClient(&tfe.Config{
		Token: tfe_token,
	})

	if err != nil {
		return map[string]interface{}{
			"body": map[string]interface{}{
				"error": err.Error(),
			},
		}
	}

	run, err := client.Runs.Create(ctx, tfe.RunCreateOptions{
		Workspace:       &tfe.Workspace{ID: workspace_id},
		AllowEmptyApply: tfe.Bool(false),
		AutoApply:       tfe.Bool(true),
		IsDestroy:       tfe.Bool(true),
	})

	if err != nil {
		return map[string]interface{}{
			"body": map[string]interface{}{
				"error destroying Run": err.Error(),
			},
		}
	}

	return map[string]interface{}{
		"body": map[string]interface{}{
			"run": run.ID,
		},
	}
}
