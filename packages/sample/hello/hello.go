package main

import "os"

func Main(args map[string]interface{}) map[string]interface{} {
	os.LookupEnv("TFE_TOKEN")

	return map[string]interface{}{"error": "no env"}
	// if !success {
	// 	return map[string]interface{}{"error": "no env"}
	// }

	// client, err := tfe.NewClient(&tfe.Config{
	// 	// BasePath: "/api/v2",
	// 	Token: token,
	// })

	// if err != nil {
	// 	return map[string]interface{}{"error": err.Error()}
	// }

	// return map[string]interface{}{
	// 	"token":   token[:5],
	// 	"version": client.RemoteAPIVersion(),
	// }
}
