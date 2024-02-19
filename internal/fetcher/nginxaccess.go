package fetcher

import (
	"bufio"
	"context"
	"io"
	"os"
	"regexp"
)

// NewNginxAccessLogFetcher creates new Nginx access log downloads fetcher from config
func NewNginxAccessLogFetcher(name, accessLogPath string, accessLogRegexp *regexp.Regexp) Fetcher {
	return &nginxAccessLogFetcher{
		name:            name,
		accessLogPath:   accessLogPath,
		accessLogRegexp: accessLogRegexp,

		logFile:         nil,
		logReader:       nil,
		downloadCounter: 0,
	}
}

type nginxAccessLogFetcher struct {
	name            string
	accessLogPath   string
	accessLogRegexp *regexp.Regexp

	logFile         *os.File
	logReader       *bufio.Reader
	downloadCounter int
}

// FetchCount fetches download count from nginx access logs and returns it
func (f *nginxAccessLogFetcher) FetchCount(ctx context.Context) (int, error) {
	err := f.reopenLogFileIfNeeded()
	if err != nil {
		return 0, err
	}

	for {
		_, err := f.logReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return f.downloadCounter, nil
			}

			return 0, err
		}

		f.downloadCounter++
	}
}

func (f *nginxAccessLogFetcher) Name() string {
	return f.name
}

// reopenLogFileIfNeeded reopens log file if it was truncated or removed
func (f *nginxAccessLogFetcher) reopenLogFileIfNeeded() error {
	if f.logFile == nil {
		return f.openLogFile(0)
	}

	fileToCheck, err := os.Open(f.accessLogPath)
	if err != nil {
		return err
	}

	defer fileToCheck.Close()

	currentPos, err := f.logFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	fileInfo, err := fileToCheck.Stat()
	if err != nil {
		return err
	}

	// not truncated, early return
	if currentPos <= fileInfo.Size() {
		return nil
	}

	// close previous file handle
	err = f.logFile.Close()
	if err != nil {
		return err
	}

	// open new file handle at the end
	err = f.openLogFile(fileInfo.Size())
	if err != nil {
		return err
	}

	return nil
}

// openLogFile opens log file at specified offset
func (f *nginxAccessLogFetcher) openLogFile(offset int64) error {
	var err error
	f.logFile, err = os.Open(f.accessLogPath)
	if err != nil {
		return err
	}

	if offset > 0 {
		_, err = f.logFile.Seek(offset, io.SeekStart)
		if err != nil {
			return err
		}
	}

	f.logReader = bufio.NewReader(f.logFile)

	return nil
}
