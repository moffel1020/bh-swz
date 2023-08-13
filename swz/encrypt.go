package swz

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func EncryptToFile(name string, key uint32, seed uint32) {
	os.Mkdir("encrypt", os.ModePerm)

	swzName := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))

	entries, err := os.ReadDir(filepath.Join("dump", swzName))
	if err != nil {
		panic(err)
	}

	fileNames := make([]string, 0)
	for _, e := range entries {
		if !e.IsDir() {
			fileNames = append(fileNames, e.Name())
		}
	}

	data := make([]string, 0)
	for _, f := range fileNames {
		if d, err := os.ReadFile(filepath.Join("dump", swzName, f)); err == nil {
			data = append(data, string(d))
		}
	}

	encrypted := Encrypt(key, seed, data)

	fmt.Println("writing to: " + filepath.Join("encrypt", swzName+".swz"))
	os.WriteFile(filepath.Join("encrypt", swzName+".swz"), encrypted, os.ModePerm)
}

func Encrypt(key uint32, seed uint32, stringEntries []string) []byte {
	rand := newWell512(seed ^ key)

	var hash uint32 = 0x2DF4A1CD
	hash_rounds := key%0x1F + 5

	for i := 0; i < int(hash_rounds); i++ {
		hash ^= rand.nextUint()
	}

	buffer := new(bytes.Buffer)
	writeUint32BE(buffer, hash)
	writeUint32BE(buffer, seed)

	for _, v := range stringEntries {
		writeStringEntry([]byte(v), rand, buffer)
	}

	return buffer.Bytes()
}

func writeStringEntry(input []byte, rand *well512, output *bytes.Buffer) {
	compressedInput := new(bytes.Buffer)
	w, err := zlib.NewWriterLevel(compressedInput, zlib.BestCompression)
	if err != nil {
		panic(err)
	}
	w.Write(input)
	w.Close()

	compressedSize := uint32(compressedInput.Len()) ^ rand.nextUint()
	decompressedSize := uint32(len(input)) ^ rand.nextUint()

	checksum := rand.nextUint()

	for i := 0; i < compressedInput.Len(); i++ {
		checksum = uint32(compressedInput.Bytes()[i]) ^ rotateRight(checksum, i%7+1)

		shift := i & 0xF
		compressedInput.Bytes()[i] ^= byte((((uint32(0xFF) << shift) & rand.nextUint()) >> shift))
	}

	writeUint32BE(output, compressedSize)
	writeUint32BE(output, decompressedSize)
	writeUint32BE(output, checksum)
	output.Write(compressedInput.Bytes())
}

func writeUint32BE(buffer *bytes.Buffer, value uint32) {
	buf := []byte{
		byte((value >> 24)) & 0xFF,
		byte((value >> 16)) & 0xFF,
		byte((value >> 8)) & 0xFF,
		byte((value >> 0)) & 0xFF,
	}

	if _, err := buffer.Write(buf); err != nil {
		panic(err)
	}
}
