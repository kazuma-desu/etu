package output

import "fmt"

// KeyValue prints the given key and value.
// When running in a terminal, the key and value are rendered with terminal styles;
// otherwise they are printed as plain text. An empty line is appended after the pair.
func KeyValue(key, value string) {
	if IsTerminal() {
		fmt.Printf("%s\n%s\n\n", keyStyle.Render(key), valueStyle.Render(value))
	} else {
		fmt.Printf("%s\n%s\n\n", key, value)
	}
}

// KeyValueWithMetadata prints the given key and value followed by ordered metadata fields.
// When output is a terminal, the key, value and metadata labels are rendered with styling;
// otherwise plain text is printed. Each metadata entry is emitted as "label: value".
// A single blank line is printed after the entire block.
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