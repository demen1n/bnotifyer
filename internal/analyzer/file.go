package analyzer

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/demen1n/go-bytesize"
	"github.com/djherbis/times"
)

const maxLength = 512

func readLogfile(filePath string) (string, error) {
	log.Println("Trying open " + filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("Error closing file %s: %v\n", filePath, cerr)
		}
	}()

	stat, err := file.Stat()
	if err != nil {
		return "", err
	}

	fileSize := stat.Size()
	if fileSize == 0 {
		return "", errors.New("file is empty")
	}

	var start int64
	var readSize int

	if fileSize < maxLength {
		start = 0
		readSize = int(fileSize)
	} else {
		start = fileSize - maxLength
		readSize = maxLength
	}

	buf := make([]byte, readSize)
	_, err = file.ReadAt(buf, start)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(buf), nil
}

// getFileTime возвращает таймштампы создания файла и модификации файла
func getFileTime(path string) (bt *time.Time, ct *time.Time, err error) {
	t, err := times.Stat(path)
	if err != nil {
		return
	}

	if t.HasBirthTime() {
		b := t.BirthTime()
		bt = &b
	}

	mt := t.ModTime()
	ct = &mt

	return
}

func collectFileStat(filepath string) (string, error) {
	br := strings.Builder{}
	bytesize.SetLocale(bytesize.LocaleRU)

	fi, err := os.Stat(filepath)
	if err != nil {
		return "", fmt.Errorf("cannot read file stat for %s: %w", filepath, err)
	}

	size := bytesize.New(float64(fi.Size()))
	br.WriteString("\t\t" + size.String() + ", " + size.Format("%.0f ", "byte", true) + "\n")

	bt, ct, err := getFileTime(filepath)
	if err != nil {
		return "", fmt.Errorf("cannot read file time for %s: %w", filepath, err)
	}

	if bt != nil {
		br.WriteString("\t\tсоздано " + bt.Format("02.01.2006 15:04") + "\n")
	}
	if ct != nil {
		br.WriteString("\t\tизменено " + ct.Format("02.01.2006 15:04") + "\n")
	}

	return br.String(), nil
}
