package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/runableapp/when-ng"
	"github.com/runableapp/when-ng/rules/common"
	"github.com/runableapp/when-ng/rules/en"
)

func main() {
	baseFlag := flag.String("base", "", "base time in RFC3339 (defaults to now)")
	flag.Parse()

	baseTime := time.Now()
	if *baseFlag != "" {
		parsed, err := time.Parse(time.RFC3339, *baseFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid -base value: %v\n", err)
			os.Exit(2)
		}
		baseTime = parsed
	}

	// Read test cases from test_cases.txt
	// Get the directory of the current source file
	_, filename, _, _ := runtime.Caller(0)
	testFile := filepath.Join(filepath.Dir(filename), "test_cases.txt")

	file, err := os.Open(testFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening test_cases.txt: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var inputs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip blank lines and lines starting with "#"
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		inputs = append(inputs, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading test_cases.txt: %v\n", err)
		os.Exit(1)
	}

	// If command-line arguments provided, use those instead
	if len(flag.Args()) > 0 {
		inputs = flag.Args()
	}

	if len(inputs) == 0 {
		fmt.Fprintf(os.Stderr, "no test cases found in test_cases.txt\n")
		os.Exit(1)
	}

	w := when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)

	fmt.Printf("==== base: %s\n", baseTime.Format(time.RFC3339))
	for _, text := range inputs {
		res, err := w.Parse(text, baseTime)
		if err != nil {
			fmt.Printf("%-40s X\n", text)
			continue
		}
		if res == nil {
			fmt.Printf("%-40s X\n", text)
			continue
		}
		fmt.Printf("%-40s %s\n", text, res.Time.Format(time.RFC3339))
	}
}
