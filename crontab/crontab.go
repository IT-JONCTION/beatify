package crontab

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"os/exec"
)

// Utility function to get the crontab file path for a given user
func getCrontabFilePath(crontabUser string) (string, error) {
	// Get the crontab file path
	crontabFile, err := getCrontabFilePath(crontabUser)
	_, err := os.Stat(crontabFile)
	if err != nil {
		return "", err
	}
	return crontabFile, nil
}

// Function to parse crontab and prompt user for approval
func ParseAndApproveCronTasks(crontabUser string) ([]CronTask, error) {
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
			// Prompt the user to enter the name for the heartbeat
			name, err := promptHeartbeatName()
			if err != nil {
				return nil, err
			}

			fields := strings.Fields(line)
			if len(fields) < 6 {
				fmt.Println("Skipping invalid cron task:", line)
				continue
			}

			spec := strings.Join(fields[:5], " ")
			task := strings.Join(fields[5:], " ")

			approvedCronTasks = append(approvedCronTasks, CronTask{
				Spec: spec,
				Task: task,
				Name: name,
			})
		}
	}

	return approvedCronTasks, nil
}

func promptHeartbeatName() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the name for the heartbeat: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	name = strings.TrimSpace(name)

	// Perform basic validation to prevent command injection
	if strings.ContainsAny(name, `";$|><&`) {
		return "", fmt.Errorf("Invalid heartbeat name")
	}

	return name, nil
}

type CronTask struct {
	Spec string
	Task string
	Name string
	HeartbeatURL string
}

// Function to prompt the user for approval
func promptApproval() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Approve this cron task? (y/n): ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	return strings.ToLower(text) == "y"
}

// Function to append curl command to crontab tasks
func AppendCurlsCommand(urls []string, crontabUser string) error {
	// Get the crontab file path
	crontabFile, err := getCrontabFilePath(crontabUser)
	if err != nil {
		return err
	}

	for _, url := range urls {
		// Construct the curl command string to append to the task
		curlCommand := fmt.Sprintf(`curl -s -o /dev/null -w "%%{http_code}" %s`, url)

		// Use sed to append the curl command to the task
		cmd := exec.Command("sed", "-i", fmt.Sprintf(`$s#$# && %s#`, curlCommand), crontabFile)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to append curl command to task: %w", err)
		}
	}

	// Reload the cron system to apply the changes
	reloadCmd := exec.Command("systemctl", "reload", "cron")
	if err := reloadCmd.Run(); err != nil {
		return fmt.Errorf("failed to reload cron system: %w", err)
	}

	return nil
}
