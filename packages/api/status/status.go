package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-tfe"
	"github.com/mcstatus-io/mcutil/v4/response"
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
		panic("no tfe token")
	}

	workspace_id, success := os.LookupEnv("WORKSPACE_ID")
	if !success {
		panic("no workspace id")
	}

	url, success := os.LookupEnv("SERVER_DOMAIN")
	if !success {
		panic("no url")
	}

	client, err := tfe.NewClient(&tfe.Config{
		Token: tfe_token,
	})

	if err != nil {
		return CreateErrorResponse(fmt.Sprintf("error creating client %s", err.Error()))
	}

	switch args["http"].(map[string]interface{})["method"] {
	case "GET":
		return get(ctx, client, workspace_id, url)

	case "POST":
		return post(ctx, client, workspace_id)

	default:
		return CreateErrorResponse("invalid http method")
	}
}

func post(ctx context.Context, client *tfe.Client, workspace_id string) map[string]interface{} {
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
		Variables:       lookupTfEnvs(),
	})

	if err != nil {
		return CreateErrorResponse(err.Error())
	}

	return CreateResponseBody(map[string]interface{}{
		"run": run.ID,
	})
}

func lookupTfEnvs() []*tfe.RunVariable {
	var vars []*tfe.RunVariable
	type S struct {
		From string
		To   string
	}

	var_name_mapping := []S{
		{From: "STOP_ADDRESS", To: "stop_function_address"},
		{From: "STOP_ADDRESS_TOKEN", To: "stop_function_token"},
		{From: "DO_TOKEN", To: "dotoken"},
		{From: "RECORD", To: "record"},
		{From: "DOMAIN", To: "domain"},
		{From: "ITZG_ENV", To: "itzg_env"},
		{From: "INSTANCE_SSH_KEY", To: "ssh_key"},
		{From: "INSTANCE_SIZE", To: "size"},
		{From: "INSTANCE_VOLUME_NAME", To: "volume_name"},
		{From: "INSTANCE_REGION", To: "region"},
		{From: "INSTANCE_AUTO_DESTROY", To: "auto_destroy"},
	}

	for _, mapping := range var_name_mapping {
		value, success := os.LookupEnv(mapping.From)
		if success || value != "" {
			vars = append(vars, &tfe.RunVariable{Key: mapping.To, Value: value})
		}
	}
	return vars
}

func get(ctx context.Context, client *tfe.Client, workspace_id string, url string) map[string]interface{} {
	type Run struct {
		run *tfe.Run
		err error
	}
	type Status struct {
		status *response.StatusModern
		err    error
	}

	tf_chan := make(chan Run)
	mc_chan := make(chan Status)

	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		tf_chan <- func() Run {
			wsp, err := client.Workspaces.ReadByID(ctx, workspace_id)
			if err != nil {
				return Run{err: err}
			}

			current_run, err := client.Runs.Read(ctx, wsp.CurrentRun.ID)
			if err != nil {
				return Run{err: err}
			}

			return Run{run: current_run}
		}()
	}()

	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		status, err := status.Modern(ctx, url, 25565)
		mc_chan <- Status{status, err}
	}()

	var tf Run
	var mc Status

	for range 2 {
		select {
		case tf = <-tf_chan:
			if tf.err != nil {
				return CreateErrorResponse(tf.err.Error())
			}
			if tf.run.IsDestroy {

				if tf.run.Status == "applied" {
					return CreateResponseBody(map[string]interface{}{
						"status": "paused",
						"at":     tf.run.CreatedAt,
					})
				}

				return CreateResponseBody(map[string]interface{}{
					"status": "pausing",
					"at":     tf.run.CreatedAt,
				})
			}

			if tf.run.Status != "applied" {
				return CreateResponseBody(map[string]interface{}{
					"status": "creating",
					"at":     tf.run.CreatedAt,
				})
			}

		case mc = <-mc_chan:
			if mc.err == nil {
				ret := map[string]interface{}{
					"status":      "running",
					"motd":        mc.status.MOTD.Raw,
					"players":     *mc.status.Players.Online,
					"max_players": *mc.status.Players.Max,
					"url":         url,
				}

				if mc.status.Favicon != nil {
					ret["icon"] = *mc.status.Favicon
				}

				return CreateResponseBody(ret)
			}

		}
	}

	return CreateResponseBody(map[string]interface{}{
		"status": "starting",
		"at":     tf.run.CreatedAt,
		"err":    mc.err.Error(),
	})
}
