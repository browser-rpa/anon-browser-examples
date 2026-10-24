// Command plugin implementing the public anon-browser plugin_api_version=1
// JSON protocol. It intentionally has no direct browser or filesystem access.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type invocation struct {
	Input map[string]any `json:"input"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var request invocation
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			write(map[string]any{"ok": false, "error": "invalid_invocation"})
			return
		}
		value, ok := request.Input["value"].(string)
		if !ok {
			write(map[string]any{"ok": false, "error": "value_must_be_string"})
			return
		}
		write(map[string]any{
			"ok": true,
			"result": map[string]any{
				"next_port": "success",
				"output":    map[string]string{"value": value},
				"variables": map[string]string{"echo_value": value},
			},
		})
		return
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func write(value any) {
	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
