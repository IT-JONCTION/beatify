package main

import (
	"fmt"
	"os"
	"time"
	"github.com/robfig/cron"
	"text/template"
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
    Report bugs to the GitHub repository: https://github.com/your/beatify

AUTHOR
    Your Name <your.email@example.com>

COPYRIGHT
    Copyright © 2023 Your Organization. This is free software; see the source
    code for copying conditions. There is NO warranty; not even for MERCHANTABILITY
    or FITNESS FOR A PARTICULAR PURPOSE.

SEE ALSO
    The BetterUptime API documentation: https://docs.betteruptime.com/api/
`

func main() {
	// Check if the user is requesting the manpage
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		// Generate and display the manpage
		tmpl := template.Must(template.New("manpage").Parse(manpageTemplate))
		err := tmpl.Execute(os.Stdout, nil)
		if err != nil {
			fmt.Println("Error rendering manpage:", err)
			return
		}
		return
	}

	// Check if both the crontab schedule and heartbeat name are provided as command-line arguments
	if len(os.Args) < 3 {
		fmt.Println("Please provide the crontab schedule and heartbeat name as command-line arguments.")
		return
	}

	// Read the crontab schedule and heartbeat name from command-line arguments
	crontab := os.Args[1]
	heartbeatName := os.Args[2]

	// Create a new cron parser
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	// Parse the crontab schedule
	schedule, err := parser.Parse(crontab)
	if err != nil {
		fmt.Println("Error parsing crontab schedule:", err)
		return
	}

	// Calculate the period in seconds
	period := int(schedule.Next(time.Now()).Sub(time.Now()).Seconds())

	// Calculate the grace period as approximately 20% of the period
	grace := int(float64(period) * 0.2)

	// Create the JSON representation
	jsonData := fmt.Sprintf(`{
		"name": "%s",
		"period": %d,
		"grace": %d
	}`, heartbeatName, period, grace)

	fmt.Println(jsonData)
}
