package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// Output locally ffmpeg -listen 1 -i rtmp://localhost:9000 -codec: copy -hls_time 3 -hls_list_size 0 -f hls live.m3u8
// Test: ./pocketstream -t OBHWYICqacQK2yFQGdQNe72O752SBVti3sU5w-Ri8KM= -d localhost:1234

const StreamUploadEndpoint = "/watch/api/stream/upload/stream.m3u8"
const StreamStartEndpoint = "/watch/api/stream/start"

func main() {
	args := Parse(os.Args[1:])
	args.Validate()
	fmt.Println(args)

	if strings.HasSuffix(args.Destination, "/") {
		args.Destination = args.Destination[:len(args.Destination)-1]
	}

	destination := args.Destination + StreamUploadEndpoint
	// Headers have CRLF suffix to prevent FFmpeg warning for every request
	ffArgs := []string{
		"-listen", "1",
		"-i", "rtmp://" + args.RtmpSource,
		"-codec:", "copy",
		"-f", "hls",
		"-headers", "Authorization: " + args.Token + "\r\n",
		"-method", "POST",
		"-hls_time", args.SegmentDuration,
		"-hls_list_size", "0",
		destination,
	}

	fmt.Println("Starting stream, informing the server!")
	startStream(&args)
	for {
		cmd := exec.Command("ffmpeg", ffArgs...)
		fmt.Println("Executing FFmpeg command!")
		executeCommand(cmd)
	}
}

func executeCommand(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		return
	}

	/*	// Read both outputs simultaneously
		go func() {
			scanner := bufio.NewScanner(os.Stdout)
			for scanner.Scan() {
				line := scanner.Text()
				fmt.Println(line)
			}
		}()

		go func() {
			scanner := bufio.NewScanner(os.Stderr)
			for scanner.Scan() {
				line := scanner.Text()
				fmt.Println(line)
			}
		}()*/

	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
	}
}

func startStream(args *Arguments) {
	req, err := http.NewRequest("POST", args.Destination+StreamStartEndpoint, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("Authorization", args.Token)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Server responded with status code: " + resp.Status)
		errBody, err := io.ReadAll(resp.Body)
		var bodyError = ""
		if err == nil {
			bodyError = string(errBody)
		}
		fmt.Println(bodyError)
		os.Exit(1)
	}
}
