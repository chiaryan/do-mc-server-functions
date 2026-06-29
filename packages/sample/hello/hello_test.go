package main

import (
	"maps"
	"testing"

	dummygo "github.com/jlemesh/dummy-go/v2"
)

func TestHelloName(t *testing.T) {
	if !maps.Equal(Main(map[string]interface{}{"name": "nam"}), map[string]interface{}{"body": "Hello nam!"}) {
		t.Errorf("failed")
	}
}

func TestAdd(t *testing.T) {
	if dummygo.Add(1, 2, 3) != 6 {
		t.Errorf("failed")
	}
}
