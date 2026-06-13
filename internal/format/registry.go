package format

// registry - registry of registered formats
var registry = map[string]Format{}

// Register registers a new format by name
func Register(name string, fileFormat Format) {
	registry[name] = fileFormat
}
