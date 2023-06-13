package crontab

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type CronTask struct {
	Spec         string
	Task         string
	Name         string
	HeartbeatURL string
}

// Constants for temp and backup file prefixes
const (
	TempFilePrefix   = "crontab"
	BackupFilePrefix = "crontab_backup"
)

// Package-level variables for the temp and backup files
var (
	TempFile   *os.File
	BackupFile *os.File
)

func IsValidUsername(username string) error {
	r := regexp.MustCompile(`^[a-zA-Z0-9_-]{3,16}$`)
	if !r.MatchString(username) {
		return fmt.Errorf("invalid username '%s': a valid username is 3-16 characters long and consists of alphanumeric characters, underscores, or dashes", username)
	}
	return nil
}

// Helper function to read crontab file
func readCrontabFile() ([]string, error) {
	// Ensure TempFile is not nil
	if TempFile == nil {
		return nil, fmt.Errorf("temp file has not been created yet")
	}

	// Open the temp crontab file
	file, err := os.Open(TempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open temp crontab file: %w", err)
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
		return nil, fmt.Errorf("error reading temp crontab file: %w", scanner.Err())
	}

	return lines, nil
}

// helper function to update a cron task
func updateCronTask(cronTask CronTask, lines []string) ([]string, error) {
	// Check if Spec, Task, and Name fields of the cronTask are not empty
	if cronTask.Spec == "" || cronTask.Task == "" || cronTask.Name == "" {
		return nil, fmt.Errorf("invalid CronTask, 'Spec', 'Task' or 'Name' field is empty")
	}

	// Check if HeartbeatURL is set and is a correctly formatted URL
	if _, err := url.ParseRequestURI(cronTask.HeartbeatURL); err != nil {
		return nil, fmt.Errorf("invalid HeartbeatURL in task '%s': %w", cronTask.Task, err)
	}

	// Construct the curl command string to append to the task
	curlCommand := fmt.Sprintf(`curl -fs --retry 3 %s > /dev/null 2>&1`, cronTask.HeartbeatURL)

	// Construct the task spec string
	taskSpec := cronTask.Spec + " " + cronTask.Task

	var updatedLines []string

	// Check if the line matches the cron task
	for _, line := range lines {
		if strings.Contains(line, taskSpec) {
			// Check if the curl command is already appended to the task
			if !strings.Contains(line, curlCommand) {
				// Append the curl command to the task
				line += " && " + curlCommand
			} else {
				return nil, fmt.Errorf("curl command is already appended to task '%s'", cronTask.Task)
			}
		}

		updatedLines = append(updatedLines, line)
	}

	return updatedLines, nil
}

// helper function to write updated cron tasks back to the file
func writeCronTasksToFile(crontabFile string, updatedLines []string) error {
	// Join the updated lines with newline characters
	fileContent := strings.Join(updatedLines, "\n") + "\n"

	// Write the updated lines back to the crontab file
	err := ioutil.WriteFile(crontabFile, []byte(fileContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated crontab file: %w", err)
	}

	return nil
}

// Function to parse crontab and prompt user for approval
func ParseAndApproveCronTasks(crontabUser string) ([]CronTask, error) {

	// Prepare the temp and backup crontab files
	err := PrepareCrontabFiles(crontabUser)
	if err != nil {
		return nil, err
	}

	// Read the temp crontab file
	lines, err := readCrontabFile()
	if err != nil {
		return nil, err
	}

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
		isApproved, exitLoop, err := promptApproval()
		if err != nil {
			return nil, fmt.Errorf("failed to get approval: %w", err)
		}
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
		return "", fmt.Errorf("failed to read input for heartbeat name: %w", err)
	}

	name = strings.TrimSpace(name)

	// Perform basic validation to prevent command injection
	if strings.ContainsAny(name, `";$|><&`) {
		return "", fmt.Errorf("heartbeat name contains invalid characters: %s", name)
	}

	return name, nil
}

