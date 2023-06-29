package config

import (
	"testing"
)

func TestPathfromFunctionName(t *testing.T) {
	p, m := getPathFromFunctionName("POST__project_fill")
	if m != "POST" {
		t.Error("Method not match")
	}
	if p != "/project/fill" {
		t.Errorf("Path not match, expected '/project/fill', got %s", p)
	}
}
