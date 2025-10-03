package analyzer

import (
	"bnotifyer/internal/config"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/demen1n/exchange-smtp"
)

const (
	backupLog  = "log_backup.txt"
	restoreLog = "log_restore.txt"
)

type Stats struct {
	BackupCount  int
	RestoreCount int
	DBChecked    int
}

type CheckResult struct {
	Error     error
	NeedEmail bool
	Subject   string
	Body      string
}

func Do(cfg *config.Config) {
	br := strings.Builder{}
	stats := &Stats{}

	// информация о дисках
	br.WriteString("Диски:\n")
	collectDriveInfo(cfg, &br)
	br.WriteString("\n")

	emailSender := exchangesmtp.NewQuickSender(
		cfg.EFS.Username(),
		cfg.EFS.Password,
		cfg.EFS.Server(),
		cfg.EFS.User,
		cfg.EFS.MailTo,
	)

	wd := int(time.Now().Weekday())

	// проверка всех баз данных
	for _, db := range cfg.DB {
		if db.Weekday >= 0 && db.Weekday != wd {
			continue
		}

		log.Println("Checking DB:", db.Name)
		result := checkDB(db.Name, db.File, db.Path, cfg.ST, &br, stats)

		// отправляем email если нужно
		if result.NeedEmail {
			if err := emailSender.Send(result.Subject, result.Body); err != nil {
				log.Printf("Error sending email: %v\n", err)
				log.Printf("Email content - Subject: %s\nBody: %s\n", result.Subject, result.Body)
			}
		}

		if result.Error != nil {
			log.Printf("Error checking DB %s: %v\n", db.Name, result.Error)
		}

		stats.DBChecked++
	}

	// отправляем итоговый отчет
	sendSummaryEmail(emailSender, stats, &br)
}

func collectDriveInfo(cfg *config.Config, br *strings.Builder) {
	var paths []string
	for _, el := range cfg.DB {
		paths = append(paths, el.Path)
	}

	drives := getDrives(paths)

	for _, drive := range drives {
		if err := drive.SetInfo(); err != nil {
			log.Printf("Warning: failed to get info for drive %s: %v\n", drive.Name, err)
			br.WriteString(fmt.Sprintf("\t%s: ошибка получения информации\n", drive.Name))
			continue
		}
		br.WriteString(fmt.Sprintf("\t%s\n", drive.String()))
	}
}

func sendSummaryEmail(sender *exchangesmtp.QuickSender, stats *Stats, br *strings.Builder) {
	if br.Len() > 0 {
		subject := fmt.Sprintf(
			"бэкапы баз данных (БД %d: бэкапов %d, ресторов %d)",
			stats.DBChecked,
			stats.BackupCount,
			stats.RestoreCount,
		)

		if err := sender.Send(subject, br.String()); err != nil {
			log.Printf("Error sending summary email: %v\n", err)
			log.Printf("Summary email - Subject: %s\nBody: %s\n", subject, br.String())
		}
	}
}

func checkDB(nameDB, file, path string, st config.SearchedText, br *strings.Builder, stats *Stats) CheckResult {
	result := CheckResult{}

	// проверка существования fdb файла
	fdbFile := filepath.Join(path, file+".fdb")
	if _, err := os.Stat(fdbFile); errors.Is(err, os.ErrNotExist) {
		result.NeedEmail = true
		result.Subject = fmt.Sprintf("Нет файла бэкапа БД %s", nameDB)
		result.Body = fmt.Sprintf("Не найден файл %s", fdbFile)
		result.Error = err
		return result
	}

	br.WriteString("БД " + nameDB + "\n")

	// проверка бэкапа
	backupResult := checkBackup(nameDB, file, path, st.Backup, br)
	if backupResult.NeedEmail {
		return backupResult
	}
	if backupResult.Error == nil {
		stats.BackupCount++
	}

	// проверка рестора
	restoreResult := checkRestore(nameDB, file, path, st.Restore, br)
	if restoreResult.NeedEmail {
		return restoreResult
	}
	if restoreResult.Error == nil {
		stats.RestoreCount++
	}

	br.WriteString("\n\n")
	return result
}

func checkBackup(nameDB, file, path, searchText string, br *strings.Builder) CheckResult {
	result := CheckResult{}

	backupText, err := readLogfile(filepath.Join(path, backupLog))
	if err != nil {
		result.NeedEmail = true
		result.Subject = fmt.Sprintf("Ошибка чтения лога бэкапа %s", nameDB)
		result.Body = fmt.Sprintf("Ошибка: %s", err)
		if os.IsNotExist(err) {
			result.Body = "Нет файла лога бэкапа"
		}
		result.Error = err
		return result
	}

	if !strings.Contains(backupText, searchText) {
		log.Println("Backup failed for " + nameDB)
		br.WriteString("\tОшибка бэкапа!\n")
		result.NeedEmail = true
		result.Subject = "Ошибка при бэкапе " + nameDB
		result.Body = "Конец лога бэкапа базы " + nameDB + ":\n" + backupText
		return result
	}

	log.Println("Backup succeeded for " + nameDB)
	br.WriteString("\tУспешный бэкап\n")

	if fs, err := collectFileStat(filepath.Join(path, file+".fbk")); err != nil {
		log.Printf("Get file stat for %s error: %v\n", file, err)
	} else {
		br.WriteString(fs)
	}

	return result
}

func checkRestore(nameDB, file, path, searchText string, br *strings.Builder) CheckResult {
	result := CheckResult{}

	restoreText, err := readLogfile(filepath.Join(path, restoreLog))
	if err != nil {
		result.NeedEmail = true
		result.Subject = fmt.Sprintf("Ошибка чтения лога рестора %s", nameDB)
		result.Body = fmt.Sprintf("Ошибка: %s", err)
		if os.IsNotExist(err) {
			result.Body = "Нет файла лога рестора"
		}
		result.Error = err
		return result
	}

	if !strings.Contains(restoreText, searchText) {
		log.Println("Restore failed for " + nameDB)
		br.WriteString("\tОшибка рестора!\n")
		result.NeedEmail = true
		result.Subject = "Ошибка при ресторе " + nameDB
		result.Body = "Конец лога рестора базы " + nameDB + ":\n" + restoreText
		return result
	}

	log.Println("Restore succeeded for " + nameDB)
	br.WriteString("\tУспешный рестор\n")

	fdbFile := filepath.Join(path, file+".fdb")
	if fs, err := collectFileStat(fdbFile); err != nil {
		log.Printf("Get file stat for %s error: %v\n", file, err)
	} else {
		br.WriteString(fs)
	}

	return result
}
