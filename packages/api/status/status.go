package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mcstatus-io/mcutil/v4/response"
	"github.com/mcstatus-io/mcutil/v4/status"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/crypto/bcrypt"
)

func CreateErrorResponse(err string) map[string]interface{} {
	return CreateResponseBody(map[string]interface{}{"error": err})
}

func CreateResponseBody(body map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"body": body,
	}
}

var url, do_token, droplet_name, password, volume_id string
var client godo.Client

func env(key string) string {
	url, success := os.LookupEnv(key)
	if !success {
		panic("no env " + key)
	}
	return url
}

func Main(ctx context.Context, args map[string]interface{}) map[string]interface{} {
	var success bool

	url = env("SERVER_DOMAIN")
	do_token = env("DO_TOKEN")

	client = *godo.NewFromToken(do_token)

	switch args["http"].(map[string]interface{})["method"] {
	case "GET":
		return get(ctx)

	case "POST":
		return post(ctx)

	case "DELETE":

		password, success = os.LookupEnv("PASSWORD_HASH")
		if !success {
			panic("no url")
		}

		hash, ok := args["http"].(map[string]interface{})["headers"].(map[string]string)["authorization"]

		if !ok {
			return map[string]any{"statusCode": 401}
		}

		if !strings.HasPrefix(hash, "Bearer ") {
			return map[string]any{"statusCode": 400}
		}

		hash = hash[7:]

		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

		if err != nil {
			return map[string]any{"statusCode": 401}
		}

		return delete(ctx)
	default:
		return CreateErrorResponse("invalid http method")
	}
}

func getDropletByName(ctx context.Context) (*godo.Droplet, error) {
	dpts, _, err := client.Droplets.ListByName(ctx, env("INSTANCE_NAME"), &godo.ListOptions{})
	if err != nil {
		return nil, errors.New("failed to get droplets")
	}
	if len(dpts) == 0 {
		return nil, errors.New("droplet not found")
	}

	return &dpts[0], nil
}

func delete(ctx context.Context) map[string]interface{} {
	dpt, err := getDropletByName(ctx)
	if err != nil {
		return CreateErrorResponse(err.Error())
	}

	_, err = client.Droplets.Delete(ctx, dpt.ID)

	if err != nil {
		return CreateErrorResponse(err.Error())
	}
	return CreateResponseBody(map[string]any{"delete": "ok"})
}

func post(ctx context.Context) map[string]interface{} {

	// wsp, err := client.Workspaces.ReadByID(context.Background(), workspace_id)
	// if err != nil {
	// 	return CreateErrorResponse(err.Error())
	// }

	// _, err = client.Runs.Read(context.Background(), wsp.CurrentRun.ID)
	// if err != nil {
	// 	return CreateErrorResponse(err.Error())
	// }

	// if current_run.Status != "applied" || !current_run.IsDestroy {
	// 	return CreateErrorResponse("server still up")
	// }
	// if the last run was a completed destroy, create the run

	_, _, err := client.Droplets.Create(ctx, &godo.DropletCreateRequest{
		Name:  droplet_name,
		Image: godo.DropletCreateImage{Slug: "ubuntu-24-04-x64"},
		Volumes: []godo.DropletCreateVolume{
			{ID: env("INSTANCE_VOLUME_ID")},
		},
		Size: env("INSTANCE_SIZE"),
	})

	if err != nil {
		return CreateErrorResponse(err.Error())
	}

	return CreateResponseBody(map[string]interface{}{"create": "ok"})
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

func get(ctx context.Context) map[string]interface{} {
	type Run struct {
		droplet godo.Droplet
		actions []godo.Action
		err     error
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
			dpt, err := getDropletByName(ctx)
			if err != nil {
				return Run{err: err}
			}

			actions, _, err := client.Droplets.Actions(ctx, dpt.ID, &godo.ListOptions{})

			if err != nil {
				return Run{err: err}
			}

			return Run{actions: actions, droplet: *dpt}
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
				if tf.err.Error() == "droplet not found" {
					return CreateResponseBody(map[string]interface{}{
						"status": "paused",
					})
				}
				return CreateErrorResponse(tf.err.Error())
			}

			for _, atn := range tf.actions {
				if atn.Type == "destroy" {
					return CreateResponseBody(map[string]interface{}{
						"status": "pausing",
					})
				}
			}
			for _, atn := range tf.actions {
				if atn.Type == "destroy" {
					return CreateResponseBody(map[string]interface{}{
						"status": "pausing",
					})
				}
			}
			for _, atn := range tf.actions {
				if atn.Type == "create" {
					if atn.Status == "in-progress" {
						return CreateResponseBody(map[string]interface{}{
							"status": "creating",
						})
					}
					if atn.Status == "errored" {
						return CreateErrorResponse("droplet creation errored")
					}
				}
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
		"err":    mc.err.Error(),
	})
}
