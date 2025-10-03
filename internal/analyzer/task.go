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

var (
	backupCount  int
	restoreCount int
)

type TextForSendError struct {
	Subject string
	Body    string
}

func Do(cfg *config.Config) {
	br := strings.Builder{}

	// инфа о жётских дисках
	br.WriteString("Диски:\n")
	var paths []string
	for _, el := range cfg.DB {
		paths = append(paths, el.Path)
	}

	drives := getDrives(paths)

	for _, drive := range drives {
		err := drive.SetInfo()
		if err != nil {
			log.Printf("drive.SetInfo() err: %v\n", err)
			return
		}

		br.WriteString(fmt.Sprintf("\t%s\n", drive.String()))
	}
	br.WriteString("\n")

	emailSender := exchangesmtp.NewQuickSender(cfg.EFS.Username(), cfg.EFS.Password, cfg.EFS.Server(), cfg.EFS.User, cfg.EFS.MailTo)

	wd := int(time.Now().Weekday())

	backupCount = 0
	restoreCount = 0

	dbChecked := 0
	for _, el := range cfg.DB {
		if el.Weekday >= 0 && el.Weekday != wd {
			continue
		}

		log.Println(el.Name)

		te, err := byDB(el.Name, el.File, el.Path, cfg.ST, &br)
		if te != nil {
			err = emailSender.Send(te.Subject, te.Body)
			if err != nil {
				log.Println("Error sending email: ", err)
				log.Println(fmt.Sprintf("Email %s:\n%s", te.Subject, te.Body))
			}
		}

		if err != nil {
			log.Println("Для БД " + el.Name + ":\n" + err.Error())
		}

		dbChecked++
	}

	if br.Len() > 0 {
		sbj := fmt.Sprintf("бэкапы баз данных (БД %d: бэкапов %d, ресторов %d)", dbChecked, backupCount, restoreCount)
		err := emailSender.Send(sbj, br.String())
		if err != nil {
			log.Println("Error sending email: ", err)
			log.Println(fmt.Sprintf("Email %s:\n%s", sbj, br.String()))
		}
	}
}

func byDB(nameDB, file, path string, st config.SearchedText, br *strings.Builder) (*TextForSendError, error) {
	// проверка на существование fdb файла
	fdbFile := filepath.Join(path, file+".fdb")
	if _, err := os.Stat(fdbFile); errors.Is(err, os.ErrNotExist) {
		return &TextForSendError{
			Subject: fmt.Sprintf("Нет файла бэкапа БД %s", nameDB),
			Body:    fmt.Sprintf("Не найден файл %s", fdbFile),
		}, err
	}

	// анализ бэкапа
	backupText, err := readLogfile(filepath.Join(path, backupLog))
	if err != nil {
		subject := fmt.Sprintf("Ошибка чтения лога бэкапа %s", nameDB)
		body := fmt.Sprintf("Ошибка: %s", err)
		if os.IsNotExist(err) {
			body = "Нет файла лога бэкапа"
		}
		return &TextForSendError{
			subject,
			body,
		}, err
	}

	br.WriteString("БД " + nameDB + "\n")

	// бэкап завершился с ошибкой
	if !strings.Contains(backupText, st.Backup) {
		log.Println("Backup fail for " + nameDB)
		br.WriteString("\tОшибка бэкапа!\n")

		return &TextForSendError{
			"Ошибка при бэкапе " + nameDB,
			"Конец лога бэкапа базы " + nameDB + ":\n" + backupText,
		}, nil
	}

	// бэкап прошёл успешно
	log.Println("Backup succeed for " + nameDB)
	br.WriteString("\tУспешный бэкап\n")
	fs, err := collectFileStat(filepath.Join(path, file+".fbk"))
	if err != nil {
		log.Printf("Get files stat for %s error: %v\n", file, err)
	}
	br.WriteString(fs)
	backupCount += 1

	// анализ рестора
	restoreText, err := readLogfile(filepath.Join(path, restoreLog))
	if err != nil {
		subject := fmt.Sprintf("Ошибка чтения лога рестора %s", nameDB)
		body := fmt.Sprintf("Ошибка: %s", err)
		if os.IsNotExist(err) {
			body = "Нет файла лога рестора"
		}
		return &TextForSendError{
			subject,
			body,
		}, err
	}

	if !strings.Contains(restoreText, st.Restore) {
		log.Println("Restore fail for " + nameDB)
		br.WriteString("\tОшибка рестора!\n")
		return &TextForSendError{
			"Ошибка при ресторе " + nameDB,
			"Конец лога рестора базы " + nameDB + ":\n" + restoreText,
		}, nil
	}

	log.Println("Restore succeed for " + nameDB)
	br.WriteString("\tУспешный рестор\n")
	fs, err = collectFileStat(fdbFile)
	if err != nil {
		log.Printf("Get files stat for %s error: %v\n", file, err)
	}
	br.WriteString(fs)
	restoreCount += 1

	br.WriteString("\n\n")

	return nil, nil
}
