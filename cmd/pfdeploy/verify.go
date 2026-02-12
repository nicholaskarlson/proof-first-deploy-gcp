package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type VerifyReport struct {
	Ok     bool          `json:"ok"`
	Checks []VerifyCheck `json:"checks"`
}

type VerifyCheck struct {
	Name string `json:"name"`
	Ok   bool   `json:"ok"`
}

type SnapshotService struct {
	ServiceName string `json:"service_name"`
	Region      string `json:"region"`
	ImageDigest string `json:"image_digest"`
}

func Verify(cfg DeployConfig, snapDir string) (VerifyReport, error) {
	b, err := os.ReadFile(filepath.Join(snapDir, "gcloud_service.json"))
	if err != nil {
		return VerifyReport{}, fmt.Errorf("missing snapshot file gcloud_service.json")
	}
	var s SnapshotService
	if err := json.Unmarshal(b, &s); err != nil {
		return VerifyReport{}, fmt.Errorf("bad snapshot json")
	}

	checks := []VerifyCheck{
		{Name: "service_name", Ok: s.ServiceName == cfg.ServiceName},
		{Name: "region", Ok: s.Region == cfg.Region},
		{Name: "image_digest", Ok: s.ImageDigest == cfg.ImageDigest},
	}
	if !(checks[0].Ok && checks[1].Ok && checks[2].Ok) {
		n := 0
		for _, c := range checks {
			if !c.Ok {
				n++
			}
		}
		return VerifyReport{}, fmt.Errorf("verify mismatch (%d field(s))", n)
	}
	return VerifyReport{Ok: true, Checks: checks}, nil
}
