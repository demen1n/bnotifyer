package analyzer

import (
	"errors"
	"github.com/djherbis/times"
	"github.com/inhies/go-bytesize"
	"log"
	"os"
	"strings"
	"time"
)

const maxLength = 512

func readLogfile(filePath string) (string, error) {
	log.Println("Trying open " + filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := make([]byte, maxLength)

	stat, err := file.Stat()
	if err != nil {
		return "", err
	}

	start := stat.Size() - maxLength
	_, err = file.ReadAt(buf, start)
	if err != nil {
		return "", err
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

	fi, err := os.Stat(filepath)
	if err != nil {
		return "", errors.New("cannot read file stat")
	} else {
		size := bytesize.New(float64(fi.Size()))
		br.WriteString("\t\t" + size.String() + ", " + size.Format("%.0f", "byte", true) + "\n")
	}

	// проверяем время
	bt, ct, err := getFileTime(filepath)
	if err != nil {
		return "", errors.New("cannot read file time")
	}
	if bt != nil {
		br.WriteString("\t\tсоздано " + bt.Format("02.01.2006 15:04") + "\n")
	}
	if ct != nil {
		br.WriteString("\t\tизменено " + ct.Format("02.01.2006 15:04") + "\n")
	}

	return br.String(), nil
}
