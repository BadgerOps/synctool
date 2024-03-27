package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)

type DownloadProgress struct {
	totalBytes      int64
	downloadedBytes int64
	startTime       time.Time
	mu              sync.Mutex
}

func (dp *DownloadProgress) AddDownloadedBytes(n int) {
	fmt.Println("Adding downloaded bytes: ", int64(n))
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.downloadedBytes += int64(n)
}

func (dp *DownloadProgress) DownloadRate() float64 {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	duration := time.Since(dp.startTime).Seconds()
	fmt.Println("current duration: ", duration)
	fmt.Println("Current bytes: ", dp.downloadedBytes)
	return float64(dp.downloadedBytes) / duration
}

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

			dp := &DownloadProgress{
				startTime:       time.Now(),
				downloadedBytes: 0,
			}

			go func() {
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()
				for range ticker.C {
					rate := dp.DownloadRate()
					fmt.Printf("Current download rate: %.2f bytes/sec\n", rate)
				}
			}()

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

	for _, url := range rawFile {
		//	startTime := time.Now()
		totalDownloadTime := time.Duration(0)
		totalFileSize := int64(0)
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

		dp := &DownloadProgress{
			totalBytes: resp.ContentLength,
			startTime:  time.Now(),
		}

		fmt.Println("Total size is: ", resp.ContentLength)
		// Create the output file
		filePath := filepath.Join(outDir, path.Base(url))
		out, err := os.Create(filePath)
		if err != nil {
			fmt.Println("Error creating output file:", err)
			continue
		}
		defer out.Close()

		// Read the response body in chunks, write each chunk to the file, and update dp.downloadedBytes
		buf := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Println("Error reading response body:", err)
				break
			}
			if n == 0 {
				break
			}

			if _, err := out.Write(buf[:n]); err != nil {
				fmt.Println("Error writing to output file:", err)
				break
			}
			dp.AddDownloadedBytes(n)
			//fmt.Println("downloading the resp body 1024k at a time...", n)

			//dp.downloadedBytes += int64(n)
			//fmt.Println("Downloaded: ", dp.downloadedBytes)
			//fmt.Printf("Downloaded and saved %s\n", url)
		}

		fmt.Printf("Total download time: %s\n", totalDownloadTime.Truncate(time.Second).String())
		fmt.Printf("Total file size: %d bytes\n", totalFileSize)
	}
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
