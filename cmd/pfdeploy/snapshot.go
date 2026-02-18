package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Snapshot(inDir string) (SnapshotService, error) {
	b, err := os.ReadFile(filepath.Join(inDir, "gcloud_service_raw.json"))
	if err != nil {
		return SnapshotService{}, fmt.Errorf("missing snapshot raw file gcloud_service_raw.json")
	}
	var s SnapshotService
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&s); err != nil {
		return SnapshotService{}, fmt.Errorf("bad snapshot raw json")
	}
	// Ensure there is no trailing junk.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return SnapshotService{}, fmt.Errorf("bad snapshot raw json")
	}
	if s.ServiceName == "" || s.Region == "" || s.ImageDigest == "" {
		return SnapshotService{}, fmt.Errorf("bad snapshot raw json")
	}
	return s, nil
}

func WriteSnapshot(outDir string, s SnapshotService) error {
	b, err := canonicalJSON(s)
	if err != nil {
		return err
	}
	files := map[string][]byte{
		"gcloud_service.json": append(b, '\n'),
	}
	if err := atomicWrite(filepath.Join(outDir, "gcloud_service.json"), files["gcloud_service.json"]); err != nil {
		return err
	}
	return atomicWrite(filepath.Join(outDir, "manifest.sha256"), []byte(manifestSha256(files)))
}
