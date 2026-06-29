package main

import dummygo "github.com/jlemesh/dummy-go/v2"

func Main(args map[string]interface{}) map[string]interface{} {
	name, ok := args["name"].(string)
	if !ok {
		name = "stranger"
	}
	dummygo.Add(1, 2, 3)
	msg := make(map[string]interface{})
	msg["body"] = "Hello " + name + "!"
	return msg
}
