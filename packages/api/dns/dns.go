package main

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
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

var client godo.Client

func env(key string) string {
	url, success := os.LookupEnv(key)
	if !success {
		panic("no env " + key)
	}
	return url
}
func verifyPassword(args map[string]interface{}) (map[string]interface{}, bool) {
	password, success := os.LookupEnv("PASSWORD_HASH")
	if !success {
		panic("no url")
	}

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

func Main(ctx context.Context, args map[string]interface{}) map[string]interface{} {
	client = *godo.NewFromToken(env("DO_TOKEN"))

	switch args["http"].(map[string]interface{})["method"] {

	case "POST":

		result, success := verifyPassword(args)
		if success {
			return result
		}

		return post(ctx)

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

func post(ctx context.Context) map[string]interface{} {
	id, err := strconv.ParseInt(env("RECORD_ID"), 10, 0)
	if err != nil {
		panic("record is not an int")
	}
	dpt, err := getDropletByName(ctx)
	if err != nil {
		return CreateErrorResponse(err.Error())
	}

	ip, err := dpt.PublicIPv4()
	if err != nil {
		return CreateErrorResponse(err.Error())
	}

	_, _, err = client.Domains.EditRecord(ctx, env("SERVER_DOMAIN"), int(id), &godo.DomainRecordEditRequest{Data: ip})
	if err != nil {
		return CreateErrorResponse(err.Error())
	}

	return CreateResponseBody(map[string]interface{}{"create": "ok"})
}
