package swz

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func EncryptToFile(dir string, key uint32, seed uint32) error {
	os.Mkdir("encrypt", os.ModePerm)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	fileNames := make([]string, 0)
	for _, e := range entries {
		if !e.IsDir() {
			fileNames = append(fileNames, filepath.Join(dir, e.Name()))
		}
	}

	data := make([]string, 0)
	for _, f := range fileNames {
		if d, err := os.ReadFile(f); err == nil {
			data = append(data, string(d))
		}
	}

	fmt.Println("decrypting", len(fileNames), "files")
	encrypted, err := Encrypt(key, seed, data)
	if err != nil {
		return err
	}

	dest := filepath.Join("encrypt", strings.TrimSuffix(filepath.Base(dir), filepath.Ext(dir))+".swz")
	fmt.Println("writing to: " + dest)
	os.WriteFile(dest, encrypted, os.ModePerm)

	return nil
}

func Encrypt(key uint32, seed uint32, stringEntries []string) ([]byte, error) {
	rand := newPrng(seed ^ key)

	var hash uint32 = 0x2DF4A1CD
	hash_rounds := key%0x1F + 5

	for i := 0; i < int(hash_rounds); i++ {
		hash ^= rand.nextUint()
	}

	buffer := new(bytes.Buffer)
	writeUint32BE(buffer, hash)
	writeUint32BE(buffer, seed)

	for _, v := range stringEntries {
		if err := writeStringEntry([]byte(v), rand, buffer); err != nil {
			return buffer.Bytes(), err
		}
	}

	return buffer.Bytes(), nil
}

func writeStringEntry(input []byte, rand *prng, output *bytes.Buffer) error {
	compressedInput := new(bytes.Buffer)
	w, err := zlib.NewWriterLevel(compressedInput, zlib.BestCompression)
	if err != nil {
		return err
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

	return nil
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
