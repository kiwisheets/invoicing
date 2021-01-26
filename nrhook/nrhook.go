package nrhook

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"os"

	"github.com/sethgrid/pester"
	"github.com/sirupsen/logrus"
)

const (
	endpoint = "https://log-api.eu.newrelic.com/log/v1"
)

type NrHook struct {
	client      *pester.Client
	application string
	licenseKey  string
}

func NewNrHook(appName string, license string) *NrHook {
	nrHook := &NrHook{
		client:      pester.New(),
		application: appName,
		licenseKey:  license,
	}

	return nrHook
}

func (h *NrHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		fmt.Fprintf(os.Stderr, "NrHook failed to fire. Unable to read entry, %v", err)
		return err
	}

	// fmt.Fprintf(os.Stderr, "Line: %v", line)

	// fire and forget
	go func(line string) {
		var buffer bytes.Buffer
		writer := gzip.NewWriter(&buffer)
		if _, err := writer.Write([]byte(line)); err != nil {
			fmt.Fprintf(os.Stderr, "failed to gzip message: %v", err)
		}
		if err := writer.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "error flushing gzip writer, %v", err)
		}
		if err := writer.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error flushing gzip writer, %v", err)
		}

		request, err := http.NewRequest("POST", endpoint, &buffer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating log request to NR: %v", err)
		}

		request.Header.Add("Content-Type", "application/gzip")
		request.Header.Add("Content-Encoding", "gzip")
		request.Header.Add("Accept", "*/*")
		request.Header.Add("X-License-Key", h.licenseKey)

		res, err := h.client.Do(request)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error sending log request to NR: %v", err)
		}
		if res != nil {
			fmt.Fprintf(os.Stderr, "nrhook status code: %v", res.Status)
		}

	}(line)

	return nil
}

func (h *NrHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
