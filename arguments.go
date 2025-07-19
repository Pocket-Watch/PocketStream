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

type Arguments struct {
	Token           string
	RtmpSource      string
	Destination     string
	SegmentDuration string
}

func (a Arguments) Validate() {
	if a.Token == "" {
		fmt.Println("ERROR No token specified!")
		os.Exit(1)
	}

	if a.RtmpSource == "" {
		fmt.Println("ERROR No RTMP source specified!")
		os.Exit(1)
	}

	if _, err := url.Parse(a.RtmpSource); err != nil {
		fmt.Println("ERROR Invalid RTMP source URL ", err)
		os.Exit(1)
	}

	if a.Destination == "" {
		fmt.Println("ERROR No destination specified!")
		os.Exit(1)
	}

	if _, err := url.Parse(a.Destination); err != nil {
		fmt.Println("ERROR Invalid destination URL ", err)
		os.Exit(1)
	}

	duration, err := strconv.Atoi(a.SegmentDuration)
	if err != nil {
		fmt.Println("ERROR Invalid segment duration: ", err)
		os.Exit(1)
	}
	if duration < 0 {
		fmt.Println("ERROR Invalid (negative) segment duration: ", err)
		os.Exit(1)
	}
}

const DEFAULT_RTMP_SOURCE = "localhost:9000"
const DEFAULT_SEGMENT_DURATION = "2"

func Parse(args []string) Arguments {
	if len(args) == 0 {
		PrintHelp()
		os.Exit(0)
	}
	arguments := Arguments{
		Token:           "",
		RtmpSource:      DEFAULT_RTMP_SOURCE,
		Destination:     "",
		SegmentDuration: DEFAULT_SEGMENT_DURATION,
	}

	for i := 0; i < len(args); i++ {
		if slices.Contains(HELP_FLAGS, args[i]) {
			PrintHelp()
			os.Exit(0)
		}

		if index := slices.Index(SOURCE_FLAGS, args[i]); index != -1 {
			if i+1 < len(args) {
				arguments.RtmpSource = args[i+1]
				i++
			} else {
				fmt.Println("Expected RTMP source.")
				os.Exit(1)
			}
			continue
		}

		if index := slices.Index(DESTINATION_FLAGS, args[i]); index != -1 {
			if i+1 < len(args) {
				arguments.Destination = args[i+1]
				i++
			} else {
				fmt.Println("Expected destination URL.")
				os.Exit(1)
			}
			continue
		}

		if index := slices.Index(TOKEN_FLAGS, args[i]); index != -1 {
			if i+1 < len(args) {
				arguments.Token = args[i+1]
				i++
			} else {
				fmt.Println("Expected authorization token.")
				os.Exit(1)
			}
			continue
		}

		if index := slices.Index(SEGMENT_DURATION_FLAGS, args[i]); index != -1 {
			if i+1 < len(args) {
				arguments.SegmentDuration = args[i+1]
				i++
			} else {
				fmt.Println("Expected segment duration.")
				os.Exit(1)
			}
			continue
		}
		fmt.Println("[WARNING] Unrecognized flag/argument:", args[i])

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
	fmt.Println("    -h, -help, --help                    Display this help message")
	fmt.Println()
	fmt.Println("FFmpeg dependency is necessary.")
	fmt.Println("Specifying ports is optional")
	fmt.Println()
	fmt.Println("Usage example:")
	fmt.Printf("  %v -t OBHWYICqacQK2yFQGdQNe72O752SBVti3sU5w-Ri8KM= --dest https://example.com", exec)
	fmt.Println()
}
