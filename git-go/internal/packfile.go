package internal

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// https://github.com/git/git/blob/795ea8776befc95ea2becd8020c7a284677b4161/Documentation/gitformat-pack.txt#L70
type PackFileObjectType int

const (
	OBJ_COMMIT    PackFileObjectType = 1
	OBJ_TREE      PackFileObjectType = 2
	OBJ_BLOB      PackFileObjectType = 3
	OBJ_TAG       PackFileObjectType = 4
	OBJ_OFS_DELTA PackFileObjectType = 6
	OBJ_REF_DELTA PackFileObjectType = 7
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

// https://git-scm.com/docs/pack-format
// https://stefan.saasen.me/articles/git-clone-in-haskell-from-the-bottom-up/#implementing-pack-file-negotiation
// https://bitbucket.org/ssaasen/git/src/master/Documentation/technical/pack-format.txt
func ParsePackFile(packFile []byte) ([]GitObject, []GitObjectDelta, error) {
	packMagicBytes := string(packFile[:4])
	packVersion := binary.BigEndian.Uint32(packFile[4:8])
	packNbObjects := binary.BigEndian.Uint32(packFile[8:12])
	if packMagicBytes != "PACK" {
		return nil, nil, errors.New("invalid pack file header, not containing PACK on magic bytes")
	}
	if packVersion != 2 {
		return nil, nil, errors.New("invalid pack file header, version != 2")
	}

	bytesRead := int64(12)
	objects := []GitObject{}
	deltas := []GitObjectDelta{}

	for i := 0; i < int(packNbObjects); i++ {
		headers, err := readObjectHeaders(packFile[bytesRead:])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse header on %v object, after %v byte read, %v", i, bytesRead, err)
		}

		bytesRead += int64(headers.HeaderSize)

		if headers.ObjectType == OBJ_COMMIT || headers.ObjectType == OBJ_TREE || headers.ObjectType == OBJ_BLOB || headers.ObjectType == OBJ_TAG {
			read, object, err := readObjectContent(packFile[bytesRead:])
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse object content on %v object, after %v byte read, %v", i, bytesRead, err)
			}

			objName, err := parseGitObjectName(headers.ObjectType)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse git object name on %v object, after %v byte read, %v", i, bytesRead, err)
			}

			if headers.ContentSize != int64(len(object)) {
				return nil, nil, fmt.Errorf("object of type %v has bad length, expected %v, has %v", headers.ObjectType, headers.ContentSize, len(object))
			}

			bytesRead += int64(read)
			objects = append(objects, GitObject{ObjectName: objName, Content: object, ContentSize: headers.ContentSize})

		} else if headers.ObjectType == OBJ_REF_DELTA {
			// 20 first bytes are sha to apply delta
			hash := packFile[bytesRead : bytesRead+20]
			bytesRead += 20

			read, object, err := readObjectContent(packFile[bytesRead:])
			if err != nil {
				return nil, nil, fmt.Errorf(fmt.Sprintf("failed to read delta, %v", err))
			}
			bytesRead += int64(read)

			if headers.ContentSize != int64(len(object)) {
				return nil, nil, fmt.Errorf("object of type %v has bad length, expected %v, has %v", headers.ObjectType, headers.ContentSize, len(object))
			}
			deltas = append(deltas, GitObjectDelta{ObjectSha: hex.EncodeToString(hash), Content: object, ContentSize: headers.ContentSize})

		} else if headers.ObjectType == OBJ_OFS_DELTA {
			return nil, nil, errors.New("OBJ_OFS_DELTA not implemented")
		} else {
			return nil, nil, fmt.Errorf("invalid object type %v on %v object, after %v byte read", headers.ObjectType, i, bytesRead)
		}

	}
	return objects, deltas, nil
}

func readObjectHeaders(packFile []byte) (*PackfileObjectHeader, error) {
	contentType := int(packFile[0] & objTypeMask >> 4)
	contentSize := int64(packFile[0] & initSizeMask)

	size, read := readVariableObjectSize(packFile, 4, contentSize)
	// add first byte read
	read += 1

	return &PackfileObjectHeader{
		ObjectType:  PackFileObjectType(contentType),
		HeaderSize:  read,
		ContentSize: size,
	}, nil
}

/*
From each byte, the seven least significant bits are
used to form the resulting integer. As long as the most significant
bit is 1, this process continues; the byte with MSB 0 provides the
last seven bits.  The seven-bit chunks are concatenated.
*/
func readVariableObjectSize(packFile []byte, initialShift int, initialSize int64) (int64, int) {
	read := 0
	size := initialSize
	shift := initialShift
	for packFile[read]&msbMask != 0 {
		read++
		size += int64(packFile[read]&sizeMask) << shift
		shift += 7
	}
	return size, read
}

func readObjectContent(packfile []byte) (int, []byte, error) {
	b := bytes.NewReader(packfile)

	r, err := zlib.NewReader(b)
	if err != nil {
		return 0, nil, fmt.Errorf("readPackFileObject failed to init zlib reader, %v", err)
	}
	defer r.Close()
	content, err := io.ReadAll(r)
	if err != nil {
		return 0, nil, fmt.Errorf("readPackFileObject failed to read zlib stream, %v", err)
	}
	bytesRead := int(b.Size()) - b.Len()
	return bytesRead, content, nil

}

func parseGitObjectName(objType PackFileObjectType) (string, error) {
	var objectTypeMapping = map[PackFileObjectType]string{
		OBJ_COMMIT: "commit",
		OBJ_TREE:   "tree",
		OBJ_BLOB:   "blob",
		OBJ_TAG:    "tag",
	}

	if name, ok := objectTypeMapping[objType]; ok {
		return name, nil
	}
	return "", errors.New("invalid PackFileObjectType code")
}
