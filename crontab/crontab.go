package crontab

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"os/exec"
	"net/url"
)

// Utility function to get the crontab file path for a given user
func getCrontabFilePath(crontabUser string) (string, error) {
	// Get the crontab file path
	crontabFile := "/var/spool/cron/crontabs/" + crontabUser
	_, err := os.Stat(crontabFile)
	if err != nil {
		return "", err
	}
	return crontabFile, nil
}

// helper function to read crontab file
func readCrontabFile(crontabFile string) ([]string, error) {
	// Read the crontab file
	file, err := os.Open(crontabFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open crontab file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string

	// Iterate over each line in the crontab file
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	if scanner.Err() != nil {
		return nil, fmt.Errorf("error reading crontab file: %w", scanner.Err())
	}

	return lines, nil
}

// helper function to update a cron task
func updateCronTask(cronTask CronTask, lines []string) ([]string, error) {
	// Check if HeartbeatURL is set and is a correctly formatted URL
	if _, err := url.ParseRequestURI(cronTask.HeartbeatURL); err != nil {
		fmt.Printf("Invalid HeartbeatURL in task '%s', skipping this task.\n", cronTask.Task)
		return lines, nil
	}

	// Construct the curl command string to append to the task
	curlCommand := fmt.Sprintf(`curl -fs --retry 3 %s > /dev/null 2>&1`, cronTask.HeartbeatURL)

	// Construct the task spec string
	taskSpec := cronTask.Spec + " " + cronTask.Task

	var updatedLines []string

	// Check if the line matches the cron task
	for _, line := range lines {
		if strings.Contains(line, taskSpec) {
			// Append the curl command to the task
			line += " && " + curlCommand
		}

		updatedLines = append(updatedLines, line)
	}

	return updatedLines, nil
}

// helper function to write updated cron tasks back to the file
func writeCronTasksToFile(crontabFile string, updatedLines []string) error {
	// Write the updated lines back to the crontab file
	err := ioutil.WriteFile(crontabFile, []byte(strings.Join(updatedLines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated crontab file: %w", err)
	}

	return nil
}

// Function to parse crontab and prompt user for approval
func ParseAndApproveCronTasks(crontabUser string) ([]CronTask, error) {
	// Read the crontab file
	crontabFile, err := getCrontabFilePath(crontabUser)
	if err != nil {
		return nil, err
	}

	// Read the crontab file
	data, err := ioutil.ReadFile(crontabFile)
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
	
		// If cron task contains "uptime.betterstack", skip and inform user
		if strings.Contains(line, "uptime.betterstack") {
			fmt.Println("Skipping cron task containing 'uptime.betterstack.com':", line)
			continue
		}
	
		// Display the cron task and ask for approval
		fmt.Println("Cron task:", line)
		isApproved, exitLoop := promptApproval()
		if exitLoop {
			break
		}
		if !isApproved {
			continue
		}
	
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
func promptApproval() (bool, bool) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Continue? (n skips this cron, N skips the rest of the crons) (y/n/N): ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(strings.ToLower(text))

	switch text {
	case "y":
		return true, false
	case "n":
		return false, false
	case "N":
		return false, true
	default:
		return false, false
	}
}

// Function to append curl command to crontab tasks
func AppendCronsCommand(cronTasks []CronTask, crontabUser string) error {
	tempFile := "temp_crontab"

	// Dump current crontab into a temp file
	cmd := exec.Command("sh", "-c", fmt.Sprintf("crontab -l > %s", tempFile))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to dump current crontab: %w", err)
	}

	// Read the temp crontab file
	lines, err := readCrontabFile(tempFile)
	if err != nil {
		return err
	}

	// Loop through the cron tasks
	for _, cronTask := range cronTasks {
		// Update the cron task
		lines, err = updateCronTask(cronTask, lines)
		if err != nil {
			return err
		}
	}

	// Write the updated lines back to the temp file
	err = writeCronTasksToFile(tempFile, lines)
	if err != nil {
		return err
	}

	// Replace current user's crontab with temp file
	cmd = exec.Command("crontab", tempFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to replace crontab with temp file: %w", err)
	}

	// Cleanup: Delete the temporary file
	if err := os.Remove(tempFile); err != nil {
		return fmt.Errorf("failed to remove temporary file: %w", err)
	}

	fmt.Println("Cron tasks updated successfully.")
	return nil
}

// Function to copy a file (for creating backups)
func backupCrontabFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}
