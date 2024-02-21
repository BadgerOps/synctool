package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "downloader",
		Usage: "Download things and save them in a dir",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file, f",
				Aliases:  []string{"f"},
				Value:    "",
				Usage:    "The path to the input `FILE` containing urls.",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "outdir, d",
				Aliases:  []string{"o"},
				Value:    "./output",
				Usage:    "The `DIRECTORY` where the files will be saved to.",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {

			// create the output directory, if it does not exist already
			outDir := c.String("outdir")
			err := os.MkdirAll(outDir, 0755)
			if err != nil {
				return cli.Exit("Error creating output dir.", 1)
			}
			rawFile, err := readFile(c.String("file"))
			if err != nil {
				fmt.Println("Error reading input file:", err)
				os.Exit(1)
			}

			// for each url in given file, download the given file
			for _, url := range rawFile {
				// download the content
				fmt.Println(" check out the values! ", url)
				resp, err := http.Get(url)
				if err != nil {
					fmt.Println("Error downloading file:", url)
					continue
				}
				defer resp.Body.Close()

				// check response status code
				if resp.StatusCode != http.StatusOK {
					fmt.Println("Error: received non-200 status code from server for file:", url)
					continue
				}

				// read response body into a byte slice
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Println("Error reading response body:", err)
				}

				// write to output directory using your writeToDir function
				err = writeToDir(filepath.Join(outDir, path.Base(url)), body)
				if err != nil {
					fmt.Println("Error writing to output file:", err)
				}

				fmt.Printf("Downloaded and saved %s\n", url)
				return cli.Exit("Downloaded ", 0)
			}
			fmt.Println("Downloaded all files listed in:", c.String("file"))
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		cli.Exit(err.Error(), 1)
		os.Exit(1)
	}
}

func readFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := make([]string, 0, 10) // allocate some buffer
	buf := make([]byte, 32*1024)   // large enough buffer for reading line by line

	for {
		n, err := file.Read(buf)
		if n > 0 {
			line := strings.TrimSpace(string(buf[:n]))
			lines = append(lines, line)
		}
		if err != nil {
			if err == io.EOF {
				break // reached EOF
			}
			return nil, err
		}
	}

	return lines, nil
}

func writeToDir(path string, data []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed writing data to file: %w", err)
	}

	return nil
}
