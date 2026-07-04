package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/asiffer/wg-easy-vpn/export"
	"github.com/rs/zerolog"
)

const DEFAULT_SECTION = "DEFAULT"

// File represents a config file: a list of sections
type File struct {
	sections []*Section
}

// NewFile creates a new empty file
func NewFile() *File {
	return &File{sections: make([]*Section, 0)}
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

	section := file.AddSection(DEFAULT_SECTION)
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
	return nil, fmt.Errorf("unknown section")
}

// GetSection returns a section given its name
func (f *File) GetorCreateSection(name string) *Section {
	sec, err := f.GetSection(name)
	if err == nil {
		return sec
	}
	sec = f.AddSection(name)
	return sec
}

func (f *File) String() string {
	str := ""
	index := 0
	// treat DEFAULT section (no name)
	sec, err := f.GetSection(DEFAULT_SECTION)
	if err == nil {
		str += sec.StringNoHeader()
		str += "\n"
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
	return os.WriteFile(path, []byte(f.String()), 0600)
}

func (f *File) WriteTo(w io.Writer) (int64, error) {
	size, err := w.Write([]byte(f.String()))
	return int64(size), err
}

func (f *File) WriteQRCodeTo(w io.Writer) (int64, error) {
	reader := strings.NewReader(f.String())
	img, err := export.Encode(reader)
	if err != nil {
		return 0, err
	}
	n, err := w.Write([]byte(export.QRCodeToString(img)))
	return int64(n), err
}

func (f *File) Log(event *zerolog.Event) *zerolog.Event {
	for _, s := range f.sections {
		if s.name != DEFAULT_SECTION {
			event = s.Log(event)
		} else {
			event = s.LogWithoutName(event)
		}
	}
	return event
}
