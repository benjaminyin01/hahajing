package com

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// FileInfo x
type FileInfo struct {
	Type    byte
	OrgName string
}

// Ed2kFileLink x
type Ed2kFileLink struct {
	FileInfo

	// attributes from ED2K
	Name  string
	Size  uint64
	Avail uint32
	Hash  []byte
}

type ed2kFileLinkJSON struct {
	FileInfo

	// attributes from ED2K
	Name  string
	Size  uint64
	Avail uint32
	Link  string
}

// ToFileInfo is converting file name from KAD or DHT via item from Internet(DouBan)
func ToFileInfo(name string) *FileInfo {
	// match
	lowerName := strings.ToLower(name)
	var fileInfo *FileInfo = &FileInfo{Type: 0, OrgName: lowerName}
	return fileInfo
}

// GetEd2kLink x
func (f *Ed2kFileLink) GetEd2kLink() string {
	return GetEd2kLink(f.Name, f.Size, f.Hash)
}

// ToJSON x
func (f *Ed2kFileLink) ToJSON() []byte {
	linkJSON := ed2kFileLinkJSON{
		FileInfo: f.FileInfo,
		Name:     f.Name,
		Size:     f.Size,
		Avail:    f.Avail,
		Link:     f.GetEd2kLink()}

	b, _ := json.Marshal(linkJSON)
	return b
}

// GetEd2kLink is getting ED2K link by file name, size and hash from eMule KAD network.
func GetEd2kLink(name string, size uint64, hash []byte) string {
	newHash := ConvertEd2kHash32(hash)
	return fmt.Sprintf("ed2k://|file|%s|%d|%s|/",
		encodeURLUtf8(stripInvalidFileNameChars(name)),
		size,
		encodeBase16(newHash[:]))
}

// ConvertEd2kHash32 x
func ConvertEd2kHash32(srcHash []byte) [16]byte {
	// change to inverse endian for each uint32
	hash := [16]byte{}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			hash[i*4+j] = srcHash[i*4+3-j]
		}
	}
	return hash
}

var reservedFileNames = [...]string{"NUL", "CON", "PRN", "AUX", "CLOCK$",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}

func stripInvalidFileNameChars(text string) string {
	var dst string
	for _, c := range text {
		if (c >= 0 && c <= 31) ||
			c == '"' || c == '*' || c == '<' || c == '>' ||
			c == '?' || c == '|' || c == '\\' || c == ':' {
			continue
		}
		dst += string(c)
	}

	for _, prefix := range reservedFileNames {
		if len(dst) < len(prefix) {
			continue
		}

		if dst[:len(prefix)] == prefix {
			if len(dst) == len(prefix) {
				dst += string('_')
			} else if dst[len(prefix)] == '.' {
				s := []rune(dst)
				s[len(prefix)] = '_'
				dst = string(s)
			}
		}
	}

	return dst
}

func encodeURLUtf8(str string) string {
	utf8 := []byte(str)
	var url string
	for _, b := range utf8 {
		if b == byte('%') || b == byte(' ') || b >= 0x7F {
			s := fmt.Sprintf("%%%02X", b)
			url += s
		} else {
			url += string(b)
		}
	}

	return url
}

func encodeBase16(buf []byte) string {
	s := hex.EncodeToString(buf)
	return strings.ToUpper(s)
}
