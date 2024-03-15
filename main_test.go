package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
)

func TestMain(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		// Create a temporary directory for testing
		tempDir, err := os.MkdirTemp("", "synctool_test")
		if err != nil {
			t.Fatal("Failed to create temporary directory:", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a temporary input file with URLs
		tempFile, err := ioutil.TempFile("", "urls.txt")
		if err != nil {
			t.Fatal("Failed to create temporary file:", err)
		}
		defer tempFile.Close()

		// Write some URLs to the temporary input file
		urls := []string{"https://blog.badgerops.net/content/images/2020/03/badger.png", "https://www.bobrossquotes.com/bobs/bob.png"}
		for _, url := range urls {
			_, err := tempFile.WriteString(url + "\n")
			if err != nil {
				t.Fatal("Failed to write to temporary file:", err)
			}
		}

		// Run the main function with the temporary input file and output directory
		fmt.Println(("The temp file name is: " + tempFile.Name()))
		os.Args = []string{"downloader", "-f", tempFile.Name(), "-o", "./synctool_test"}
		main()

		// Check if the files were downloaded and saved correctly
		for _, url := range urls {
			filePath := filepath.Join("./synctool_test", path.Base(url))
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("File not found: %s", filePath)
			}
		}
		return
	}
}
func TestOsExit(t *testing.T) {
	if os.Getenv("BE_CRASHER") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=TestOsExit")
		cmd.Env = append(os.Environ(), "BE_CRASHER=1")
		err := cmd.Run()
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			return
		}
		t.Fatalf("Process ran with err %v, want os.Exit(1)", err)
	}
}
