package cli

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"github.com/IT-JONCTION/beatify/config"
	"github.com/IT-JONCTION/beatify/crontab"
	"github.com/IT-JONCTION/beatify/heartbeat"
)

var (
	authToken   string
	crontabUser string
	showHelp    bool
)

var manpageTemplate = `
BEATIFY(1)              User Commands             BEATIFY(1)

NAME
    beatify - Automate heartbeats for monitoring cron tasks with BetterUptime

SYNOPSIS
    beatify [OPTIONS]

DESCRIPTION
    The beatify is a command-line tool that automates the creation of heartbeats
    for monitoring cron tasks using the BetterUptime service. It reads the user's
    crontab, presents each cron task for approval to create a heartbeat, calls
    the BetterUptime API to create the approved heartbeats, and updates the
    crontab to append a curl request to each approved cron task.

OPTIONS
    -a, --auth-token AUTH_TOKEN
        Optional. The authentication token for the BetterUptime API. If not
        provided, the tool will prompt for it during runtime.

    -u, --user USER
        Optional. The crontab user to edit. If not provided, the tool will
        default to the current user's crontab.

    -h, --help
        Display the help message and exit.

EXAMPLES
    To run Beatify and create heartbeats for cron tasks:
        beatify -a YOUR_AUTH_TOKEN -u www-data

EXIT STATUS
    0 if successful, or an error code if an error occurs.

REPORTING BUGS
    Report bugs to the GitHub repository: https://github.com/IT-JONCTION/beatify

AUTHOR
    Your Name <wayne@it-jonction-lab.com>

COPYRIGHT
    Copyright © 2023 IT Jonction Lab. This is free software; see the source
    code for copying conditions. There is NO warranty; not even for MERCHANTABILITY
    or FITNESS FOR A PARTICULAR PURPOSE.

SEE ALSO
    The BetterUptime API documentation: https://docs.betteruptime.com/api/
`


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
	
		// Check if authToken is set
		if authToken == "" {
			// Prompt the user to enter the authToken
			authToken = config.PromptAuthToken()
		}

		// Check if crontabUser option is set
		if crontabUser == "" {
			// If not set, obtain currently logged-in user
			currentUser, err := user.Current()
			if err != nil {
					fmt.Println("Error obtaining current user:", err)
					os.Exit(1)
			}
			crontabUser = currentUser.Username
	}

	if crontabUser != "" {
		// Parse and approve cron tasks
		cronTasks, err := crontab.ParseAndApproveCronTasks(crontabUser)
		if err != nil {
			fmt.Println("Error parsing crontab:", err)
			os.Exit(1)
		}
		// Iterate over cronTasks and call PrepareConfigJson for each task
		var jsonData string
		for _, cronTask := range cronTasks {
			data, err := heartbeat.PrepareConfigJson(cronTask.Spec, cronTask.Name)
			if err != nil {
				fmt.Println("Error preparing config JSON:", err)
				continue // Skip to the next iteration of the loop
			}
			jsonData = data

			fmt.Println(jsonData)
			// Create the Heartbeat
			responseBodyUrl, err := heartbeat.CreateHeartbeat(authToken, data)
			if err != nil {
				fmt.Println("Error creating heartbeat:", err)
				continue // Skip to the next iteration of the loop
			}
			fmt.Println("Heartbeat created successfully:", responseBodyUrl)
		}
	}
}