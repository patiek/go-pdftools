package pdftk

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

const fakePDFtkEnv = "GO_PDFTOOLS_FAKE_PDFTK"

// When re-executed with fakePDFtkEnv set, the test binary stands in for pdftk
// so that commands can be tested without pdftk installed.
func TestMain(m *testing.M) {
	switch mode := os.Getenv(fakePDFtkEnv); mode {
	case "":
		os.Exit(m.Run())
	case "echo":
		// echo the arguments followed by stdin
		fmt.Println(strings.Join(os.Args[1:], " "))
		_, _ = io.Copy(os.Stdout, os.Stdin)
	case "dump_data":
		if len(os.Args) != 5 || !strings.HasPrefix(os.Args[1], "A=") || os.Args[2] != "dump_data" || os.Args[3] != "output" || os.Args[4] != "-" {
			fmt.Fprintf(os.Stderr, "Error: unexpected arguments %q\n", os.Args[1:])
			os.Exit(1)
		}
		fmt.Printf("InfoBegin\nInfoKey: Title\nInfoValue: %s\nNumberOfPages: 7\nPageMediaBegin\n", os.Args[1])
	case "fail":
		fmt.Fprintln(os.Stderr, "Error: something went wrong")
		os.Exit(1)
	case "sleep":
		time.Sleep(time.Minute)
	default:
		fmt.Fprintf(os.Stderr, "unknown fake pdftk mode %q\n", mode)
		os.Exit(2)
	}
	os.Exit(0)
}

// Run the test binary in place of pdftk in the given mode.
func fakePDFtk(t *testing.T, mode string) Option {
	t.Setenv(fakePDFtkEnv, mode)
	return OptionExecutable(os.Args[0])
}
