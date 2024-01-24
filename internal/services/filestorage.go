package services

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/leonf08/metrics-yp.git/internal/services/repo"
)

type (
	FileStorage struct {
		s *saver
		l *loader
	}

	saver struct {
		file    *os.File
		encoder *json.Encoder
	}

	loader struct {
		file    *os.File
		decoder *json.Decoder
	}
)

func NewFileStorage(path string) (*FileStorage, error) {
	s, err := newSaver(path)
	if err != nil {
		return nil, err
	}

	l, err := newLoader(path)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		s: s,
		l: l,
	}, nil
}

func newSaver(path string) (*saver, error) {
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
	}, nil
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

func (fs *FileStorage) Save(r Repository) error {
	m, ok := r.(*repo.MemStorage)
	if !ok {
		return errors.New("invalid type assertion for in-memory storage")
	}

	err := fs.s.file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = fs.s.file.Seek(0, 0)
	if err != nil {
		return err
	}

	fs.s.encoder.SetIndent("", "    ")

	return fs.s.encoder.Encode(&m.Storage)
}

func (fs *FileStorage) Load(r Repository) error {
	m, ok := r.(*repo.MemStorage)
	if !ok {
		return errors.New("invalid type assertion for in-memory storage")
	}

	info, err := fs.l.file.Stat()
	if err != nil {
		return err
	}

	if info.Size() > 0 {
		if err := fs.l.decoder.Decode(&m.Storage); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileStorage) Close() {
	fs.s.file.Close()
	fs.l.file.Close()
}
