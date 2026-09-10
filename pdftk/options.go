package pdftk

type Option func(cmd *command)

// Set executable name instead of using default "pdftk"
func OptionExecutable(name string) Option {
	return func(cmd *command) {
		cmd.name = name
	}
}

// Flatten the PDF before output
func OptionFlatten() Option {
	return func(cmd *command) {
		cmd.args = append(cmd.args, "flatten")
	}
}
