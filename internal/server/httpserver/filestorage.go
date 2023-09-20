package httpserver

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/leonf08/metrics-yp.git/internal/storage"
)

type saver struct {
	file    *os.File
	encoder *json.Encoder
	storage Repository
}

type loader struct {
	file    *os.File
	decoder *json.Decoder
}

func newSaver(path string, st Repository) (*saver, error) {
	if path == "" {
		return &saver{}, nil
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		return nil, err
	}

	return &saver{
		file:    file,
		encoder: json.NewEncoder(file),
		storage: st,
	}, nil
}

func (s *saver) saveMetrics() error {
	err := s.file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = s.file.Seek(0, 0)
	if err != nil {
		return err
	}

	s.encoder.SetIndent("", "    ")

	return s.encoder.Encode(s.storage)
}

func (s *saver) close() error {
	return s.file.Close()
}

func newLoader(path string) (*loader, error) {
	if path == "" {
		return &loader{}, nil
	}

	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0o666)
	if err != nil {
		return nil, err
	}

	return &loader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (l *loader) loadMetrics() (*storage.MemStorage, error) {
	m := storage.NewStorage()

	info, err := l.file.Stat()
	if err != nil {
		return nil, err
	}

	if info.Size() > 0 {
		if err := l.decoder.Decode(m); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (l *loader) close() error {
	return l.file.Close()
}
