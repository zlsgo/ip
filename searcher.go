package ip

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	HeaderInfoLength      = 256
	VectorIndexRows       = 256
	VectorIndexCols       = 256
	VectorIndexSize       = 8
	SegmentIndexBlockSize = 14
)

type IndexPolicy int

const (
	VectorIndexPolicy IndexPolicy = 1
	BTreeIndexPolicy  IndexPolicy = 2
)

func (i IndexPolicy) String() string {
	switch i {
	case VectorIndexPolicy:
		return "VectorIndex"
	case BTreeIndexPolicy:
		return "BtreeIndex"
	default:
		return "unknown"
	}
}

type Searcher struct {
	handle *os.File

	ioCount int

	vectorIndex []byte

	contentBuff []byte
}

func baseNew(dbFile string, vIndex []byte, cBuff []byte) (*Searcher, error) {
	var err error

	if cBuff != nil {
		return &Searcher{
			vectorIndex: nil,
			contentBuff: cBuff,
		}, nil
	}

	handle, err := os.OpenFile(dbFile, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}

	return &Searcher{
		handle:      handle,
		vectorIndex: vIndex,
	}, nil
}

func NewWithBuffer(cBuff []byte) (*Searcher, error) {
	return baseNew("", nil, cBuff)
}

func (s *Searcher) Close() {
	if s.handle != nil {
		err := s.handle.Close()
		if err != nil {
			return
		}
	}
}

func (s *Searcher) GetIOCount() int {
	return s.ioCount
}

func (s *Searcher) SearchByStr(str string) (string, error) {
	ip, err := CheckIP(str)
	if err != nil {
		return "", err
	}

	return s.Search(ip)
}

func (s *Searcher) Search(ip uint32) (string, error) {
	s.ioCount = 0
	var (
		il0  = (ip >> 24) & 0xFF
		il1  = (ip >> 16) & 0xFF
		idx  = il0*VectorIndexCols*VectorIndexSize + il1*VectorIndexSize
		sPtr uint32
		ePtr uint32
	)
	if s.vectorIndex != nil {
		sPtr = binary.LittleEndian.Uint32(s.vectorIndex[idx:])
		ePtr = binary.LittleEndian.Uint32(s.vectorIndex[idx+4:])
	} else if s.contentBuff != nil {
		sPtr = binary.LittleEndian.Uint32(s.contentBuff[HeaderInfoLength+idx:])
		ePtr = binary.LittleEndian.Uint32(s.contentBuff[HeaderInfoLength+idx+4:])
	} else {
		var buff = make([]byte, VectorIndexSize)
		err := s.read(int64(HeaderInfoLength+idx), buff)
		if err != nil {
			return "", fmt.Errorf("read vector index block at %d: %w", HeaderInfoLength+idx, err)
		}

		sPtr = binary.LittleEndian.Uint32(buff)
		ePtr = binary.LittleEndian.Uint32(buff[4:])
	}

	var dataLen, dataPtr = 0, uint32(0)
	var buff = make([]byte, SegmentIndexBlockSize)
	var l, h = 0, int((ePtr - sPtr) / SegmentIndexBlockSize)
	for l <= h {
		m := (l + h) >> 1
		p := sPtr + uint32(m*SegmentIndexBlockSize)
		err := s.read(int64(p), buff)
		if err != nil {
			return "", fmt.Errorf("read segment index at %d: %w", p, err)
		}

		sip := binary.LittleEndian.Uint32(buff)
		if ip < sip {
			h = m - 1
		} else {
			eip := binary.LittleEndian.Uint32(buff[4:])
			if ip > eip {
				l = m + 1
			} else {
				dataLen = int(binary.LittleEndian.Uint16(buff[8:]))
				dataPtr = binary.LittleEndian.Uint32(buff[10:])
				break
			}
		}
	}

	if dataLen == 0 {
		return "", nil
	}

	var regionBuff = make([]byte, dataLen)
	err := s.read(int64(dataPtr), regionBuff)
	if err != nil {
		return "", fmt.Errorf("read region at %d: %w", dataPtr, err)
	}

	return string(regionBuff), nil
}

func (s *Searcher) read(offset int64, buff []byte) error {
	if s.contentBuff != nil {
		cLen := copy(buff, s.contentBuff[offset:])
		if cLen != len(buff) {
			return fmt.Errorf("incomplete read: readed bytes should be %d", len(buff))
		}
	} else {
		_, err := s.handle.Seek(offset, 0)
		if err != nil {
			return fmt.Errorf("seek to %d: %w", offset, err)
		}

		s.ioCount++
		rLen, err := s.handle.Read(buff)
		if err != nil {
			return fmt.Errorf("handle read: %w", err)
		}

		if rLen != len(buff) {
			return fmt.Errorf("incomplete read: readed bytes should be %d", len(buff))
		}
	}

	return nil
}
