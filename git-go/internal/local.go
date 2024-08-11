package internal

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type LocalRepository struct {
	RootName string
}

func (r *LocalRepository) GitDir() string {
	return r.RootName + "/.git"
}

func (r *LocalRepository) ObjectsName() string {
	return r.GitDir() + "/objects"
}

func (r *LocalRepository) RefsName() string {
	return r.GitDir() + "/refs"
}

func (r *LocalRepository) HeadName() string {
	return r.GitDir() + "/HEAD"
}

func (r *LocalRepository) Init() error {
	err := os.Mkdir(r.GitDir(), 0755)
	if err != nil {
		return fmt.Errorf("error creating directory %v, %v", r.GitDir(), err)
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

func (r *LocalRepository) WriteObject(hashHex string, content []byte) error {
	var compressedData bytes.Buffer
	writer := zlib.NewWriter(&compressedData)
	_, err := writer.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write to zlib writer, %v", err)
	}
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close zlip writer, %v", err)
	}
	hashDir := r.ObjectsName() + "/" + hashHex[:2]
	err = os.Mkdir(hashDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create dir %v, %v", hashDir, err)
	}
	hashFile := hashDir + "/" + hashHex[2:]
	err = os.WriteFile(hashFile, compressedData.Bytes(), 0755)
	if err != nil {
		return fmt.Errorf("failed to create file %v, %v", hashFile, err)
	}
	return nil
}

func (r *LocalRepository) ReadObject(hashHex string) (string, error) {
	filePath := filepath.Join(r.ObjectsName(), hashHex[:2], hashHex[2:])
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
	return decompressedData.String(), nil
}

func (r *LocalRepository) CatFile(hashHex string) (string, error) {
	content, err := r.ReadObject(hashHex)
	if err != nil {
		return "", fmt.Errorf("failed to read object %v", err)
	}
	return strings.Split(content, "\x00")[1], nil
}

func (r *LocalRepository) WriteBlobObject(filename string) ([]byte, string, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file %v, %v", filename, err)
	}

	fileContent := "blob\x20" + strconv.Itoa(len(file)) + "\x00" + string(file)

	hasher := sha1.New()
	_, err = hasher.Write([]byte(fileContent))
	if err != nil {
		return nil, "", fmt.Errorf("failed to write to sha1 hasher, %v", err)
	}
	hash := hasher.Sum(nil)
	hashHex := hex.EncodeToString(hash)

	r.WriteObject(hashHex, []byte(fileContent))

	return hash, hashHex, nil
}

func (r *LocalRepository) WriteTreeObject(dirname string) ([]byte, string, error) {
	files, err := os.ReadDir(dirname)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read dir %v, %v", dirname, err)
	}
	type entry struct {
		name    string
		mode    string
		byteSha []byte
	}
	var entries []entry
	for _, file := range files {
		fullname := fmt.Sprintf("%v/%v", dirname, file.Name())
		if file.Name() == ".git" {
			continue
		}
		if file.IsDir() {
			hash, _, err := r.WriteTreeObject(fullname)
			if err != nil {
				return nil, "", fmt.Errorf("failed to write tree object %v, %v", fullname, err)
			}
			entries = append(entries, entry{name: file.Name(), byteSha: hash, mode: "040000"})
			continue
		}
		hash, _, err := r.WriteBlobObject(fullname)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create blob object for %v, %v", fullname, err)
		}
		file.Type()
		entries = append(entries, entry{name: file.Name(), byteSha: hash, mode: "100644"})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	var treeContent []string
	treeContentLength := 0
	for _, entry := range entries {
		treeEntry := fmt.Sprintf("%v\x20%v\x00%v", entry.mode, entry.name, string(entry.byteSha))
		treeContent = append(treeContent, treeEntry)
		treeContentLength += len(treeEntry)
	}

	treeHeader := fmt.Sprintf("tree\x20%v\x00", treeContentLength)
	treeObject := []byte(treeHeader)
	for _, entry := range treeContent {
		treeObject = append(treeObject, []byte(entry)...)
	}

	hasher := sha1.New()
	_, err = hasher.Write(treeObject)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash tree object %v", err)
	}
	hash := hasher.Sum(nil)
	hashHex := hex.EncodeToString(hash)

	err = r.WriteObject(hashHex, treeObject)
	if err != nil {
		return nil, "", fmt.Errorf("failed to write object %v", err)
	}

	return hash, hashHex, nil
}

func (r *LocalRepository) ReadTreeObject(hashHex string, nameonly bool) ([]string, error) {
	content, err := r.ReadObject(hashHex)
	if err != nil {
		return nil, fmt.Errorf("failed to read object %v", err)
	}
	split := strings.Split(content, "\x00")
	data := split[1 : len(split)-1]

	if nameonly {
		names := []string{}
		for _, dByte := range data {
			splitByte := strings.Split(dByte, " ")[1]
			names = append(names, splitByte)
		}
		return names, nil
	}
	return nil, errors.New("read tree all not implemented")
}

func (r *LocalRepository) WriteCommitObject(treeSha string, parentSha string, message string) (string, error) {
	currentTime := time.Now()
	commitContent := fmt.Sprintf("tree %s\n", treeSha)
	if parentSha != "" {
		commitContent += fmt.Sprintf("parent %s\n", parentSha)
	}
	commitContent += fmt.Sprintf(
		"author author_name <author_email> %d %s\n\n%s\n",
		currentTime.Unix(),
		"+0000",
		message)
	commitHeader := fmt.Sprintf("commit %v\x00", len(commitContent))
	commitObject := commitHeader + commitContent

	hasher := sha1.New()
	_, err := hasher.Write([]byte(commitObject))
	if err != nil {
		return "", fmt.Errorf("failed to hash tree object %v", err)
	}
	hash := hasher.Sum(nil)
	hashHex := hex.EncodeToString(hash)

	r.WriteObject(hashHex, []byte(commitObject))
	return hashHex, nil
}
