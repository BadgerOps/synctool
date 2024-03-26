package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

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

			getURL(rawFile, outDir)
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

func getURL(rawFile []string, outDir string) {
	totalDownloadTime := time.Duration(0)
	totalFileSize := int64(0)

	for _, url := range rawFile {
		startTime := time.Now()
		fmt.Println("Downloading file from URL: ", url)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error downloading file:", url)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Error: received non-200 status code from server for file:", url)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
		}

		filePath := filepath.Join(outDir, path.Base(url))
		err = writeToDir(filePath, body)
		if err != nil {
			fmt.Println("Error writing to output file:", err)
		}

		// Calculate download time and file size
		downloadTime := time.Since(startTime)
		totalDownloadTime += downloadTime

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Println("Error getting file info:", err)
		} else {
			totalFileSize += fileInfo.Size()
		}

		fmt.Printf("Downloaded and saved %s\n", url)
	}

	fmt.Printf("Total download time: %s\n", totalDownloadTime.Truncate(time.Second).String())
	fmt.Printf("Total file size: %d bytes\n", totalFileSize)
}

func readFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
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
