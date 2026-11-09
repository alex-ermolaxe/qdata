package engine

// CommandSpecs returns the command catalog used by the engine interfaces.
func CommandSpecs() []CommandSpec {
	return []CommandSpec{
		whereCommand,
		selectCommand,
		excludeCommand,
		sortCommand,
		showCommand,
		saveCommand,
		flattenCommand,
		transformCommand,
		resetCommand,
		countCommand,
		schemaCommand,
		infoCommandSpec(),
		exitCommand,
	}
}
