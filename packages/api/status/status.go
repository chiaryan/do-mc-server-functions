package main

import (
	"context"
	"os"

	"github.com/hashicorp/go-tfe"
	"github.com/mcstatus-io/mcutil/v4/status"
)

func CreateErrorResponse(err string) map[string]interface{} {
	return CreateResponseBody(map[string]interface{}{"error": err})
}

func CreateResponseBody(body map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"body": body,
	}
}

func Main(ctx context.Context, args map[string]interface{}) map[string]interface{} {
	tfe_token, success := os.LookupEnv("TFE_TOKEN")
	if !success {
		return CreateErrorResponse("no tfe token")
	}

	workspace_id, success := os.LookupEnv("WORKSPACE_ID")
	if !success {
		return CreateErrorResponse("no workspace id")
	}

	url, success := os.LookupEnv("SERVER_DOMAIN")
	if !success {
		CreateErrorResponse("no url")
	}

	client, err := tfe.NewClient(&tfe.Config{
		Token: tfe_token,
	})

	if err != nil {
		return CreateErrorResponse(err.Error())
	}

	switch args["http"].(map[string]interface{})["method"] {
	case "GET":

		wsp, err := client.Workspaces.ReadByID(context.Background(), workspace_id)
		if err != nil {
			return CreateErrorResponse(err.Error())
		}

		current_run, err := client.Runs.Read(context.Background(), wsp.CurrentRun.ID)
		if err != nil {
			return CreateErrorResponse(err.Error())
		}

		if current_run.IsDestroy {

			if current_run.Status == "applied" {
				return CreateResponseBody(map[string]interface{}{
					"status": "paused",
					"at":     current_run.CreatedAt,
				})
			}

			return CreateResponseBody(map[string]interface{}{
				"status": "pausing",
				"at":     current_run.CreatedAt,
			})
		}

		if current_run.Status != "applied" {
			return CreateResponseBody(map[string]interface{}{
				"status": "creating",
				"at":     current_run.CreatedAt,
			})
		}

		status, err := status.Modern(context.Background(), url, 25565)

		if err != nil {
			return CreateResponseBody(map[string]interface{}{
				"status": "starting",
				"url":    url,
			})
		}

		ret := map[string]interface{}{
			"status":      "running",
			"motd":        status.MOTD.Raw,
			"players":     *status.Players.Online,
			"max_players": *status.Players.Max,
			"url":         url,
			"icon":        *status.Favicon,
		}

		if status.Favicon != nil {
			ret["icon"] = *status.Favicon
		}

		return CreateResponseBody(ret)

	case "POST":
		wsp, err := client.Workspaces.ReadByID(context.Background(), workspace_id)
		if err != nil {
			return CreateErrorResponse(err.Error())
		}

		current_run, err := client.Runs.Read(context.Background(), wsp.CurrentRun.ID)
		if err != nil {
			return CreateErrorResponse(err.Error())
		}

		if current_run.Status != "applied" || !current_run.IsDestroy {
			return CreateErrorResponse("server still up")
		}
		// if the last run was a completed destroy, create the run

		run, err := client.Runs.Create(ctx, tfe.RunCreateOptions{
			Workspace:       &tfe.Workspace{ID: workspace_id},
			AllowEmptyApply: tfe.Bool(false),
			AutoApply:       tfe.Bool(true),
		})

		if err != nil {
			return CreateErrorResponse(err.Error())
		}

		return CreateResponseBody(map[string]interface{}{
			"run": run.ID,
		})

	case "DELETE":

		wsp, err := client.Workspaces.ReadByID(context.Background(), workspace_id)
		if err != nil {
			return CreateErrorResponse(err.Error())
		}

		current_run, err := client.Runs.Read(context.Background(), wsp.CurrentRun.ID)
		if err != nil {
			return CreateErrorResponse(err.Error())
		}

		// if the last run was a non-destroy run, create the destroy run
		if current_run.IsDestroy {
			return CreateErrorResponse("server is paused")
		}

		run, err := client.Runs.Create(ctx, tfe.RunCreateOptions{
			Workspace:       &tfe.Workspace{ID: workspace_id},
			AllowEmptyApply: tfe.Bool(false),
			AutoApply:       tfe.Bool(true),
			IsDestroy:       tfe.Bool(true),
		})

		if err != nil {
			return CreateErrorResponse(err.Error())
		}

		return CreateResponseBody(map[string]interface{}{
			"run": run.ID,
		})
	default:
		return CreateErrorResponse("invalid http method")
	}
}
