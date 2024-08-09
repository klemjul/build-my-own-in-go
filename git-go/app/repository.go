package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strings"
)

type Repository struct {
	RootName string
}

func (r *Repository) ObjectsName() string {
	return r.RootName + "/objects"
}

func (r *Repository) RefsName() string {
	return r.RootName + "/refs"
}

func (r *Repository) HeadName() string {
	return r.RootName + "/HEAD"
}

func (r *Repository) Init() error {
	err := os.Mkdir(r.RootName, 0755)
	if err != nil {
		return fmt.Errorf("error creating directory %v, %v", r.RootName, err)
	}
	err = os.Mkdir(r.ObjectsName(), 0755)
	if err != nil {
		return fmt.Errorf("error creating directory %v, %v", r.ObjectsName(), err)
	}
	err = os.Mkdir(r.RefsName(), 0755)
	if err != nil {
		return fmt.Errorf("error creating directory %v, %v", r.RefsName(), err)
	}
	err = os.WriteFile(r.HeadName(), []byte("ref: refs/heads/main\n"), 0755)
	if err != nil {
		return fmt.Errorf("error creating directory %v, %v", r.HeadName(), err)
	}
	return nil
}

func (r *Repository) CatFile(hash string) (string, error) {
	filePath := r.ObjectsName() + "/" + hash[:2] + "/" + hash[2:]
	file, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file %v, %v", filePath, err)
	}
	reader, err := zlib.NewReader(bytes.NewReader(file))
	if err != nil {
		return "", fmt.Errorf("failed to create zlib reader, %v", err)
	}
	defer reader.Close()
	var decompressedData bytes.Buffer
	_, err = io.Copy(&decompressedData, reader)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content to reader, %v", err)
	}
	return strings.Split(decompressedData.String(), "\x00")[1], err
}
