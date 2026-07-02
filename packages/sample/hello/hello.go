package main

import (
	"os"
)

func Main(args map[string]interface{}) map[string]string {
	os.LookupEnv("TFE_TOKEN")

	return map[string]string{"error": "no env"}

	// if !success {
	// 	return map[string]string{"error": "no env"}
	// }

	// client, err := tfe.NewClient(&tfe.Config{
	// 	// BasePath: "/api/v2",
	// 	Token: token,
	// })

	// if err != nil {
	// 	return map[string]string{"error": err.Error()}
	// }

	// return map[string]string{
	// 	"token":   token[:5],
	// 	"version": client.RemoteAPIVersion(),
	// }

	// name, ok := args["name"].(string)
	// if !ok {
	// 	name = "stranger"
	// }
	// dummygo.Add(1, 2, 3)
	// msg := make(map[string]interface{})
	// msg["body"] = "Hello " + name + "!"
	// return msg
}
