package tailfile

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestTailFile(t *testing.T) {
	lines, _ := readLastNLines("1.txt", 100)
	fmt.Println(lines)
}

func readLastNLines(filename string, n int64) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}
	fileSize := fileInfo.Size()
	// 没有内容
	if fileSize <= 0 {
		return "", nil
	}
	var lines []string
	var buffer []byte
	var lineCount int64
	newline := byte('\n')

	// 从最后往前读 读完或者读够行数就结束
	for offset := int64(1); offset <= fileSize; offset++ {
		file.Seek(-offset, io.SeekEnd)
		char := make([]byte, 1)
		_, err := file.Read(char)
		if err != nil {
			return "", err
		}
		if char[0] == newline {
			if len(buffer) > 0 {
				// 反转
				for i, j := 0, len(buffer)-1; i < j; i, j = i+1, j-1 {
					buffer[i], buffer[j] = buffer[j], buffer[i]
				}
				lines = append(lines, string(buffer))
				buffer = []byte{}
				lineCount++
				if lineCount >= n {
					break
				}
			}
		} else {
			buffer = append(buffer, char[0])
		}
	}

	// 最后一行不是以\n结尾
	if len(buffer) > 0 {
		for i, j := 0, len(buffer)-1; i < j; i, j = i+1, j-1 {
			buffer[i], buffer[j] = buffer[j], buffer[i]
		}
		lines = append(lines, string(buffer))
	}

	// 反转
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}

	return strings.Join(lines, "\n"), nil
}
