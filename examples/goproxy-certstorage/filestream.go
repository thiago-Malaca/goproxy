package main

import (
	"errors"
	"os"
)

type FileStream struct {
	path string
	f    *os.File
}

func NewFileStream(path string) *FileStream {
	return &FileStream{path, nil}
}

func (fs *FileStream) Write(b []byte) (nr int, err error) {
	if fs.f == nil {
		fs.f, err = os.Create(fs.path)
		if err != nil {
			return 0, err
		}
	}
	// log.Println("Write:", b)
	return fs.f.Write(b)
}

func (fs *FileStream) Close() error {
	// fmt.Println("Close", fs.path)
	if fs.f == nil {
		return errors.New("FileStream was never written into")
	}
	return fs.f.Close()
}
