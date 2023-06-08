package main

import (
	"github.com/IT-JONCTION/beatify/cli"
)

func main() {
	cli.HandleCommandLineOptions()



	// // Read the crontab schedule and heartbeat name from command-line arguments
	// crontab := os.Args[1]
	// heartbeatName := os.Args[2]

	// // Create a new cron parser
	// parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	// // Parse the crontab schedule
	// schedule, err := parser.Parse(crontab)
	// if err != nil {
	// 	fmt.Println("Error parsing crontab schedule:", err)
	// 	return
	// }

	// // Calculate the period in seconds
	// period := int(schedule.Next(time.Now()).Sub(time.Now()).Seconds())

	// // Calculate the grace period as approximately 20% of the period
	// grace := int(float64(period) * 0.2)

	// // Create the JSON representation
	// jsonData := fmt.Sprintf(`{
	// 	"name": "%s",
	// 	"period": %d,
	// 	"grace": %d
	// }`, heartbeatName, period, grace)

	// fmt.Println(jsonData)
}
