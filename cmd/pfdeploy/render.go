package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type DeployManifest struct {
	ProjectID   string      `json:"project_id"`
	Region      string      `json:"region"`
	ServiceName string      `json:"service_name"`
	ImageDigest string      `json:"image_digest"`
	Runtime     RuntimeSpec `json:"runtime"`
	Env         []KV        `json:"env"`
}

type RuntimeSpec struct {
	CPU         string `json:"cpu,omitempty"`
	Memory      string `json:"memory,omitempty"`
	Concurrency int    `json:"concurrency,omitempty"`
	TimeoutSec  int    `json:"timeout_seconds,omitempty"`
	MaxInst     int    `json:"max_instances,omitempty"`
}

type TriggerManifest struct {
	TriggerName     string   `json:"trigger_name"`
	Bucket          string   `json:"bucket"`
	InputPrefix     string   `json:"input_prefix"`
	CeTypeAllowlist []string `json:"ce_type_allowlist"`
}

type IAMManifest struct {
	ServiceAccount string   `json:"service_account,omitempty"`
	Roles          []string `json:"roles"`
}

type Artifacts struct {
	Deploy  DeployManifest
	Trigger TriggerManifest
	IAM     IAMManifest
}

func Render(cfg DeployConfig) (Artifacts, error) {
	allow := append([]string(nil), cfg.Eventarc.CeTypeAllowlist...)
	sort.Strings(allow)

	roles := append([]string(nil), cfg.IAM.Roles...)
	sort.Strings(roles)
	if len(roles) == 0 {
		roles = []string{"roles/run.invoker"}
	}

	return Artifacts{
		Deploy: DeployManifest{
			ProjectID:   cfg.ProjectID,
			Region:      cfg.Region,
			ServiceName: cfg.ServiceName,
			ImageDigest: cfg.ImageDigest,
			Runtime: RuntimeSpec{
				CPU:         cfg.CPU,
				Memory:      cfg.Memory,
				Concurrency: cfg.Concurrency,
				TimeoutSec:  cfg.TimeoutSec,
				MaxInst:     cfg.MaxInst,
			},
			Env: cfg.EnvPairs(),
		},
		Trigger: TriggerManifest{
			TriggerName:     cfg.Eventarc.TriggerName,
			Bucket:          cfg.Eventarc.Bucket,
			InputPrefix:     cfg.Eventarc.InputPrefix,
			CeTypeAllowlist: allow,
		},
		IAM: IAMManifest{
			ServiceAccount: cfg.IAM.ServiceAccount,
			Roles:          roles,
		},
	}, nil
}

func canonicalJSON(v any) ([]byte, error) { return json.MarshalIndent(v, "", "  ") }

func shaLine(filename string, b []byte) string {
	h := sha256.Sum256(b)
	return fmt.Sprintf("%s  %s", hex.EncodeToString(h[:]), filename)
}

func manifestSha256(files map[string][]byte) string {
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	var lines []string
	for _, n := range names {
		lines = append(lines, shaLine(n, files[n]))
	}
	return strings.Join(lines, "\n") + "\n"
}
