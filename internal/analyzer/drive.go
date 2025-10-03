package analyzer

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/demen1n/go-bytesize"
	"golang.org/x/sys/windows"
)

type Drive struct {
	Name  string
	Free  uint64
	Total uint64
}

func (d *Drive) SetInfo() error {
	var freeBytes, totalBytes uint64

	drivePtr, err := windows.UTF16PtrFromString(d.Name)
	if err != nil {
		return fmt.Errorf("failed to convert drive name: %w", err)
	}

	err = windows.GetDiskFreeSpaceEx(
		drivePtr,
		&freeBytes,
		&totalBytes,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to get disk space for %s: %w", d.Name, err)
	}

	d.Free = freeBytes
	d.Total = totalBytes
	return nil
}

func (d *Drive) String() string {
	bytesize.SetLocale(bytesize.LocaleRU)
	free := bytesize.New(float64(d.Free))
	total := bytesize.New(float64(d.Total))

	return fmt.Sprintf("Диск %s\\: свободно %s из %s", d.Name, free.String(), total.String())
}

func getDrives(path []string) []Drive {
	drives := make(map[string]bool)

	for _, folder := range path {
		folder = strings.TrimSpace(folder)
		if folder == "" {
			continue
		}

		drive := filepath.VolumeName(folder)
		if drive != "" {
			drive = strings.ToUpper(drive)
			drives[drive] = true
		}
	}

	var res []Drive
	for drive := range drives {
		res = append(res, Drive{Name: drive})
	}

	return res
}
