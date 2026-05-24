package storage

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Publisher struct {
	FeedDir string
}

func NewPublisher(dataDir string) Publisher {
	return Publisher{
		FeedDir: filepath.Join(dataDir, "feeds"),
	}
}

func (publisher Publisher) Publish(filename string, write func(io.Writer) error) error {
	if err := os.MkdirAll(publisher.FeedDir, 0o755); err != nil {
		return fmt.Errorf("create feed dir: %w", err)
	}

	var output bytes.Buffer
	if err := write(&output); err != nil {
		return fmt.Errorf("generate feed: %w", err)
	}
	if err := validateXML(output.Bytes()); err != nil {
		return err
	}

	temp, err := os.CreateTemp(publisher.FeedDir, filename+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp feed: %w", err)
	}

	tempName := temp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tempName)
		}
	}()

	if _, err := temp.Write(output.Bytes()); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write temp feed: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temp feed: %w", err)
	}

	target := filepath.Join(publisher.FeedDir, filename)
	if err := os.Rename(tempName, target); err != nil {
		return fmt.Errorf("publish feed: %w", err)
	}

	committed = true
	return nil
}

func validateXML(data []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		if _, err := decoder.Token(); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("feed XML is not well-formed: %w", err)
		}
	}
}
