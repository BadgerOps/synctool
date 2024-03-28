package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// set up log formatting and log level, I want actual timestamps
func init() {
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logrus.SetFormatter(formatter)
}

// Use mutex to help mitigate collisions, I followed along with https://gobyexample.com/mutexes
// Will add threading support next

type DownloadProgress struct {
	totalBytes      int64
	downloadedBytes int64
	startTime       time.Time
	mu              sync.Mutex
}

func (dp *DownloadProgress) AddDownloadedBytes(n int) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.downloadedBytes += int64(n)
	//logrus.Trace("Downloaded bytes now equal to: ", dp.downloadedBytes)
}

func (dp *DownloadProgress) DownloadRate() float64 {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	duration := time.Since(dp.startTime).Seconds()
	logrus.Trace("current duration: ", duration)
	logrus.Trace("Current bytes: ", dp.downloadedBytes)
	return float64(dp.downloadedBytes) / duration
}

func getURL(rawFile []string, outDir string, dp *DownloadProgress) {

	for _, url := range rawFile {
		totalDownloadTime := time.Duration(0)
		startTime := time.Now()
		logrus.Info("Downloading file from URL: ", url)

		resp, err := http.Get(url)
		if err != nil {
			logrus.Error("Error downloading file:", url)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logrus.Error("Error: received non-200 status code from server for file:", url)
			continue
		}

		logrus.Trace("Total upstream file size is: ", resp.ContentLength)
		dp.totalBytes = resp.ContentLength
		// Create the output file
		filePath := filepath.Join(outDir, path.Base(url))
		out, err := os.Create(filePath)
		if err != nil {
			logrus.Error("Error creating output file:", err)
			continue
		}
		defer out.Close()

		/*
			Read the response body in chunks
			write each chunk to output file
			update dp.downloadedBytes
		*/
		buf := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if err != nil && err != io.EOF {
				logrus.Error("Error reading response body:", err)
				break
			}
			if n == 0 {
				break
			}

			if _, err := out.Write(buf[:n]); err != nil {
				logrus.Error("Error writing to output file:", err)
				break
			}
			dp.AddDownloadedBytes(n)
		}
		downloadTime := time.Since(startTime)
		totalDownloadTime += downloadTime
		logrus.Infof("Total download time: %s\n", totalDownloadTime.Truncate(time.Second).String())
		logrus.Infof("Total file size: %s\n", bytesToReadable(dp.totalBytes))
	}
}

// https://gist.github.com/chadleeshaw/5420caa98498c46a84ce94cd9655287a convert bytes to human readable number
func bytesToReadable(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
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

// Start up main func
func main() {
	var logLevel string
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
			&cli.StringFlag{
				Name:        "loglevel",
				Value:       "info",
				Usage:       "Set log level (debug, info, warn, error, fatal, panic)",
				Destination: &logLevel,
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
				logrus.Error("Error reading input file:", err)
				os.Exit(1)
			}
			level, err := logrus.ParseLevel(c.String("loglevel"))
			if err != nil {
				return err
			}
			logrus.SetLevel(level)

			dp := &DownloadProgress{
				startTime:       time.Now(),
				downloadedBytes: 0,
			}

			go func() {
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()
				for range ticker.C {
					rate := int64(math.Round(dp.DownloadRate()))
					logrus.Info("Current download rate: ", bytesToReadable(rate))
				}
			}()

			getURL(rawFile, outDir, dp)
			logrus.Info("Downloaded all files listed in:", c.String("file"))
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		cli.Exit(err.Error(), 1)
		os.Exit(1)
	}
}
