package analyzer

import (
	"bnotifyer/internal/config"
	"fmt"
	"github.com/demen1n/exchange-smtp"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	backupLog  = "log_backup.txt"
	restoreLog = "log_restore.txt"
)

var (
	backupCount  int
	restoreCount int
)

type TextForSend struct {
	Subject string
	Body    string
}

func Do(cfg *config.Config) {
	tts := collectInfo(cfg.DB, cfg.ST)
	emailSender := exchangesmtp.NewQuickSender(cfg.EFS.Username(), cfg.EFS.Password, cfg.EFS.Server(), cfg.EFS.User, cfg.EFS.MailTo)
	var err error

	for _, t := range tts {
		err = emailSender.Send(t.Subject, t.Body)
		if err != nil {
			log.Println("Error sending email: ", err)
			log.Println(fmt.Sprintf("Email %s:\n%s", t.Subject, t.Body))
		}
	}
}

func collectInfo(db []config.Database, st config.SearchedText) []TextForSend {
	br := strings.Builder{}
	tts := make([]TextForSend, 0)

	wd := int(time.Now().Weekday())

	backupCount = 0
	restoreCount = 0
	dbChecked := 0
	for _, el := range db {
		if el.Weekday >= 0 && el.Weekday != wd {
			continue
		}

		log.Println(el.Name)

		te, err := byDB(el.Name, el.File, el.Path, st, &br)
		if te != nil {
			tts = append(tts, *te)
		}

		if err != nil {
			log.Println("Для БД " + el.Name + ":\n" + err.Error())
		}

		dbChecked++
	}

	if br.Len() > 0 {
		sbj := fmt.Sprintf("бэкапы баз данных (БД %d: бэкапов %d, ресторов %d)", dbChecked, backupCount, restoreCount)
		tts = append(tts, TextForSend{Subject: sbj, Body: br.String()})
	}

	return tts
}

func byDB(nameDB, file, path string, st config.SearchedText, br *strings.Builder) (*TextForSend, error) {
	// анализ бэкапа
	backupText, err := readLogfile(filepath.Join(path, backupLog))
	if err != nil {
		subject := fmt.Sprintf("Ошибка чтения лога бэкапа %s", nameDB)
		body := fmt.Sprintf("Ошибка: %s", err)
		if os.IsNotExist(err) {
			body = "Нет файла лога бэкапа"
		}
		return &TextForSend{
			subject,
			body,
		}, err
	}

	br.WriteString("БД " + nameDB + "\n")

	// бэкап завершился с ошибкой
	if !strings.Contains(backupText, st.Backup) {
		log.Println("Backup fail for " + nameDB)
		br.WriteString("\tОшибка бэкапа!\n")

		return &TextForSend{
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
		return &TextForSend{
			subject,
			body,
		}, err
	}

	if !strings.Contains(restoreText, st.Restore) {
		log.Println("Restore fail for " + nameDB)
		br.WriteString("\tОшибка рестора!\n")
		return &TextForSend{
			"Ошибка при ресторе " + nameDB,
			"Конец лога рестора базы " + nameDB + ":\n" + restoreText,
		}, nil
	}

	log.Println("Restore succeed for " + nameDB)
	br.WriteString("\tУспешный рестор\n")
	fs, err = collectFileStat(filepath.Join(path, file+".fdb"))
	if err != nil {
		log.Printf("Get files stat for %s error: %v\n", file, err)
	}
	br.WriteString(fs)
	restoreCount += 1

	br.WriteString("\n\n")

	return nil, nil
}
