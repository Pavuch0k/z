package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pavuch0k/z/internal/compiler"
)

var Version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "version", "-v", "--version":
		fmt.Printf("zc %s\n", Version)
	case "build":
		if err := cmdBuild(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "run":
		if err := cmdRun(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `zc — транслятор языка Z

Использование:
  zc version
  zc run <файл.z>
  zc build <файл.z> [-o <имя>]
`)
}

func cmdBuild(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("build: missing input file")
	}
	in := args[0]
	out := ""
	for i := 1; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			out = args[i+1]
			i++
		}
	}
	if out == "" {
		base := strings.TrimSuffix(filepath.Base(in), filepath.Ext(in))
		out = base
	}
	return compiler.CompileFile(in, out)
}

func cmdRun(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("run: missing input file")
	}
	return compiler.RunFile(args[0])
}
