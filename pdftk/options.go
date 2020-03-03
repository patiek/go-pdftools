package pdftk

type Option func(cmd command)

// Set executable name instead of using default "pdftk"
func OptionExecutable(name string) Option {
	return func(cmd command) {
		// replace the command entirely
		c := createCmd(name, cmd.Stdout, cmd.Stdin, cmd.Args...)
		cmd.Cmd = c.Cmd
	}
}

// Flatten the PDF before output
func OptionFlatten() Option {
	return func(cmd command) {
		cmd.Args = append(cmd.Args, "flatten")
	}
}
