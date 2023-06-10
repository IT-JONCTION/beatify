package cli

import (
	"github.com/spf13/pflag"
	"fmt"
	"os"
	"os/user"
	"github.com/IT-JONCTION/beatify/config"
	"github.com/IT-JONCTION/beatify/crontab"
	"github.com/IT-JONCTION/beatify/heartbeat"
	"golang.org/x/time/rate"
	"time"
)

var (
	authToken   string
	crontabUser string
	showHelp    bool
	heartbeatGroupID string
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
        Provide the authentication token for the BetterUptime API. If not
        provided, the tool will prompt for it during runtime.

    -u, --user USER
        Optional. The crontab user to edit. If not provided, the tool will
        default to the current user's crontab.

    -h, --help
        Display the help message and exit.

    -g, --heartbeat-group HEARTBEAT_GROUP
        Optional. The heartbeat group to add the heartbeat to. If not provided,
        the tool will default to creating the heartbeats without a group.

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
	pflag.StringVarP(&authToken, "auth-token", "a", "", "Authentication token for the BetterUptime API")
	pflag.StringVarP(&crontabUser, "user", "u", "", "Crontab user to edit")
	pflag.StringVarP(&heartbeatGroupID, "heartbeat-group", "g", "", "Heartbeat group to add the heartbeat to")
	pflag.BoolVarP(&showHelp, "help", "h", false, "Show help message")

	// Customize usage message
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", manpageTemplate)
	}
}

func HandleCommandLineOptions() {
	pflag.Parse()

	if showHelp {
		// Display the help message and exit
		pflag.Usage()
		os.Exit(0)
	}

	// Check if authToken is set
	if authToken == "" {
		// Prompt the user to enter the authToken
		authToken = config.PromptAuthToken()
	}

	// Check if heartbeatGroup is set
	if heartbeatGroupID != "" {
		// Get the ID of the heartbeat group if it exists
		heartbeatGroupID = heartbeat.GetHeartbeatGroupID(authToken, heartbeatGroupID)
		if heartbeatGroupID == "" {
				// If heartbeatGroup does not exist, create it
				var err error
				heartbeatGroupID, err = heartbeat.CreateHeartbeatGroup(authToken, heartbeatGroupID)
				if err != nil {
						fmt.Println("Error creating heartbeat group:", err)
						os.Exit(1)
				}
		}
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

		limiter := rate.NewLimiter(3, 1) // 3 requests per second, no burst
	
		// Iterate over cronTasks and call PrepareConfigJson for each task
		for i, cronTask := range cronTasks {

			ctx := limiter.ReserveN(time.Now(), 1)
			if !ctx.OK() {
				fmt.Println("Waiting for API.")
				return
			}
		
			delay := ctx.Delay()
			time.Sleep(delay)

			data, err := heartbeat.PrepareConfigJson(cronTask.Spec, cronTask.Name, heartbeatGroupID)
			if err != nil {
					fmt.Println("Error preparing config JSON:", err)
					continue // Skip to the next iteration of the loop
			}

			// Create the Heartbeat
			responseBodyURL, err := heartbeat.CreateHeartbeat(authToken, data)
			if err != nil {
					fmt.Println("Error creating heartbeat:", err)
					continue // Skip to the next iteration of the loop
			}
			fmt.Println("Heartbeat created successfully:", responseBodyURL)

			// Set cronTask.HeartbeatURL to the response URL
			cronTasks[i].HeartbeatURL = responseBodyURL
		}

		err = crontab.AppendCronsCommand(cronTasks, crontabUser)
		if err != nil {
			fmt.Println("Error appending curl command to cron tasks:", err)
		}
		fmt.Println("Curl commands appended to cron tasks successfully.")
	}	
}
