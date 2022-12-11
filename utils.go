package ip

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var shiftIndex = []int{24, 16, 8, 0}

func numToIp(num int) string {
	var arr = make([]int, 4)
	arr[0] = (num >> 24) & 0xff
	arr[1] = (num >> 16) & 0xff
	arr[2] = (num >> 8) & 0xff
	arr[3] = num & 0xff
	return fmt.Sprintf("%d.%d.%d.%d", arr[0], arr[1], arr[2], arr[3])
}

func CheckIP(ip string) (uint32, error) {
	var ps = strings.Split(ip, ".")
	if len(ps) != 4 {
		return 0, fmt.Errorf("invalid ip address `%s`", ip)
	}

	var val = uint32(0)
	for i, s := range ps {
		d, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("the %dth part `%s` is not an integer", i, s)
		}

		if d < 0 || d > 255 {
			return 0, fmt.Errorf("the %dth part `%s` should be an integer bettween 0 and 255", i, s)
		}

		val |= uint32(d) << shiftIndex[i]
	}

	return val, nil
}

func loadContent(handle *os.File) ([]byte, error) {
	fi, err := handle.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}

	size := fi.Size()

	_, err = handle.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("seek to get xdb file length: %w", err)
	}

	var buff = make([]byte, size)
	rLen, err := handle.Read(buff)
	if err != nil {
		return nil, err
	}

	if rLen != len(buff) {
		return nil, fmt.Errorf("incomplete read: readed bytes should be %d", len(buff))
	}

	return buff, nil
}

func loadContentFromFile(dbFile string) ([]byte, error) {
	handle, err := os.OpenFile(dbFile, os.O_RDONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("open xdb file `%s`: %w", dbFile, err)
	}

	return loadContent(handle)
}
