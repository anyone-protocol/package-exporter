package fetcher

import (
	"bufio"
	"context"
	"io"
	"os"
	"regexp"

	"github.com/ATOR-Development/downloads-exporter/internal/counter"
)

// NewNginxAccessLogFetcher creates new Nginx access log downloads fetcher from config
func NewNginxAccessLogFetcher(name, accessLogPath string, accessLogRegexp *regexp.Regexp, labels map[string]*regexp.Regexp, counter *counter.Counter) Fetcher {
	return &nginxAccessLogFetcher{
		name:            name,
		accessLogPath:   accessLogPath,
		accessLogRegexp: accessLogRegexp,
		labels:          labels,
		counter:         counter,

		logFile:   nil,
		logReader: nil,
	}
}

type nginxAccessLogFetcher struct {
	name            string
	accessLogPath   string
	accessLogRegexp *regexp.Regexp
	labels          map[string]*regexp.Regexp
	counter         *counter.Counter

	logFile   *os.File
	logReader *bufio.Reader
}

// FetchCount fetches download count from nginx access logs and returns it
func (f *nginxAccessLogFetcher) FetchCount(ctx context.Context) ([]*counter.Result, error) {
	err := f.reopenLogFileIfNeeded()
	if err != nil {
		return nil, err
	}

	for {
		line, err := f.logReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return f.counter.Results(), nil
			}

			return nil, err
		}

		if f.accessLogRegexp.MatchString(line) {
			labels := make(map[string]string)
			for labelName, labelRegexp := range f.labels {
				submatch := labelRegexp.FindStringSubmatch(line)
				if len(submatch) >= 2 {
					labels[labelName] = submatch[1]
				}
			}

			f.counter.Increment(labels)
		}
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
