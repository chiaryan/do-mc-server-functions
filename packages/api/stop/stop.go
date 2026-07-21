package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func CreateErrorResponse(err string) map[string]interface{} {
	return CreateResponseBody(map[string]interface{}{"error": err})
}

func CreateResponseBody(body map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"body": body,
	}
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
			// hcl value requires double quote

			hclstr := string(hclwrite.TokensForValue(cty.StringVal(value)).Bytes())
			vars = append(vars, &tfe.RunVariable{Key: mapping.To, Value: hclstr})
		}
	}
	return vars
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
	if current_run.IsDestroy && current_run.Status == "applied" {
		return CreateErrorResponse("server is paused")
	}

	run, err := client.Runs.Create(ctx, tfe.RunCreateOptions{
		Workspace:       &tfe.Workspace{ID: workspace_id},
		AllowEmptyApply: tfe.Bool(false),
		AutoApply:       tfe.Bool(true),
		IsDestroy:       tfe.Bool(true),
		Variables:       lookupTfEnvs(),
	})

	if err != nil {
		return CreateErrorResponse(err.Error())
	}

	return CreateResponseBody(map[string]interface{}{
		"run": run.ID,
	})
}
