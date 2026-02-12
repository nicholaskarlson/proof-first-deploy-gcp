package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func usage() {
	fmt.Fprintln(os.Stderr, "pfdeploy demo|render|verify ...")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "demo":
		os.Exit(cmdDemo(os.Args[2:]))
	case "render":
		os.Exit(cmdRender(os.Args[2:]))
	case "verify":
		os.Exit(cmdVerify(os.Args[2:]))
	default:
		usage()
		os.Exit(2)
	}
}

func cmdRender(args []string) int {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "config.yaml")
	outDir := fs.String("out", "", "out dir")
	_ = fs.Parse(args)
	if *cfgPath == "" || *outDir == "" {
		return 2
	}

	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		_ = ResetOutDir(*outDir)
		_ = WriteText(filepath.Join(*outDir, "error.txt"), err.Error()+"\n")
		return 0
	}

	_ = ResetOutDir(*outDir)

	arts, err := Render(cfg)
	if err != nil {
		_ = WriteText(filepath.Join(*outDir, "error.txt"), err.Error()+"\n")
		return 0
	}
	if err := WriteArtifacts(*outDir, arts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func cmdVerify(args []string) int {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "config.yaml")
	snapDir := fs.String("snapshot", "", "snapshot dir")
	outDir := fs.String("out", "", "out dir")
	_ = fs.Parse(args)
	if *cfgPath == "" || *snapDir == "" || *outDir == "" {
		return 2
	}

	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		_ = ResetOutDir(*outDir)
		_ = WriteText(filepath.Join(*outDir, "error.txt"), err.Error()+"\n")
		return 0
	}

	_ = ResetOutDir(*outDir)

	rep, err := Verify(cfg, *snapDir)
	if err != nil {
		_ = WriteText(filepath.Join(*outDir, "error.txt"), err.Error()+"\n")
		return 0
	}

	b, _ := json.MarshalIndent(rep, "", "  ")
	_ = WriteText(filepath.Join(*outDir, "verify_report.json"), string(b)+"\n")
	return 0
}

func cmdDemo(args []string) int {
	fs := flag.NewFlagSet("demo", flag.ContinueOnError)
	outBase := fs.String("out", "", "out dir")
	_ = fs.Parse(args)
	if *outBase == "" {
		return 2
	}
	_ = ResetOutDir(*outBase)

	cases, err := ListCases("fixtures/input")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	for _, c := range cases {
		inDir := filepath.Join("fixtures/input", c)
		expDir := filepath.Join("fixtures/expected", c)
		outDir := filepath.Join(*outBase, c)
		_ = ResetOutDir(outDir)

		cfg, err := LoadConfig(filepath.Join(inDir, "config.yaml"))
		if err != nil {
			_ = WriteText(filepath.Join(outDir, "error.txt"), err.Error()+"\n")
		} else if FileExists(filepath.Join(inDir, "gcloud_service.json")) {
			rep, err := Verify(cfg, inDir)
			if err != nil {
				_ = WriteText(filepath.Join(outDir, "error.txt"), err.Error()+"\n")
			} else {
				b, _ := json.MarshalIndent(rep, "", "  ")
				_ = WriteText(filepath.Join(outDir, "verify_report.json"), string(b)+"\n")
			}
		} else {
			arts, err := Render(cfg)
			if err != nil {
				_ = WriteText(filepath.Join(outDir, "error.txt"), err.Error()+"\n")
			} else {
				_ = WriteArtifacts(outDir, arts)
			}
		}

		if err := DiffTrees(expDir, outDir); err != nil {
			fmt.Fprintf(os.Stderr, "demo mismatch %s: %v\n", c, err)
			return 1
		}
	}

	fmt.Println("OK")
	return 0
}
