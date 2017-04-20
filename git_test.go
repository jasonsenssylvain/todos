package main

import (
	"fmt"
	"testing"
)

func Test_GitDirectoryRoot(t *testing.T) {
	res, err := GitDirectoryRoot()
	if err != nil {
		t.Fail()
	} else {
		fmt.Println("git root is " + res)
	}
}

func Test_GitRemoteUrl(t *testing.T) {
	res, err := GitRemoteUrl()
	if err != nil {
		t.Fail()
	} else {
		fmt.Println("git url is " + res)
	}
}

func Test_GitOwner(t *testing.T) {
	res, err := GitOwner()
	if err != nil {
		t.Fail()
	} else {
		fmt.Println("git owner is " + res)
	}
}
