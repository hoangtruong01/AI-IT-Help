// Command check_openapi_coverage fails when the runtime API routes and the
// OpenAPI operation inventory diverge. It intentionally uses only the Go
// standard library so it can run in developer machines and CI without another
// package manager.
package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const openAPIPath = "docs/openapi/eomp-openapi-spec.yaml"

var (
	routePattern     = regexp.MustCompile(`Handle(?:Func)?\(\s*"(GET|POST|PUT|PATCH|DELETE)\s+(/api/v1/[^"\s]+)"`)
	pathPattern      = regexp.MustCompile(`^  (/api/v1/[^:]+):\s*$`)
	operationPattern = regexp.MustCompile(`^    (get|post|put|patch|delete):(?:\s|$)`)
)

func main() {
	runtime, err := runtimeOperations("services")
	if err != nil {
		fail(err)
	}
	documented, err := documentedOperations(openAPIPath)
	if err != nil {
		fail(err)
	}

	missing := difference(runtime, documented)
	extra := difference(documented, runtime)
	if len(missing) != 0 || len(extra) != 0 {
		fmt.Fprintf(os.Stderr, "OpenAPI coverage mismatch: runtime=%d documented=%d\n", len(runtime), len(documented))
		printOperations("Missing from OpenAPI", missing)
		printOperations("Not registered at runtime", extra)
		os.Exit(1)
	}

	fmt.Printf("OpenAPI coverage OK: %d/%d runtime operations documented\n", len(documented), len(runtime))
}

func runtimeOperations(root string) (map[string]struct{}, error) {
	operations := make(map[string]struct{})
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "main.go" || filepath.Base(filepath.Dir(path)) != "server" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		matches := routePattern.FindAllStringSubmatch(string(content), -1)
		if err := validateRouteCompatibility(path, matches); err != nil {
			return err
		}
		for _, match := range matches {
			operations[match[1]+" "+match[2]] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan runtime routes: %w", err)
	}
	if len(operations) == 0 {
		return nil, fmt.Errorf("no /api/v1 runtime routes found under %s", root)
	}
	return operations, nil
}

func validateRouteCompatibility(path string, matches [][]string) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("conflicting runtime routes in %s: %v", path, recovered)
		}
	}()
	mux := http.NewServeMux()
	for _, match := range matches {
		mux.HandleFunc(match[1]+" "+match[2], func(http.ResponseWriter, *http.Request) {})
	}
	return nil
}

func documentedOperations(path string) (map[string]struct{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open OpenAPI document: %w", err)
	}
	defer file.Close()

	operations := make(map[string]struct{})
	currentPath := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if match := pathPattern.FindStringSubmatch(line); match != nil {
			currentPath = match[1]
			continue
		}
		if match := operationPattern.FindStringSubmatch(line); match != nil && currentPath != "" {
			operations[strings.ToUpper(match[1])+" "+currentPath] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read OpenAPI document: %w", err)
	}
	if len(operations) == 0 {
		return nil, fmt.Errorf("no /api/v1 operations found in %s", path)
	}
	return operations, nil
}

func difference(left, right map[string]struct{}) []string {
	result := make([]string, 0)
	for operation := range left {
		if _, exists := right[operation]; !exists {
			result = append(result, operation)
		}
	}
	sort.Strings(result)
	return result
}

func printOperations(label string, operations []string) {
	if len(operations) == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, label+":")
	for _, operation := range operations {
		fmt.Fprintln(os.Stderr, "  - "+operation)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
