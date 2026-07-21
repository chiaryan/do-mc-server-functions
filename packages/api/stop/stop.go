package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-tfe"
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
		panic("no tfe token")
	}

	workspace_id, success := os.LookupEnv("WORKSPACE_ID")
	if !success {
		panic("no workspace id")
	}

	if args["http"].(map[string]interface{})["method"] != "DELETE" {
		return CreateErrorResponse("invalid http method")
	}

	client, err := tfe.NewClient(&tfe.Config{
		Token: tfe_token,
	})

	if err != nil {
		return CreateErrorResponse(fmt.Sprintf("error creating client %s", err.Error()))
	}

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
}
