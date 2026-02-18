package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func usage() {
	fmt.Fprintln(os.Stderr, "pfdeploy demo|render|snapshot|verify ...")
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
	case "snapshot":
		os.Exit(cmdSnapshot(os.Args[2:]))
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
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *cfgPath == "" || *outDir == "" {
		return 2
	}

	if err := ResetOutDir(*outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		if werr := WriteText(filepath.Join(*outDir, "error.txt"), err.Error()+"\n"); werr != nil {
			fmt.Fprintln(os.Stderr, werr)
			return 1
		}
		return 0
	}

	arts, err := Render(cfg)
	if err != nil {
		if werr := WriteText(filepath.Join(*outDir, "error.txt"), err.Error()+"\n"); werr != nil {
			fmt.Fprintln(os.Stderr, werr)
			return 1
		}
		return 0
	}
	if err := WriteArtifacts(*outDir, arts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func cmdSnapshot(args []string) int {
	fs := flag.NewFlagSet("snapshot", flag.ContinueOnError)
	inDir := fs.String("in", "", "input dir (contains gcloud_service_raw.json)")
	outDir := fs.String("out", "", "out dir")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *inDir == "" || *outDir == "" {
		return 2
	}

	if err := ResetOutDir(*outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	s, err := Snapshot(*inDir)
	if err != nil {
		if werr := WriteText(filepath.Join(*outDir, "error.txt"), err.Error()+"\n"); werr != nil {
			fmt.Fprintln(os.Stderr, werr)
			return 1
		}
		return 0
	}
	if err := WriteSnapshot(*outDir, s); err != nil {
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
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *cfgPath == "" || *snapDir == "" || *outDir == "" {
		return 2
	}

	if err := ResetOutDir(*outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		if werr := WriteText(filepath.Join(*outDir, "error.txt"), err.Error()+"\n"); werr != nil {
			fmt.Fprintln(os.Stderr, werr)
			return 1
		}
		return 0
	}

	rep, err := Verify(cfg, *snapDir)
	if err != nil {
		if werr := WriteText(filepath.Join(*outDir, "error.txt"), err.Error()+"\n"); werr != nil {
			fmt.Fprintln(os.Stderr, werr)
			return 1
		}
		return 0
	}
	b, merr := json.MarshalIndent(rep, "", "  ")
	if merr != nil {
		fmt.Fprintln(os.Stderr, merr)
		return 1
	}
	if err := WriteText(filepath.Join(*outDir, "verify_report.json"), string(b)+"\n"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func cmdDemo(args []string) int {
	fs := flag.NewFlagSet("demo", flag.ContinueOnError)
	outBase := fs.String("out", "./out", "output base dir")
	if err := fs.Parse(args); err != nil {
		return 2
	}

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

		// Snapshot case: normalize raw snapshot into gcloud_service.json + manifest.sha256.
		rawSnapPath := filepath.Join(inDir, "gcloud_service_raw.json")
		if FileExists(rawSnapPath) {
			s, err := Snapshot(inDir)
			if err != nil {
				fmt.Printf("demo mismatch %s: %v\\n", c, err)
				return 1
			}
			if err := WriteSnapshot(outDir, s); err != nil {
				fmt.Printf("demo mismatch %s: %v\\n", c, err)
				return 1
			}
			if err := DiffTrees(outDir, expDir); err != nil {
				fmt.Println(err.Error())
				return 1
			}
			continue
		}

		// Expected-fail: must produce only error.txt.
		if FileExists(expErr) {
			// Two modes: verify (if snapshot file exists), else render.
			snapPath := filepath.Join(inDir, "gcloud_service.json")
			if FileExists(snapPath) {
				cfg, err := LoadConfig(cfgPath)
				if err == nil {
					rep, verr := Verify(cfg, inDir)
					if verr == nil {
						b, merr := json.MarshalIndent(rep, "", "  ")
						if merr != nil {
							fmt.Printf("demo mismatch %s: %v\n", c, merr)
							return 1
						}
						if werr := WriteText(filepath.Join(outDir, "verify_report.json"), string(b)+"\n"); werr != nil {
							fmt.Printf("demo mismatch %s: %v\n", c, werr)
							return 1
						}
						err = nil
					} else {
						err = verr
					}
				}
				if err == nil {
					fmt.Printf("demo mismatch %s: expected error, got none\n", c)
					return 1
				}
				if werr := WriteText(filepath.Join(outDir, "error.txt"), err.Error()+"\n"); werr != nil {
					fmt.Printf("demo mismatch %s: %v\n", c, werr)
					return 1
				}
			} else {
				cfg, err := LoadConfig(cfgPath)
				if err == nil {
					arts, rerr := Render(cfg)
					if rerr == nil {
						if werr := WriteArtifacts(outDir, arts); werr != nil {
							fmt.Printf("demo mismatch %s: %v\n", c, werr)
							return 1
						}
						err = nil
					} else {
						err = rerr
					}
				}
				if err == nil {
					fmt.Printf("demo mismatch %s: expected error, got none\n", c)
					return 1
				}
				if werr := WriteText(filepath.Join(outDir, "error.txt"), err.Error()+"\n"); werr != nil {
					fmt.Printf("demo mismatch %s: %v\n", c, werr)
					return 1
				}
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
			b, merr := json.MarshalIndent(rep, "", "  ")
			if merr != nil {
				fmt.Printf("demo mismatch %s: %v\n", c, merr)
				return 1
			}
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
