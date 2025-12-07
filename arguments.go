package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
)

var HELP_FLAGS = []string{"-h", "-help", "--help"}
var SOURCE_FLAGS = []string{"-src", "--source"}
var TOKEN_FLAGS = []string{"-t", "--token"}
var DESTINATION_FLAGS = []string{"-d", "--dest", "--domain"}
var SEGMENT_DURATION_FLAGS = []string{"-s", "--segment", "--seg"}
var OUTPUT_DIRECTORY_FLAGS = []string{"-o", "--out"}
var FFMPEG_UPLOAD_FLAGS = []string{"-fu", "--ff-upload"}
var SAVE_ERRORS_FLAGS = []string{"--save-errors"}

type Arguments struct {
	Token           string
	RtmpSource      string
	Destination     string
	SegmentDuration string
	OutputDirectory string
	FFmpegUpload    bool
	SaveErrors      bool
}

func (a Arguments) Validate() {
	if a.Token == "" {
		LogError(&a, "No token specified!")
		os.Exit(1)
	}

	if a.RtmpSource == "" {
		LogError(&a, "No RTMP source specified")
		os.Exit(1)
	}

	if _, err := url.Parse(a.RtmpSource); err != nil {
		LogError(&a, "Invalid RTMP source URL: %v", err)
		os.Exit(1)
	}

	if a.Destination == "" {
		LogError(&a, "No destination specified!")
		os.Exit(1)
	}

	if _, err := url.Parse(a.Destination); err != nil {
		LogError(&a, "Invalid destination URL: %v", err)
		os.Exit(1)
	}

	duration, err := strconv.ParseFloat(a.SegmentDuration, 64)
	if err != nil {
		LogError(&a, "Invalid segment duration: %v", err)
		os.Exit(1)
	}
	if duration < 0 {
		LogError(&a, "Invalid (negative) segment duration: %v", err)
		os.Exit(1)
	}
}

const DEFAULT_RTMP_SOURCE = "localhost:9000"
const DEFAULT_SEGMENT_DURATION = "2"
const DEFAULT_OUTPUT_DIRECTORY = "stream"
const DEFAULT_ERROR_FILE = "stream-error-log.txt"

func Parse(args []string) Arguments {
	if len(args) == 0 {
		PrintHelp()
		os.Exit(0)
	}
	arguments := Arguments{
		RtmpSource:      DEFAULT_RTMP_SOURCE,
		SegmentDuration: DEFAULT_SEGMENT_DURATION,
		OutputDirectory: DEFAULT_OUTPUT_DIRECTORY,
	}

	for i := 0; i < len(args); i++ {
		if slices.Contains(HELP_FLAGS, args[i]) {
			PrintHelp()
			os.Exit(0)
		}

		if slices.Contains(FFMPEG_UPLOAD_FLAGS, args[i]) {
			arguments.FFmpegUpload = true
			continue
		}

		if slices.Contains(SAVE_ERRORS_FLAGS, args[i]) {
			arguments.SaveErrors = true
			continue
		}

		if index := slices.Index(SOURCE_FLAGS, args[i]); index != -1 {
			i++
			if i < len(args) {
				arguments.RtmpSource = args[i]
			} else {
				LogError(&arguments, "Expected RTMP source.")
				os.Exit(1)
			}
			continue
		}

		if index := slices.Index(DESTINATION_FLAGS, args[i]); index != -1 {
			i++
			if i < len(args) {
				arguments.Destination = args[i]
			} else {
				LogError(&arguments, "Expected destination URL.")
				os.Exit(1)
			}
			continue
		}

		if index := slices.Index(TOKEN_FLAGS, args[i]); index != -1 {
			i++
			if i < len(args) {
				arguments.Token = args[i]
			} else {
				LogError(&arguments, "Expected authorization token.")
				os.Exit(1)
			}
			continue
		}

		if index := slices.Index(SEGMENT_DURATION_FLAGS, args[i]); index != -1 {
			i++
			if i < len(args) {
				arguments.SegmentDuration = args[i]
			} else {
				LogError(&arguments, "Expected segment duration.")
				os.Exit(1)
			}
			continue
		}

		if index := slices.Index(OUTPUT_DIRECTORY_FLAGS, args[i]); index != -1 {
			i++
			if i < len(args) {
				arguments.OutputDirectory = args[i]
			} else {
				LogError(&arguments, "Expected output directory.")
				os.Exit(1)
			}
			continue
		}

		fmt.Println("[WARN ] Unrecognized flag/argument:", args[i])
	}
	return arguments
}

func PrintHelp() {
	exec, err := os.Executable()
	if err != nil {
		exec = os.Args[0]
	}
	exec = filepath.Base(exec)

	fmt.Println("PocketStream - live")
	fmt.Println()
	fmt.Printf("Usage: %v [arguments...]\n", exec)
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("    -t, --token    [base64]              Authorization token to be passed in the headers")
	fmt.Println("    -src, --source [host:port]           Rtmp source stream address (default: " + DEFAULT_RTMP_SOURCE + ")")
	fmt.Println("    -d, --dest     [scheme://host:port]  Destination domain where the server is running on")
	fmt.Println("    -s, --segment  [seconds]             Segment duration in seconds (default: " + DEFAULT_SEGMENT_DURATION + ")")
	fmt.Println("    -o, --out      [directory]           Directory for HLS chunks (not needed if --ff-upload is used")
	fmt.Println("    -fu, --ff-upload                     Use ffmpeg to upload segments and playlists (if not behind proxy)")
	fmt.Println("    --save-errors                        Save errors to " + DEFAULT_ERROR_FILE + " (useful when used behind UI)")
	fmt.Println("    -h, -help, --help                    Display this help message")
	fmt.Println()
	fmt.Println("FFmpeg dependency is necessary.")
	fmt.Println("Specifying ports is optional. Segment duration can be given as a decimal")
	fmt.Println()
	fmt.Println("Usage example:")
	fmt.Printf("  %v -t OBHWYICqacQK2yFQGdQNe72O752SBVti3sU5w-Ri8KM= --dest https://example.com", exec)
	fmt.Println()
}

func LogError(arguments *Arguments, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	output := fmt.Sprintf("[ERROR] %v\n", message)

	fmt.Print(output)

	if arguments.SaveErrors {
		logFile := DEFAULT_ERROR_FILE
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()

		if _, err := f.WriteString(output); err != nil {
			fmt.Println(err)
			return
		}
	}
}
