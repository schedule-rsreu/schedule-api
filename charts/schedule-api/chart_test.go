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

func TestMigrationRunsWithoutShellAndRetriesDatabaseStartup(t *testing.T) {
	command := exec.Command("helm", "template", "schedule-api", ".", "--is-upgrade", "--set-string", "image.tag=test")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("helm template failed: %v\n%s", err, output)
	}
	manifest := string(output)
	if strings.Contains(manifest, "/bin/sh") || strings.Contains(manifest, "nc -z") {
		t.Fatal("migration requires a shell that is absent from the runtime")
	}
	if !strings.Contains(manifest, "backoffLimit: 6") {
		t.Fatal("migration must retry temporary database startup failures")
	}
	if !strings.Contains(manifest, "/usr/local/bin/goose") || !strings.Contains(manifest, "$(POSTGRES_DSN)") {
		t.Fatal("migration must pass the database DSN directly to goose")
	}
}
