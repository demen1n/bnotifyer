# bnotifyer

A monitoring tool for Firebird database backup and restore operations with email notifications.

## Features

- Monitors backup (.fbk) and restore (.fdb) operations by analyzing log files
- Checks disk space availability
- Sends email notifications for successful/failed operations
- Supports multiple databases with individual scheduling
- Configurable weekday-based checks
- Russian locale support for file sizes and messages

## Requirements

- Go 1.16+
- Windows OS (uses Windows-specific APIs for disk info)
- Access to Exchange/SMTP server for email notifications

## Installation

```bash
go get github.com/yourusername/bnotifyer
go build ./cmd/bnotifyer
```

## Configuration

Create a `config.yml` file:

```yaml
searched_text:
  text_backup: "finishing, closing, and going home"
  text_restore: "finishing, closing, and going home"

email_for_send:
  server_url: "mail.example.com"
  server_port: 587
  user: "monitor@example.com"
  password: "your-password"
  mail_to:
    - "admin@example.com"

db:
  - name: "Production DB"
    path: "D:\\Backups\\ProdDB"
    file: "proddb"
    weekday: 1  # 0=Sunday, 1=Monday, etc. Use -1 for daily checks
  
  - name: "Test DB"
    path: "D:\\Backups\\TestDB"
    file: "testdb"
    weekday: -1
```

## Usage

```bash
./bnotifyer
```

The application will:
1. Read the configuration file
2. Check each configured database based on weekday settings
3. Analyze `log_backup.txt` and `log_restore.txt` in each database path
4. Verify the presence of `.fbk` and `.fdb` files
5. Send email notifications for errors or a summary report

## Log Files

The application creates `info.log` in the working directory with execution details.

## License

MIT