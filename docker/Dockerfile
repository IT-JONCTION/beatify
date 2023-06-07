FROM golang:latest

# Install cron and other necessary dependencies
RUN apt-get update && apt-get install -y cron

# Set the working directory inside the container
WORKDIR /app

# Copy your Go application into the container
COPY . .

# Copy the crontab file into the container
COPY crontab /var/spool/cron/crontabs/root

# Give appropriate permissions to the crontab file
RUN chown root:crontab /var/spool/cron/crontabs/root
RUN chmod 600 /var/spool/cron/crontabs/root

# Create the log file for cron
RUN touch /var/log/cron.log

# Run cron in the foreground
CMD cron && tail -f /var/log/cron.log
