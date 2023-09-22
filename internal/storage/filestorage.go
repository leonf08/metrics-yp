package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type FileStorage struct {
	s *saver
	l *loader
}

type saver struct {
	file    *os.File
	encoder *json.Encoder
}

type loader struct {
	file    *os.File
	decoder *json.Decoder
}

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

func (fs *FileStorage) SaveInFile(m any) error {
	mt, ok := m.(*MemStorage)
	if !ok {
		return errors.New("invalid type assertion")
	}
	return fs.s.save(mt)
}

func (fs *FileStorage) LoadFromFile() (*MemStorage, error) {
	return fs.l.load()
}

func (fs *FileStorage) CloseFileStorage() {
	fs.s.close()
	fs.l.close()
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

func (s *saver) save(m *MemStorage) error {
	err := s.file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = s.file.Seek(0, 0)
	if err != nil {
		return err
	}

	s.encoder.SetIndent("", "    ")

	return s.encoder.Encode(m)
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

func (l *loader) load() (*MemStorage, error) {
	m := &MemStorage{
		Storage: make(map[string]any),
	}

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
