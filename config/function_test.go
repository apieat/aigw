package config

import (
	"fmt"
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

func TestPathAndMethodInMap(t *testing.T) {
	var allowdMap = map[string]map[string]bool{
		"/project/fill": map[string]bool{"POST": true},
	}
	if !allowdMap["/project/fill"]["POST"] {
		t.Error("Path and method must be in map")
	}
	if allowdMap["/project/fill"]["GET"] {
		t.Error("Path and method must not be in map")
	}
	if allowdMap["/project/create"]["GET"] {
		t.Error("Path and method must not be in map")
	} else {
		fmt.Println(allowdMap["/project/create"]["GET1"])
	}
}
