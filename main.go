package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cakturk/go-netstat/netstat"
)

// Output locally ffmpeg -listen 1 -i rtmp://localhost:9000 -codec: copy -hls_time 3 -hls_list_size 0 -f hls live.m3u8
// Test: ./pocketstream -t OBHWYICqacQK2yFQGdQNe72O752SBVti3sU5w-Ri8KM= -d localhost:1234

const StreamUploadEndpoint = "/api/stream/upload/stream.m3u8"
const StreamUpload = "/api/stream/upload/"
const StreamStartEndpoint = "/api/stream/start"
const PlaylistName = "stream.m3u8"
const M3U8ContentType = "application/vnd.apple.mpegurl"

var client = http.Client{}

func main() {
	args := Parse(os.Args[1:])
	args.Validate()
	fmt.Println(args)

	if strings.HasSuffix(args.Destination, "/") {
		args.Destination = args.Destination[:len(args.Destination)-1]
	}

	destination := args.Destination + StreamUploadEndpoint

	var ffArgs []string
	if args.FFmpegUpload {
		ffArgs = ffmpegUploadArgs(args, destination)
	} else {
		ffArgs = ffmpegFileArgs(args)
	}

	fmt.Println("Starting stream, informing the server!")
	startStream(&args)

	cmd := exec.Command("ffmpeg", ffArgs...)
	fmt.Println("Executing FFmpeg command:", cmd.String())

	if args.FFmpegUpload {
		executeCommandStdPipe(&args, cmd)
		return
	}

	handleStreaming(&args, cmd)
}

func handleStreaming(args *Arguments, cmd *exec.Cmd) {
	if err := os.MkdirAll(args.OutputDirectory, 0o755); err != nil {
		fmt.Println("Error creating output directory:", err)
		os.Exit(1)
	}

	_, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	if err := cmd.Start(); err != nil {
		return
	}

	checkPortClaimedPeriodically(args.RtmpSource, 250*time.Millisecond, 10)

	// FFmpeg writes only to STDERR
	go func() {
		lastPath := ""
		scanner := bufio.NewScanner(stdErr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			index := strings.Index(line, "[hls @")
			from := index + 6
			if index == -1 || !strings.Contains(line[from:], "Opening") {
				continue
			}
			path := parseHlsPath(line, from)
			if lastPath == "" {
				lastPath = path
				continue
			}
			go uploadRequest(args, lastPath)
			lastPath = path
		}
	}()

	fmt.Println("PocketStream is ready")

	if err := cmd.Wait(); err != nil {
		fmt.Println("QUIT", err)
	}
}

func parseHlsPath(line string, from int) string {
	line = line[from:]
	apostrophe1 := strings.Index(line, "'")
	path := line[apostrophe1+1:]
	apostrophe2 := strings.Index(path, "'")
	path = path[:apostrophe2]
	return strings.TrimSuffix(path, ".tmp")
}

func uploadRequest(args *Arguments, path string) {
	destination := args.Destination + StreamUpload + filepath.Base(path)

	if !fileExists(path) {
		fmt.Println("ERROR: File at", path, "doesn't exist!")
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Uploading", path, "of size", len(data))
	reader := bytes.NewReader(data)
	req, err := http.NewRequest("POST", destination, reader)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", args.Token)
	if strings.HasSuffix(path, ".m3u8") {
		req.Header.Set("Content-Type", M3U8ContentType)
	}
	response, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	fmt.Println("Status:", response.Status)
}

func ffmpegUploadArgs(args Arguments, destination string) []string {
	// Headers have CRLF suffix to prevent FFmpeg warning for every request
	return []string{
		"-listen", "1",
		"-i", "rtmp://" + args.RtmpSource,
		"-c", "copy",
		"-f", "hls",
		"-headers", "Authorization: " + args.Token + "\r\n",
		"-method", "POST",
		"-hls_time", args.SegmentDuration,
		"-hls_list_size", "0",
		destination,
	}
}

func ffmpegFileArgs(args Arguments) []string {
	return []string{
		"-listen", "1",
		"-i", "rtmp://" + args.RtmpSource,
		"-c", "copy",
		"-f", "hls",
		"-hls_time", args.SegmentDuration,
		"-hls_list_size", "0",
		filepath.Join(args.OutputDirectory, PlaylistName),
	}
}

func executeCommandStdPipe(args *Arguments, cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return
	}

	checkPortClaimedPeriodically(args.RtmpSource, 250*time.Millisecond, 10)
	fmt.Println("PocketStream is ready")

	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
	}
}

func checkPortClaimedPeriodically(address string, interval time.Duration, attempts int) {
	addressV4 := strings.Replace(address, "localhost", "127.0.0.1", 1)
	addressV6 := strings.Replace(address, "localhost", "::1", 1)
	start := time.Now()
	for i := 0; i < attempts; i++ {
		socketsV4, err := netstat.TCPSocks(func(s *netstat.SockTabEntry) bool {
			remoteV4 := s.RemoteAddr.String()
			return remoteV4 == addressV4
		})

		if err != nil {
			fmt.Println("[CHECK TCPv4 ERROR]", err)
			os.Exit(1)
		}

		socketsV6, err := netstat.TCP6Socks(func(s *netstat.SockTabEntry) bool {
			localV6 := s.LocalAddr.String()
			return localV6 == addressV6
		})

		if err != nil {
			fmt.Println("[CHECK TCPv6 ERROR]", err)
			os.Exit(1)
		}

		if len(socketsV6) > 0 || len(socketsV4) > 0 {
			fmt.Println("Confirmed server listening at", address, "after", time.Since(start).String())
			return
		}

		time.Sleep(interval)
	}

	fmt.Println("ERROR Failed to determine whether address", address, "is claimed after", attempts, "attempts")
}

func startStream(args *Arguments) {
	req, err := http.NewRequest("POST", args.Destination+StreamStartEndpoint, nil)
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

	req.Header.Set("Authorization", args.Token)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("ERROR Server responded with status code:", resp.Status)

		errBody, err := io.ReadAll(resp.Body)
		var bodyError = ""
		if err == nil {
			bodyError = string(errBody)
		}

		fmt.Println(bodyError)
		os.Exit(1)
	}
}

func toString(num int) string {
	return strconv.Itoa(num)
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
