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

	if err := ResetOutDir(*outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		_ = WriteText(filepath.Join(*outDir, "error.txt"), err.Error()+"\n")
		return 0
	}

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

	if err := ResetOutDir(*outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		_ = WriteText(filepath.Join(*outDir, "error.txt"), err.Error()+"\n")
		return 0
	}

	rep, err := Verify(cfg, *snapDir)
	if err != nil {
		_ = WriteText(filepath.Join(*outDir, "error.txt"), err.Error()+"\n")
		return 0
	}
	b, _ := json.MarshalIndent(rep, "", "  ")
	if err := WriteText(filepath.Join(*outDir, "verify_report.json"), string(b)+"\n"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func cmdDemo(args []string) int {
	fs := flag.NewFlagSet("demo", flag.ContinueOnError)
	outBase := fs.String("out", "./out", "output base dir")
	_ = fs.Parse(args)

	if err := ResetOutDir(*outBase); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	cases, err := ListCases("./fixtures/input")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	for _, c := range cases {
		inDir := filepath.Join("./fixtures/input", c)
		expDir := filepath.Join("./fixtures/expected", c)
		outDir := filepath.Join(*outBase, c)

		if err := ResetOutDir(outDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}

		cfgPath := filepath.Join(inDir, "config.yaml")
		expErr := filepath.Join(expDir, "error.txt")

		// Expected-fail: must produce only error.txt.
		if FileExists(expErr) {
			// Two modes: verify (if snapshot file exists), else render.
			snapPath := filepath.Join(inDir, "gcloud_service.json")
			if FileExists(snapPath) {
				cfg, err := LoadConfig(cfgPath)
				if err == nil {
					rep, verr := Verify(cfg, inDir)
					if verr == nil {
						b, _ := json.MarshalIndent(rep, "", "  ")
						_ = WriteText(filepath.Join(outDir, "verify_report.json"), string(b)+"\n")
						err = nil
					} else {
						err = verr
					}
				}
				if err == nil {
					fmt.Printf("demo mismatch %s: expected error, got none\n", c)
					return 1
				}
				_ = WriteText(filepath.Join(outDir, "error.txt"), err.Error()+"\n")
			} else {
				cfg, err := LoadConfig(cfgPath)
				if err == nil {
					arts, rerr := Render(cfg)
					if rerr == nil {
						_ = WriteArtifacts(outDir, arts)
						err = nil
					} else {
						err = rerr
					}
				}
				if err == nil {
					fmt.Printf("demo mismatch %s: expected error, got none\n", c)
					return 1
				}
				_ = WriteText(filepath.Join(outDir, "error.txt"), err.Error()+"\n")
			}

			if err := DiffTrees(outDir, expDir); err != nil {
				fmt.Println(err.Error())
				return 1
			}
			continue
		}

		cfg, err := LoadConfig(cfgPath)
		if err != nil {
			fmt.Printf("demo mismatch %s: %v\n", c, err)
			return 1
		}

		// Two modes: verify (if snapshot file exists), else render.
		snapPath := filepath.Join(inDir, "gcloud_service.json")
		if FileExists(snapPath) {
			rep, err := Verify(cfg, inDir)
			if err != nil {
				fmt.Printf("demo mismatch %s: %v\n", c, err)
				return 1
			}
			b, _ := json.MarshalIndent(rep, "", "  ")
			if err := WriteText(filepath.Join(outDir, "verify_report.json"), string(b)+"\n"); err != nil {
				fmt.Printf("demo mismatch %s: %v\n", c, err)
				return 1
			}
		} else {
			arts, err := Render(cfg)
			if err != nil {
				fmt.Printf("demo mismatch %s: %v\n", c, err)
				return 1
			}
			if err := WriteArtifacts(outDir, arts); err != nil {
				fmt.Printf("demo mismatch %s: %v\n", c, err)
				return 1
			}
		}

		if err := DiffTrees(outDir, expDir); err != nil {
			fmt.Println(err.Error())
			return 1
		}
	}
	fmt.Println("OK")
	return 0
}
