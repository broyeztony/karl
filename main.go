package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"karl/ast"
	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	sub := os.Args[1]
	if sub == "-h" || sub == "--help" {
		usage()
		return
	}
	switch sub {
	case "parse":
		format, positional, err := parseArgs(os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			usage()
			os.Exit(2)
		}
		if len(positional) != 1 {
			usage()
			os.Exit(2)
		}
		if err := validateExtension(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(2)
		}
		data, err := readInput(positional[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "read error: %v\n", err)
			os.Exit(1)
		}
		program, err := parseProgram(data, displayName(positional[0]))
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		switch format {
		case "pretty":
			fmt.Print(ast.Format(program))
		case "json":
			out, err := ast.FormatJSON(program)
			if err != nil {
				fmt.Fprintf(os.Stderr, "format error: %v\n", err)
				os.Exit(1)
			}
			fmt.Print(out)
		default:
			fmt.Fprintf(os.Stderr, "unknown format: %s\n", format)
			os.Exit(2)
		}
	case "run":
		positional, err := parseArgsNoFlags(os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			usage()
			os.Exit(2)
		}
		if len(positional) != 1 {
			usage()
			os.Exit(2)
		}
		if err := validateExtension(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(2)
		}
		data, err := readInput(positional[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "read error: %v\n", err)
			os.Exit(1)
		}
		filename := displayName(positional[0])
		program, err := parseProgram(data, filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		val, err := runProgram(program, string(data), filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, interpreter.FormatRuntimeError(err, string(data), filename))
			os.Exit(1)
		}
		fmt.Println(val.Inspect())
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", sub)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  karl parse <file.k> [--format=pretty|json]\n")
	fmt.Fprintf(os.Stderr, "  karl run <file.k>\n")
	fmt.Fprintf(os.Stderr, "  <file> can be '-' to read from stdin\n")
}

func parseArgs(args []string) (string, []string, error) {
	format := "pretty"
	positional := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			return format, positional, fmt.Errorf("help requested")
		case arg == "-":
			positional = append(positional, arg)
		case strings.HasPrefix(arg, "--format="):
			format = strings.TrimPrefix(arg, "--format=")
		case arg == "--format":
			if i+1 >= len(args) {
				return format, positional, fmt.Errorf("--format requires a value")
			}
			format = args[i+1]
			i++
		case strings.HasPrefix(arg, "-"):
			return format, positional, fmt.Errorf("unknown flag: %s", arg)
		default:
			positional = append(positional, arg)
		}
	}

	return format, positional, nil
}

func parseArgsNoFlags(args []string) ([]string, error) {
	positional := []string{}
	for _, arg := range args {
		if arg == "-" {
			positional = append(positional, arg)
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
		positional = append(positional, arg)
	}
	return positional, nil
}

func readInput(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func displayName(path string) string {
	if path == "-" {
		return "<stdin>"
	}
	return path
}

func validateExtension(path string) error {
	if path == "-" {
		return nil
	}
	if !strings.HasSuffix(path, ".k") {
		return fmt.Errorf("file must have .k extension: %s", path)
	}
	return nil
}

func parseProgram(data []byte, filename string) (*ast.Program, error) {
	p := parser.New(lexer.New(string(data)))
	program := p.ParseProgram()
	if errs := p.ErrorsDetailed(); len(errs) > 0 {
		return nil, fmt.Errorf("%s", parser.FormatParseErrors(errs, string(data), filename))
	}
	return program, nil
}

func runProgram(program *ast.Program, source string, filename string) (interpreter.Value, error) {
	eval := interpreter.NewEvaluatorWithSourceAndFilename(source, filename)
	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if err != nil {
		return nil, err
	}
	if sig != nil {
		return nil, fmt.Errorf("break/continue outside loop")
	}
	return val, nil
}
