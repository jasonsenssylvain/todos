package main

import (
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	termutil "github.com/andrew-d/go-termutil"
)

//TODO fix ugly code?
//
func ReadLines(path string) ([]string, error) {
	f, err := os.OpenFile(path, os.O_CREATE, 0660)
	defer f.Close()

	if err != nil {
		return nil, err
	}

	return ReadLinesFromFile(f)
}

func ReadLinesFromFile(file *os.File) ([]string, error) {
	sc := bufio.NewScanner(file)
	lines := []string{}

	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	return lines, sc.Err()
}

func ReadStdin() ([]string, error) {
	if termutil.Isatty(os.Stdin.Fd()) {
		return nil, errors.New("stdin is empty")
	}

	bytes, err := ioutil.ReadAll(os.Stdin)
	for err != nil {
		return nil, err
	}

	return strings.Split(string(bytes), "\n"), nil
}

func WriteFile(filepath string, lines []string, executable bool) error {
	var mode os.FileMode
	if executable {
		mode = 0755
	} else {
		mode = 0660
	}

	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, mode)
	defer f.Close()

	if err != nil {
		return err
	}

	writer := bufio.NewWriter(f)
	for _, line := range lines {
		writer.WriteString(line + "\n")
	}
	os.Chmod(filepath, mode)
	return writer.Flush()
}
