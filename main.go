package main

import (
	"bufio"
	"context"
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
type FileProgress struct {
	totalBytes      int64
	downloadedBytes int64
	startTime       time.Time
}

type DownloadProgress struct {
	progress map[string]*FileProgress
	mu       sync.Mutex
}

func (dp *DownloadProgress) AddDownloadedBytes(url string, n int) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	if _, ok := dp.progress[url]; !ok {
		dp.progress[url] = &FileProgress{
			totalBytes:      0,
			downloadedBytes: 0,
			startTime:       time.Now(),
		}
	}
	dp.progress[url].downloadedBytes += int64(n)
	//logrus.Trace("Downloaded bytes now equal to: ", dp.progress[url].downloadedBytes)
}

func (dp *DownloadProgress) DownloadRate(url string) float64 {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	if _, ok := dp.progress[url]; !ok {
		return 0
	}
	duration := time.Since(dp.progress[url].startTime).Seconds()
	logrus.Trace("current duration: ", duration)
	logrus.Trace("Current bytes: ", dp.progress[url].downloadedBytes)
	return float64(dp.progress[url].downloadedBytes) / duration
}

// main function to download URL's
// TODO: break this up a little more would probably make sense...
func getURL(url string, outDir string, threadID int) {
	dp := &DownloadProgress{
		progress: make(map[string]*FileProgress),
	}
	totalDownloadTime := time.Duration(0)
	startTime := time.Now()
	// set up context and exit (cancel) for the monitorDownload goroutine
	ctx, cancel := context.WithCancel(context.Background())
	// start up our monitoring for this thread
	monitorDownload(ctx, dp, url, threadID)
	logrus.Info("Downloading file from URL: ", url)

	resp, err := http.Get(url)
	if err != nil {
		logrus.Error("Error downloading file from URL:", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Error("Error: received non-200 status code from server for file:", url)
	}

	logrus.Tracef("Total upstream file size for %s is: %d", url, resp.ContentLength)
	// Create the output file
	filePath := filepath.Join(outDir, path.Base(url))
	out, err := os.Create(filePath)
	if err != nil {
		logrus.Error("Error creating output file:", err)
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
		dp.AddDownloadedBytes(url, n)
	}
	downloadTime := time.Since(startTime)
	totalDownloadTime += downloadTime
	// now exit the context when we're done so we stop logging download rate
	cancel()
	logrus.Infof("Total download time for url %s: %s\n", url, totalDownloadTime.Truncate(time.Second).String())
	logrus.Infof("Total file size for url %s: %s\n", url, bytesToReadable(int64(dp.progress[url].totalBytes)))
}

func monitorDownload(ctx context.Context, dp *DownloadProgress, url string, threadID int) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				// If the context is cancelled, stop the goroutine
				return
			case <-ticker.C:
				duration := time.Since(dp.progress[url].startTime).Seconds()
				rate := math.Round(float64(dp.progress[url].downloadedBytes))
				logrus.WithFields(logrus.Fields{
					"threadID": threadID,
				}).Infof("Current download rate for: %s is %s/s", url, bytesToReadable(int64(rate/duration)))
			}
		}
	}()
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
			&cli.IntFlag{
				Name:  "threads",
				Value: 4,
				Usage: "Number of download threads",
			},
		},
		Action: func(c *cli.Context) error {
			urlChan := make(chan string)
			var wg sync.WaitGroup
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

			for i := 0; i < c.Int("threads"); i++ {
				wg.Add(1)
				go func(threadID int) {
					defer wg.Done()
					for url := range urlChan {
						getURL(url, outDir, threadID)
						if err != nil {
							logrus.WithFields(logrus.Fields{
								"threadID": threadID,
							}).Error(err)
						}
					}
				}(i)
			}

			for _, url := range rawFile {
				urlChan <- url
			}
			close(urlChan)
			wg.Wait()
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
