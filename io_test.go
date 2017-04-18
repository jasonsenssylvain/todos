package todos

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Test_ReadStdin(t *testing.T) {
	src, err := ReadStdin()
	if err != nil {
		t.Fail()
	} else {
		fmt.Println(src)
	}
}

func Test_WriteFile(t *testing.T) {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	lines := []string{"hi", "second line"}
	fmt.Println("path is " + path)
	err := WriteFile(path+"..test", lines, false)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	} else {
		f, _ := os.OpenFile(path+"..test", os.O_CREATE, 0660)
		sclines, _ := ReadLinesFromFile(f)
		fmt.Println(sclines)
	}
}
