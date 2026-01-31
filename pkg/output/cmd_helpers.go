package output

import "fmt"

// KeyValue prints a styled key-value pair, falling back to plain text when piped.
func KeyValue(key, value string) {
	if IsTerminal() {
		fmt.Printf("%s\n%s\n\n", keyStyle.Render(key), valueStyle.Render(value))
	} else {
		fmt.Printf("%s\n%s\n\n", key, value)
	}
}

// KeyValueWithMetadata prints a styled key-value pair with ordered metadata fields.
func KeyValueWithMetadata(key, value string, metadata [][2]string) {
	if IsTerminal() {
		fmt.Println(keyStyle.Render(key))
		fmt.Println(valueStyle.Render(value))
		for _, kv := range metadata {
			fmt.Printf("%s %s\n", valueStyle.Render(kv[0]+":"), kv[1])
		}
	} else {
		fmt.Println(key)
		fmt.Println(value)
		for _, kv := range metadata {
			fmt.Printf("%s: %s\n", kv[0], kv[1])
		}
	}
	fmt.Println()
}
