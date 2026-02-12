package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type DeployConfig struct {
	ProjectID   string            `yaml:"project_id"`
	Region      string            `yaml:"region"`
	ServiceName string            `yaml:"service_name"`
	ImageDigest string            `yaml:"image_digest"` // must end with @sha256:...
	CPU         string            `yaml:"cpu,omitempty"`
	Memory      string            `yaml:"memory,omitempty"`
	Concurrency int               `yaml:"concurrency,omitempty"`
	TimeoutSec  int               `yaml:"timeout_seconds,omitempty"`
	MaxInst     int               `yaml:"max_instances,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	Eventarc    EventarcConfig    `yaml:"eventarc"`
	IAM         IAMConfig         `yaml:"iam,omitempty"`
}

type EventarcConfig struct {
	TriggerName     string   `yaml:"trigger_name"`
	Bucket          string   `yaml:"bucket"`
	InputPrefix     string   `yaml:"input_prefix"`
	CeTypeAllowlist []string `yaml:"ce_type_allowlist"`
}

type IAMConfig struct {
	ServiceAccount string   `yaml:"service_account,omitempty"`
	Roles          []string `yaml:"roles,omitempty"`
}

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func LoadConfig(path string) (DeployConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return DeployConfig{}, err
	}
	var cfg DeployConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return DeployConfig{}, err
	}
	return cfg, cfg.Validate()
}

func (c DeployConfig) Validate() error {
	if strings.TrimSpace(c.ProjectID) == "" || strings.TrimSpace(c.Region) == "" || strings.TrimSpace(c.ServiceName) == "" {
		return fmt.Errorf("project_id, region, service_name are required")
	}
	nameRe := regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	if !nameRe.MatchString(c.ServiceName) {
		return fmt.Errorf("service_name must match %s", nameRe.String())
	}
	digestRe := regexp.MustCompile(`@sha256:[0-9a-f]{64}$`)
	if !digestRe.MatchString(c.ImageDigest) {
		return fmt.Errorf("image_digest must end with @sha256:<64 hex>")
	}
	if strings.TrimSpace(c.Eventarc.TriggerName) == "" || strings.TrimSpace(c.Eventarc.Bucket) == "" ||
		strings.TrimSpace(c.Eventarc.InputPrefix) == "" || len(c.Eventarc.CeTypeAllowlist) == 0 {
		return fmt.Errorf("eventarc fields required (trigger_name, bucket, input_prefix, ce_type_allowlist)")
	}
	for k := range c.Env {
		if forbiddenKey(k) {
			return fmt.Errorf("forbidden env key: %s (no secrets rule)", k)
		}
	}
	return nil
}

func forbiddenKey(k string) bool {
	u := strings.ToUpper(k)
	for _, sub := range []string{"SECRET", "TOKEN", "PASSWORD", "KEY"} {
		if strings.Contains(u, sub) {
			return true
		}
	}
	return false
}

func (c DeployConfig) EnvPairs() []KV {
	p := make([]KV, 0, len(c.Env))
	for k, v := range c.Env {
		p = append(p, KV{Key: k, Value: v})
	}
	sort.Slice(p, func(i, j int) bool { return p[i].Key < p[j].Key })
	return p
}
