package main

import (
	"os/exec"
	"strings"
)

func defaultProject() (string, error) {
	b, err := exec.Command("gcloud", "config", "get-value", "core/project", "-q").Output()
	return strings.TrimSpace(string(b)), err
}
