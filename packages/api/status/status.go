package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/mcstatus-io/mcutil/v4/response"
	"github.com/mcstatus-io/mcutil/v4/status"
	"go.yaml.in/yaml/v4"
	"golang.org/x/crypto/bcrypt"
)

func CreateErrorResponse(err string) map[string]interface{} {
	return CreateResponseBody(map[string]interface{}{"error": err})
}

func CreateResponseBody(body map[string]interface{}) map[string]interface{} {
	ret := map[string]interface{}{
		"body": map[string]any{
			"value": fmt.Sprintf("%v", body),
		},
	}
	fmt.Printf("returning body %v", ret)
	return ret
}

var url, do_token, droplet_name, password, volume_id string
var client godo.Client

func env(key string) string {
	value, success := os.LookupEnv(key)
	if !success || value == "" {
		panic("no env " + key)
	}
	return value
}

func Main(ctx context.Context, args map[string]interface{}) map[string]interface{} {

	fmt.Printf("running with %v", args)

	url = env("SERVER_DOMAIN")
	do_token = env("DO_TOKEN")

	client = *godo.NewFromToken(do_token)

	switch args["http"].(map[string]interface{})["method"] {
	case "GET":
		fmt.Printf("calling get\n")
		return get(ctx)

	case "POST":
		fmt.Printf("calling post\n")
		return post(ctx)

	case "DELETE":

		fmt.Printf("calling delete\n")
		result, success := verifyPassword(args)
		if success {
			return result
		}

		return delete(ctx)
	default:
		return CreateErrorResponse("invalid http method")
	}
}

func verifyPassword(args map[string]interface{}) (map[string]interface{}, bool) {
	password := env("FUNCTIONS_PASSWORD_HASH")

	hash, ok := args["http"].(map[string]interface{})["headers"].(map[string]string)["authorization"]

	if !ok {
		return map[string]any{"statusCode": 401}, true
	}

	if !strings.HasPrefix(hash, "Bearer ") {
		return map[string]any{"statusCode": 400}, true
	}

	hash = hash[7:]

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		return map[string]any{"statusCode": 401}, true
	}
	return nil, false
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

	runcmd := []string{
		fmt.Sprintf(
			"curl -X POST \"%s/api/dns\" -H \"Content-Type: application/json\" -H \"Authorization: Bearer %s\"",
			env("FUNCTIONS_URL"),
			env("FUNCTIONS_PASSWORD"),
		),
		"docker run -v /mnt/data:/data -i -p 25565:25565 --env-file .env itzg/minecraft-server",
	}

	if strings.ToLower(env("AUTO_DESTROY")) == "true" {
		runcmd = append(runcmd,
			fmt.Sprintf("while true; do curl -X DELETE \"%s/api/status\" -H \"Content-Type: application/json\" -H \"Authorization: Bearer %s\"; sleep 300; done",
				env("FUNCTIONS_URL"),
				env("FUNCTIONS_PASSWORD")),
		)
	}

	document, err := yaml.Marshal(map[string]interface{}{
		"mounts": [][]string{{
			fmt.Sprintf("/dev/disk/by-id/scsi-0DO_Volume_%v", env("INSTANCE_VOLUME_NAME")),
			"/mnt/data", "ext4", "defaults,nofail,discard", "0", "0",
		}},

		"runcmd": runcmd,
		"write_files": []map[string]interface{}{{
			"content": env("ITZG_ENV"),
			"path":    "/.env",
		}},
	})

	_, _, err = client.Droplets.Create(ctx, &godo.DropletCreateRequest{
		Name:  droplet_name,
		Image: godo.DropletCreateImage{Slug: "ubuntu-24-04-x64"},
		Volumes: []godo.DropletCreateVolume{
			{ID: env("INSTANCE_VOLUME_ID")},
		},
		Size:       env("INSTANCE_SIZE"),
		UserData:   fmt.Sprintf("#cloud-config\n%v", string(document)),
		Monitoring: true,
		SSHKeys:    []godo.DropletCreateSSHKey{{Fingerprint: env("INSTANCE_SSH_KEY")}},
	})

	if err != nil {
		return CreateErrorResponse(err.Error())
	}

	return CreateResponseBody(map[string]interface{}{"delete": "ok"})
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
