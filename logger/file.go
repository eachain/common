package logger

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

const (
	defaultKeepDays = 7
)

type fileWriter struct {
	dir    string
	bakDir string
	file   string

	fp *atomic.Value // the output file fp
}

func (fw *fileWriter) Write(p []byte) (int, error) {
	return fw.fp.Load().(*os.File).Write(p)
}

// SetOutputFile 可以设置输出日志的目录，备份目录和日志文件名。
// 日志默认存储7天备份，每天凌晨切换。
func SetOutputFile(dir, bakDir, file string) error {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	fp, err := os.OpenFile(filepath.Join(dir, file), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	fw := &fileWriter{
		dir:    dir,
		bakDir: bakDir,
		file:   file,
		fp:     &atomic.Value{},
	}
	fw.fp.Store(fp)

	SetOutput(fw)

	go fw.refreshFile()
	go fw.removeOldFiles()
	return nil
}

func sleepUntilTommorrow() time.Time {
	now := time.Now()
	tommorrow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Add(24 * time.Hour)
	time.Sleep(tommorrow.Sub(now))
	return now
}

func bakFilename(file string, t time.Time) string {
	return file + "-" + t.Format("20060102")
}

func (fw *fileWriter) refreshFile() {
	for {
		yesterday := sleepUntilTommorrow()
		err := os.MkdirAll(fw.bakDir, 0755)
		if err != nil {
			Warnf("logger: make bak dir:", err)
			continue
		}

		oldpath := filepath.Join(fw.dir, fw.file)
		newpath := filepath.Join(fw.bakDir, bakFilename(fw.file, yesterday))

		err = os.Rename(oldpath, newpath)
		if err != nil {
			Warnf("logger: rename '%v' to '%v': %v", oldpath, newpath, err)
			continue
		}

		fp, err := os.OpenFile(oldpath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			Warnf("logger: create new log file: %v", err)
			continue
		}

		oldfp := fw.fp.Load().(*os.File)
		fw.fp.Store(fp)
		oldfp.Close()
	}
}

func (fw *fileWriter) removeOldFiles() {
	const Day = 24 * time.Hour
	b := make([]byte, 1)
	for {
		rand.Read(b)
		time.Sleep(time.Duration(b[0]%120) * time.Second)

		date := time.Now().Add(-defaultKeepDays * Day)
		for i := 0; i < 10; i++ {
			date = date.Add(-Day)
			os.Remove(filepath.Join(fw.bakDir, bakFilename(fw.file, date)))
		}
		sleepUntilTommorrow()
	}
}
