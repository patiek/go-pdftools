package pdftk

// Adjusts how a command runs pdftk.
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
		cmd.outputArgs = append(cmd.outputArgs, "flatten")
	}
}

// Directory for temp copies of inputs that are not unread *os.Files; pdftk
// must be able to open files there. Copies are mode 0600 and removed when the
// command returns. Defaults to os.TempDir().
func OptionTempDir(dir string) Option {
	return func(cmd *command) {
		cmd.tempDir = dir
	}
}
