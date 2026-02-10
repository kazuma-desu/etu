package output

import "fmt"

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
