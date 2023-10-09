package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type (
	fileStorage struct {
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

func newFileStorage(path string) (*fileStorage, error) {
	s, err := newSaver(path)
	if err != nil {
		return nil, err
	}

	l, err := newLoader(path)
	if err != nil {
		return nil, err
	}

	return &fileStorage{
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

func (fs *fileStorage) save(m *MemStorage) error {
	err := fs.s.file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = fs.s.file.Seek(0, 0)
	if err != nil {
		return err
	}

	fs.s.encoder.SetIndent("", "    ")

	return fs.s.encoder.Encode(m)
}

func (fs *fileStorage) load() (*MemStorage, error) {
	m := &MemStorage{
		Storage: make(map[string]any, 30),
	}

	info, err := fs.l.file.Stat()
	if err != nil {
		return nil, err
	}

	if info.Size() > 0 {
		if err := fs.l.decoder.Decode(m); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (fs *fileStorage) close() {
	fs.s.file.Close()
	fs.l.file.Close()
}
