// parser.go
//
// Component to parse wireguard config file

package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	// CommentPrefixes are the strings which start a comment
	CommentPrefixes = []string{"#", ";", "//"}
)

func checkRune(c rune) bool {
	upper := (c >= 65) && (c <= 90)
	lower := (c >= 97) && (c <= 122)
	number := (c >= 48) && (c <= 57)
	underscore := (c == '_')
	return upper || lower || number || underscore
}

func checkKey(key string) error {
	for _, r := range key {
		if !checkRune(r) {
			return fmt.Errorf("Key %s contains invalid characters ('%c')",
				key, r)
		}
	}
	return nil
}

// Section represents a basic block like [Interface] or [Peer]
type Section struct {
	name string
	data map[string]string
}

// NewSection creates a new empty section
func NewSection(name string) *Section {
	return &Section{
		name: name,
		data: make(map[string]string),
	}
}

// Name returns the name of the section
func (s *Section) Name() string {
	return s.name
}

// func (s *Section) Write(w io.Writer) (int, error) {
// 	// section name
// 	n, err := w.Write([]byte(fmt.Sprintf("[%s]\n", s.name)))
// 	if err != nil {
// 		return n, err
// 	}
// 	// key/value pairs
// 	for key, value := range s.data {
// 		m, err := w.Write([]byte(fmt.Sprintf("%s = %s\n", key, value)))
// 		if err != nil {
// 			return n + m, err
// 		}
// 		n += m
// 	}
// 	return n, nil
// }

// HasKey returns whether the section has the given key
func (s *Section) HasKey(key string) bool {
	_, exist := s.data[key]
	return exist
}

// Set defines a pair key/value
func (s *Section) Set(key string, value string) error {
	if err := checkKey(key); err != nil {
		return err
	}
	s.data[key] = value
	return nil
}

// Get returns the raw value (string) related to a key
func (s *Section) Get(key string) (string, error) {
	value, exist := s.data[key]
	if exist {
		return value, nil
	}
	return "", fmt.Errorf("Unknown key %s", key)
}

// GetInt returns a value given a key and tries to convert it
func (s *Section) GetInt(key string) (int, error) {
	value, err := s.Get(key)
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// GetBytesFromBase64 returns the byte array representing the base64 encoded string
func (s *Section) GetBytesFromBase64(key string) ([]byte, error) {
	value, err := s.Get(key)
	if err != nil {
		return nil, err
	}
	b, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GetKeyFromBase64 returns the key represented by the base64 encoded string
func (s *Section) GetKeyFromBase64(key string) (Key, error) {
	value, err := s.GetBytesFromBase64(key)
	if err != nil {
		return nil, err
	}
	k := NewKey()
	err = k.UpdateFromBytes(value)
	if err != nil {
		return nil, err
	}
	return k, nil
}

// GetIP returns the net.IP object parsed from the value
// func (s *Section) GetIP(key string) (net.IP, error) {
// 	value, err := s.Get(key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ipnet, err := parseAddressAndMask(value)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return ipnet.IP, nil
// }

// GetIPArray returns an array of net.IP objects parsed from the value
func (s *Section) GetIPArray(key string) ([]net.IP, error) {
	value, err := s.Get(key)
	if err != nil {
		return nil, err
	}
	// split
	values := strings.Split(value, ",")
	ipList := make([]net.IP, len(values))
	for i, v := range values {
		// parse IP
		ip := net.ParseIP(strings.TrimSpace(v))
		// ip, _, err := net.ParseCIDR(strings.TrimSpace(v))
		if ip == nil {
			return nil, fmt.Errorf("Error while parsing IP %s", s)
		}
		// if err != nil {
		// 	return nil, err
		// }
		ipList[i] = ip
	}

	return ipList, nil
}

// GetIPNet returns the net.IPNet object parsed from the value
// func (s *Section) GetIPNet(key string) (*net.IPNet, error) {
// 	value, err := s.Get(key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ipnet, err := parseAddressAndMask(value)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return ipnet, nil
// }

// GetNetSlice returns an array of *net.IPNet object parsed from the value
func (s *Section) GetNetSlice(key string) (*NetSlice, error) {
	addr, err := s.Get(key)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("KEY: %s, VALUE: %s\n", key, addr)
	networks, err := NewNetSliceFromString(addr)
	// networks, err := mapIPNetStrList(strings.Split(addr, ","))
	if err != nil {
		return nil, fmt.Errorf("Error while parsing %s (%v)", addr, err)
	}
	return &networks, nil
}

func (s *Section) String() string {
	str := fmt.Sprintf("[%s]\n", s.name)
	for key, value := range s.data {
		str += fmt.Sprintf("%s = %s\n", key, value)
	}
	return str
}

// File represents a config file: a list of sections
type File struct {
	sections []*Section
}

// NewFile creates a new empty file
func NewFile() *File {
	return &File{sections: make([]*Section, 0)}
}

func removeComment(line string) string {
	for _, pre := range CommentPrefixes {
		if strings.HasPrefix(line, pre) {
			return ""
		}
	}
	return line
}

// ParseFile reads a file and store data to a File object
func ParseFile(p string) (*File, error) {
	buf, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(buf)
	file := NewFile()
	rex, err := regexp.Compile(`\[(.*?)\]`)
	if err != nil {
		return nil, err
	}

	section := file.AddSection("DEFAULT")
	for {
		// get line
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			return file, nil
		} else if err != nil {
			return file, err
		}

		// trim line
		line = strings.TrimSpace(line)
		// remove comments
		line = removeComment(line)
		// check
		if len(line) > 0 {
			if title := rex.FindString(line); title != "" {
				// section case
				// remove [ and ]
				section = file.AddSection(title[1 : len(title)-1])
			} else if index := strings.Index(line, "="); index > 0 {
				// key - value pair case
				key := strings.TrimSpace(line[:index])
				value := strings.TrimSpace(line[index+1:])
				// check if key is valid
				if err := section.Set(key, value); err != nil {
					return nil, err
				}
			}

		}
	}
}

// Sections returns the list of the sections
func (f *File) Sections() []*Section {
	return f.sections
}

// HasSection check whether the file contains the given section
func (f *File) HasSection(name string) bool {
	for _, sec := range f.sections {
		if sec.name == name {
			return true
		}
	}
	return false
}

// AddSection adds a new section to the file
func (f *File) AddSection(name string) *Section {
	sec := NewSection(name)
	f.sections = append(f.sections, sec)
	return sec
}

// GetSection returns a section given its name
func (f *File) GetSection(name string) (*Section, error) {
	for _, sec := range f.sections {
		if sec.name == name {
			return sec, nil
		}
	}
	return nil, fmt.Errorf("Unknown section")
}

func (f *File) String() string {
	str := ""
	index := 0
	// treat DEFAULT section (no name)
	sec, err := f.GetSection("DEFAULT")
	if err == nil {
		if len(sec.data) > 0 {
			for k, v := range sec.data {
				str += fmt.Sprintf("%s = %s\n", k, v)
			}
			str += "\n"
		}
		index = 1
	}
	for _, s := range f.sections[index:] {
		str += s.String()
		str += "\n"
	}
	return str
}

// Save stores the config to a file
func (f *File) Save(path string) error {
	return ioutil.WriteFile(path, []byte(f.String()), 0600)
}
