package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func ResetOutDir(outDir string) error {
	clean := filepath.Clean(outDir)
	switch clean {
	case ".", "..", string(os.PathSeparator):
		return fmt.Errorf("refusing to clear unsafe out dir %q", outDir)
	}
	_ = os.RemoveAll(clean)
	return os.MkdirAll(clean, 0o755)
}

func WriteText(path, s string) error { return atomicWrite(path, []byte(s)) }

func atomicWrite(path string, b []byte) error {
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func WriteArtifacts(outDir string, arts Artifacts) error {
	d, _ := canonicalJSON(arts.Deploy)
	t, _ := canonicalJSON(arts.Trigger)
	i, _ := canonicalJSON(arts.IAM)

	files := map[string][]byte{
		"deploy_manifest.json":  append(d, '\n'),
		"trigger_manifest.json": append(t, '\n'),
		"iam_manifest.json":     append(i, '\n'),
	}
	for name, b := range files {
		if err := atomicWrite(filepath.Join(outDir, name), b); err != nil {
			return err
		}
	}
	return atomicWrite(filepath.Join(outDir, "manifest.sha256"), []byte(manifestSha256(files)))
}

func ListCases(root string) ([]string, error) {
	ents, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range ents {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func FileExists(p string) bool { _, err := os.Stat(p); return err == nil }

func DiffTrees(expDir, outDir string) error {
	exp, err := listFiles(expDir)
	if err != nil {
		return err
	}
	out, err := listFiles(outDir)
	if err != nil {
		return err
	}
	if len(exp) != len(out) {
		return fmt.Errorf("file count mismatch")
	}
	for i := range exp {
		if exp[i] != out[i] {
			return fmt.Errorf("path mismatch")
		}
		eb, _ := os.ReadFile(filepath.Join(expDir, exp[i]))
		ob, _ := os.ReadFile(filepath.Join(outDir, out[i]))
		if !bytes.Equal(eb, ob) {
			return fmt.Errorf("content mismatch: %s", exp[i])
		}
	}
	return nil
}

func listFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}
