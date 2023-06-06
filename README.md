# Beatify

Beatify is a command-line tool that automates the creation of heartbeats for monitoring cron tasks using the BetterUptime service.

## Usage
beatify [OPTIONS]

## Description

Beatify reads the user's crontab, presents each cron task for approval to create a heartbeat, calls the BetterUptime API to create the approved heartbeats, and updates the crontab to append a curl request to each approved cron task.

## Options

- `-a, --auth-token AUTH_TOKEN`: Optional. The authentication token for the BetterUptime API. If not provided, the tool will prompt for it during runtime.
- `-u, --user USER`: Optional. The crontab user to edit. If not provided, the tool will default to the current user's crontab.
- `-h, --help`: Display the help message and exit.

## Examples

To run Beatify and create heartbeats for cron tasks:
beatify -a <YOUR_AUTH_TOKEN> -u www-data


## Exit Status

0 if successful, or an error code if an error occurs.

## Reporting Bugs

Report bugs to the GitHub repository: [https://github.com/IT-JONCTION/beatify](https://github.com/IT-JONCTION/beatify)

## Author

Your Name <wayne@it-jonction-lab.com>

## Copyright

Copyright © 2023 IT Jonction Lab. This is free software; see the source code for copying conditions. There is NO warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

## See Also

The BetterUptime API documentation: [https://docs.betteruptime.com/api/](https://docs.betteruptime.com/api/)




