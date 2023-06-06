package cli

import (
	"flag"
	"fmt"
	"os"
)

var (
	authToken   string
	crontabUser string
	showHelp    bool
)

func init() {
	// Define command-line flags
	flag.StringVar(&authToken, "a", "", "Authentication token for the BetterUptime API")
	flag.StringVar(&authToken, "auth-token", "", "Authentication token for the BetterUptime API (shorthand)")

	flag.StringVar(&crontabUser, "u", "", "Crontab user to edit")
	flag.StringVar(&crontabUser, "user", "", "Crontab user to edit (shorthand)")

	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.BoolVar(&showHelp, "help", false, "Show help message (shorthand)")

	// Customize usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", manpageTemplate)
	}
}

func HandleCommandLineOptions() {
	flag.Parse()

	if showHelp {
		// Display the help message and exit
		flag.Usage()
		os.Exit(0)
	}

	// Rest of your CLI tool logic...
}
