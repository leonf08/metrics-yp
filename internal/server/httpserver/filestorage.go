package httpserver

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/leonf08/metrics-yp.git/internal/storage"
)

type Saver struct {
	file    *os.File
	encoder *json.Encoder
	storage *storage.MemStorage
}

type Loader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewSaver(path string, st *storage.MemStorage) (*Saver, error) {
	if path == "" {
		return &Saver{}, nil
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Saver{
		file:    file,
		encoder: json.NewEncoder(file),
		storage: st,
	}, nil
}

func (s *Saver) SaveMetrics() error {
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

func (s *Saver) Close() error {
	return s.file.Close()
}

func NewLoader(path string) (*Loader, error) {
	if path == "" {
		return &Loader{}, nil
	}

	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Loader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (l *Loader) LoadMetrics() (*storage.MemStorage, error) {
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

func (l *Loader) Close() error {
	return l.file.Close()
}