func promptApproval() (bool, bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Continue? (n skips this cron, N skips the rest of the crons) (y/n/N): ")
	text, err := reader.ReadString('\n')
	if err != nil {
		return false, false, fmt.Errorf("error reading input: %w", err)
	}
	text = strings.TrimSpace(text)

	switch text {
	case "y":
		return true, false, nil
	case "n":
		return false, false, nil
	case "N":
		return false, true, nil
	default:
		return false, false, nil
	}
}

// Lower-level utility function to dump the crontab to a specified file
func DumpCrontabToFile(crontabUser, fileType string) error {
	if err := IsValidUsername(crontabUser); err != nil {
		return err
	}

	var err error
	var filePath string

	// Determine file type and create file accordingly
	switch fileType {
	case "temp":
		if TempFile == nil {
			TempFile, err = ioutil.TempFile("", TempFilePrefix)
			if err != nil {
				return fmt.Errorf("failed to create temp file: %w", err)
			}
		}
		filePath = TempFile.Name()
	case "backup":
		// Use the user's home directory to store backup file
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		backupFilePath := fmt.Sprintf("%s/%s.bak", homeDir, BackupFilePrefix)
		BackupFile, err = os.OpenFile(backupFilePath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return fmt.Errorf("failed to open or create backup file: %w", err)
		}
		filePath = BackupFile.Name()
	default:
		return fmt.Errorf("invalid file type: %s", fileType)
	}

	// Get the crontab file path
	crontabFilePath := ""
	if crontabUser != "" {
		// User-specific crontab file
		crontabFilePath = fmt.Sprintf("/var/spool/cron/crontabs/%s", crontabUser)
	} else {
		// Current user's crontab file
		crontabFilePath = fmt.Sprintf("/var/spool/cron/crontabs/%s", os.Getenv("USER"))
	}

	// Read the crontab file
	bytes, err := ioutil.ReadFile(crontabFilePath)
	if err != nil {
		return fmt.Errorf("failed to read crontab file: %w", err)
	}

	// Write the crontab content to the file
	err = ioutil.WriteFile(filePath, bytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write crontab to file: %w", err)
	}

	return nil
}

// Higher-level utility function to prepare the temporary and backup files
func PrepareCrontabFiles(crontabUser string) error {
	if err := IsValidUsername(crontabUser); err != nil {
		return err
	}
	// Dump crontab to temp file
	if err := DumpCrontabToFile(crontabUser, "temp"); err != nil {
		return err
	}

	// Backup crontab
	if err := DumpCrontabToFile(crontabUser, "backup"); err != nil {
		return err
	}

	return nil
}

// Function to append curl command to crontab tasks
func AppendCronsCommand(cronTasks []CronTask, crontabUser string) error {

	if err := IsValidUsername(crontabUser); err != nil {
		return err
	}

	// Prepare the temporary and backup crontab files
	if err := PrepareCrontabFiles(crontabUser); err != nil {
		return err
	}

	// Read the temp crontab file
	lines, err := readCrontabFile()
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
	err = writeCronTasksToFile(TempFile.Name(), lines)
	if err != nil {
		return err
	}

	err = reloadCrontab(TempFile.Name(), crontabUser)
	if err != nil {
		return err
	}
	return nil
}

// Function to load reload crontab from a file
func reloadCrontab(fileName, crontabUser string) error {
	if err := IsValidUsername(crontabUser); err != nil {
		return err
	}
	var cmd *exec.Cmd
	// Check if crontabUser is supplied
	if crontabUser != "" {
		// Load crontab from specified file for specific user
		cmd = exec.Command("sh", "-c", fmt.Sprintf("crontab -u %s %s", crontabUser, fileName))
	} else {
		// Load crontab from specified file
		cmd = exec.Command("sh", "-c", fmt.Sprintf("crontab %s", fileName))
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to load crontab from %s: %w\nOutput: %s", fileName, err, string(output))
	}

	// Cleanup: Delete the temporary file
	if err := os.Remove(fileName); err != nil {
		return fmt.Errorf("failed to remove temporary file: %w", err)
	}

	fmt.Println("End.")

	return nil
}
