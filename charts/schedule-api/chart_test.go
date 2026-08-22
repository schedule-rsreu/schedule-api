package chart_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestDWHURLUsesClusterService(t *testing.T) {
	command := exec.Command("helm", "template", "schedule-api", ".", "--set-string", "image.tag=test")
	command.Dir = "."
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("helm template failed: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "value: \"http://schedule-dwh.schedule-api.svc.cluster.local\"") {
		t.Fatal("DWH_URL does not use the cluster service")
	}
}
