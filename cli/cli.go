package cli

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"strings"
	"bufio"
	"io/ioutil"
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
			authToken = promptAuthToken()
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
		cronTasks, err := ParseAndApproveCronTasks()
		if err != nil {
			fmt.Println("Error parsing crontab:", err)
			os.Exit(1)
		}
		_ = cronTasks // Suppress the "declared and not used" warning
	}

	// Rest of your CLI tool logic...
}

func promptAuthToken() string {
	var authToken string

	fmt.Print("Enter the authentication token for the BetterUptime API: ")
	fmt.Scanln(&authToken)
	fmt.Println()

	// Trim any leading or trailing whitespace from the input
	authToken = strings.TrimSpace(authToken)

	return authToken
}

// Function to parse crontab and prompt user for approval
func ParseAndApproveCronTasks() ([]CronTask, error) {
	// Read the crontab file
	data, err := ioutil.ReadFile("/var/spool/cron/crontabs/" + crontabUser)
	if err != nil {
		return nil, err
	}

	// Split file content into lines
	lines := strings.Split(string(data), "\n")

	approvedCronTasks := []CronTask{}

	for _, line := range lines {
		// Ignore empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// If cron task contains "uptime.betterstack.com", skip and inform user
		if strings.Contains(line, "uptime.betterstack.com") {
			fmt.Println("Skipping cron task containing 'uptime.betterstack.com':", line)
			continue
		}

		// Display the cron task and ask for approval
		fmt.Println("Cron task:", line)
		if promptApproval() {
			fields := strings.Fields(line)
			if len(fields) < 6 {
				fmt.Println("Skipping invalid cron task:", line)
				continue
			}

			// First 5 fields are the spec, the rest is the task
			spec := strings.Join(fields[:5], " ")
			task := strings.Join(fields[5:], " ")

			approvedCronTasks = append(approvedCronTasks, CronTask{
				spec: spec,
				task: task,
			})
		}
	}

	return approvedCronTasks, nil
}

type CronTask struct {
	spec string
	task string
}

// Function to prompt the user for approval
func promptApproval() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Approve this cron task? (y/n): ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	return strings.ToLower(text) == "y"
}