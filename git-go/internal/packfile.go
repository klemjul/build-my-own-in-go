package internal

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type PackFileObjectType int

const (
	OBJ_COMMIT    PackFileObjectType = 1
	OBJ_TREE                         = 2
	OBJ_BLOB                         = 3
	OBJ_TAG                          = 4
	OBJ_OFS_DELTA                    = 6
	OBJ_REF_DELTA                    = 7
)

const (
	msbMask      = byte(0b_1000_0000)
	objTypeMask  = byte(0b_0111_0000)
	initSizeMask = byte(0b_0000_1111)
	sizeMask     = byte(0b_0111_1111)
)

type PackfileObjectHeader struct {
	ObjectType  PackFileObjectType
	HeaderSize  int
	ContentSize int64
}

func parsePackFile(raw []byte) ([]PackfileObjectHeader, [][]byte, int, error) {
	if len(raw) < 32 {
		return nil, nil, 0, errors.New("invalid pack file header")
	}

	packMagicBytes := string(raw[:4])
	packVersion := binary.BigEndian.Uint32(raw[4:8])
	packNbObjects := binary.BigEndian.Uint32(raw[8:12])
	fmt.Printf("%v V%v, nb = %v ", packMagicBytes, packVersion, packNbObjects)

	bytesRead := int64(12)
	packHeaders := make([]PackfileObjectHeader, 3)
	packObjects := make([][]byte, 3)
	// TODO: fix length issue
	for i := range 3 {
		headers, err := readPackFileObjectHeaders(raw[bytesRead:])
		packHeaders[i] = *headers
		bytesRead += int64(headers.HeaderSize)
		if err != nil {
			return nil, nil, 0, err
		}
		read, content, err := readPackFileObject(raw[bytesRead:])
		if err != nil {
			return nil, nil, 0, err
		}
		packObjects[i] = content
		bytesRead += int64(read)
	}
	return packHeaders, packObjects, int(packNbObjects), nil
}

func readPackFileObjectHeaders(packFile []byte) (*PackfileObjectHeader, error) {
	bytesRead := 1

	contentType := int(packFile[0] & objTypeMask >> 4)
	contentSize := int64(packFile[0] & initSizeMask)
	sizeShift := uint(4)

	if packFile[0]&msbMask == 0 {
		return nil, errors.New("invalid packfile, first bit MSB it not 1")
	}
	for {
		nextByte := packFile[bytesRead]
		bytesRead += 1

		contentSize += int64(nextByte&sizeMask) << sizeShift
		sizeShift += 7
		if nextByte&msbMask == 0 {
			break
		}
	}

	return &PackfileObjectHeader{
		ObjectType:  PackFileObjectType(contentType),
		HeaderSize:  bytesRead,
		ContentSize: contentSize,
	}, nil
}

func readPackFileObject(packfile []byte) (int, []byte, error) {
	b := bytes.NewReader(packfile)
	r, err := zlib.NewReader(b)
	if err != nil {
		return 0, nil, err
	}
	defer r.Close()
	content, err := io.ReadAll(r)
	if err != nil {
		return 0, nil, err
	}
	bytesRead := int(b.Size()) - b.Len()
	return bytesRead, content, nil
}
