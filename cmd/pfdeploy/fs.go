package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ResetOutDir(outDir string) error {
	if strings.TrimSpace(outDir) == "" {
		return fmt.Errorf("refusing to clear empty out dir")
	}

	clean := filepath.Clean(outDir)

	// Never allow current/parent.
	switch clean {
	case ".", "..":
		return fmt.Errorf("refusing to clear unsafe out dir %q", outDir)
	}

	// Reject filesystem roots.
	if clean == string(os.PathSeparator) || clean == "/" || clean == "\\" {
		return fmt.Errorf("refusing to clear unsafe out dir %q", outDir)
	}

	// Windows volume root: "C:" or "C:\"
	// UNC root also shows up via VolumeName; treat volume roots as unsafe.
	vol := filepath.VolumeName(clean)
	if vol != "" {
		if clean == vol || clean == vol+string(os.PathSeparator) {
			return fmt.Errorf("refusing to clear unsafe out dir %q", outDir)
		}
	}

	if err := os.RemoveAll(clean); err != nil {
		return err
	}
	return os.MkdirAll(clean, 0o755)
}

func WriteText(path, s string) error { return atomicWrite(path, []byte(s)) }

func atomicWrite(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func WriteArtifacts(outDir string, arts Artifacts) error {
	d, err := canonicalJSON(arts.Deploy)
	if err != nil {
		return err
	}
	t, err := canonicalJSON(arts.Trigger)
	if err != nil {
		return err
	}
	i, err := canonicalJSON(arts.IAM)
	if err != nil {
		return err
	}

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

func FileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

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
		return fmt.Errorf("file count mismatch: expected %d, got %d", len(exp), len(out))
	}

	for i := range exp {
		if exp[i] != out[i] {
			return fmt.Errorf("path mismatch: expected %s, got %s", exp[i], out[i])
		}
		eb, err := os.ReadFile(filepath.Join(expDir, exp[i]))
		if err != nil {
			return err
		}
		ob, err := os.ReadFile(filepath.Join(outDir, out[i]))
		if err != nil {
			return err
		}
		if !bytes.Equal(eb, ob) {
			return fmt.Errorf("content mismatch: %s", exp[i])
		}
	}

	return nil
}

func listFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}
