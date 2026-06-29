package main

import (
	"maps"
	"testing"
)

func TestHelloName(t *testing.T) {
	if !maps.Equal(Main(map[string]interface{}{"name": "nam"}), map[string]interface{}{"body": "Hello nam!"}) {
		t.Errorf("failed")
	}
}
