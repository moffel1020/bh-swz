package swz

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func DecryptFile(file string, key uint32) {
	input, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	// name of file without extension or path
	swzName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

	os.MkdirAll(filepath.Join("dump", swzName), os.ModePerm)

	newData := Decrypt(input, key)
	fmt.Println("total files:", len(newData))

	for i, v := range newData {
		var fileName string
		if v[0] == '<' {
			if v[0:10] == "<LevelDesc" {
				fileName = strings.Split(strings.Split(v, "LevelName=\"")[1], "\"")[0] + ".xml"
			} else {
				fileName = "unknown" + fmt.Sprint(i) + ".xml"
			}
		} else {
			fileName = strings.Split(v, "\n")[0] + ".csv"
		}
		os.WriteFile(filepath.Join("dump", swzName, fileName), []byte(v), os.ModePerm)
	}
}

func Decrypt(input []byte, key uint32) []string {
	reader := bytes.NewReader(input)

	checksum := readUint32BE(reader)
	seed := readUint32BE(reader)

	fmt.Println("seed: " + fmt.Sprint(seed))

	rand := newWell512(seed ^ key)

	var hash uint32 = 0x2DF4A1CD
	hash_rounds := key%0x1F + 5

	for i := 0; i < int(hash_rounds); i++ {
		hash ^= rand.nextUint()
	}

	if hash != checksum {
		panic("hash is not equal to checksum")
	}

	results := make([]string, 0)
	for i := 0; reader.Len() > 0; i++ {
		text, err := readStringEntry(reader, rand)

		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}

		results = append(results, text)
	}

	return results
}

func readStringEntry(reader *bytes.Reader, rand *well512) (string, error) {

	if reader.Len() < 4 {
		return "", io.EOF
	}

	compressedSize := readUint32BE(reader) ^ rand.nextUint()
	_ = readUint32BE(reader) ^ rand.nextUint()
	checksum := readUint32BE(reader)

	buffer := make([]byte, compressedSize)
	_, err := reader.Read(buffer)
	if err != nil {
		panic(err)
	}

	hash := rand.nextUint()

	for i := 0; i < int(compressedSize); i++ {
		shift := i & 0xF
		buffer[i] ^= byte(((uint32(0xFF) << shift) & rand.nextUint()) >> shift)

		hash = uint32(buffer[i]) ^ rotateRight(hash, i%7+1)
	}

	if hash != checksum {
		panic("hash is not equal to checksum")
	}

	text := new(strings.Builder)
	r, err := zlib.NewReader(bytes.NewBuffer(buffer))
	if err != nil {
		panic(err)
	}
	io.Copy(text, r)
	r.Close()

	return text.String(), nil
}

func readUint32BE(reader *bytes.Reader) uint32 {
	buffer := make([]byte, 4)
	if _, err := reader.Read(buffer); err != nil {
		panic(err)
	}

	return (uint32(buffer[3]) | (uint32(buffer[2]) << 8) | (uint32(buffer[1]) << 16) | (uint32(buffer[0]) << 24))
}

func rotateRight(v uint32, bits int) uint32 {
	return (v >> bits) | (v << (32 - bits))
}
