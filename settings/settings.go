// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package settings provides a central interface to webborer settings.
package settings

import (
	"errors"
	"flag"
	"fmt"
	"github.com/Matir/webborer/logging"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

// ScanSettings store all of the settings for the running scan.  It's basically
// a mapping from command-line flags into a single struct that can be passed
// into setup functions to get the desired behavior.
type ScanSettings struct {
	// Starting point and scope of scan
	BaseURLs StringSliceFlag
	// Number of threads to run
	Threads int
	// Number of workers to run
	Workers int
	// Exclusions
	ExcludePaths StringSliceFlag
	// Proxies
	Proxies StringSliceFlag
	// Operating mode
	RunMode RunModeOption
	// Parse HTML for links?
	ParseHTML bool
	// Time to sleep between requests, per thread
	SleepTime time.Duration
	// Log file path
	LogfilePath string
	// Level of logging
	LogLevel string
	// Wordlist for scanning
	WordlistPath string
	// Extensions for mangling
	Extensions StringSliceFlag
	// Whether or not to mangle
	Mangle bool
	// How long should internal queues be sized
	QueueSize int
	// Timeout for network requests
	Timeout time.Duration
	// Output type
	OutputFormat string
	// Output path
	OutputPath string
	// User-Agent for requests
	UserAgent string
	// Whether to include redirects in reporting
	IncludeRedirects bool
	// How to handle Robots.txt
	RobotsMode RobotsModeOption
	// Whether to allow upgrade from http to https
	AllowHTTPSUpgrade bool
	// Spider which http response codes
	SpiderCodes IntSliceFlag
	// HTTP Auth Username
	HTTPUsername string
	// HTTP Auth Password
	HTTPPassword string
	// Headers *always* sent
	Header HeaderFlag
	// Headers sometimes sent
	OptionalHeader HeaderFlag
	// Progress bar
	ProgressBar bool
	// Whether or not to do CPU Profiling
	DebugCPUProf bool
	// Config file used when loading (for debugging only)
	configPath string
	// Have flags been set up?
	flagsSet bool
}

var DefaultUserAgent = "WebBorer 0.01"
var outputFormats []string

// Constructs a ScanSettings struct with all of the defaults to be used.
func NewScanSettings() *ScanSettings {
	settings := &ScanSettings{
		Threads:        runtime.NumCPU(),
		Extensions:     []string{"html", "php", "asp", "aspx"},
		Mangle:         true,
		QueueSize:      1024,
		Timeout:        30 * time.Second,
		LogLevel:       "WARNING",
		SpiderCodes:    IntSliceFlag{200},
		ProgressBar:    true,
		RunMode:        RunModeEnumeration,
		Header:         make(HeaderFlag),
		OptionalHeader: make(HeaderFlag),
	}
	settings.InitFlags()
	return settings
}

// Create settings that includes configuration files and command line flags.
// Generally, this should be called very early and is the best way to get the
// settings.
func GetScanSettings() (*ScanSettings, error) {
	settings := NewScanSettings()
	settings.LoadFromDefaultConfigFiles()
	settings.ParseFlags()
	if err := settings.Validate(); err != nil {
		return nil, err
	}
	return settings, nil
}

// Setup all of the flags.  Should be called *early*
func (settings *ScanSettings) InitFlags() {
	if settings.flagsSet {
		return
	}

	flag.Var(&settings.BaseURLs, "url", "Starting `URL` & scopes.")
	runModeHelp := fmt.Sprintf("Run `mode`. Options: [%s]", strings.Join(runModeStrings[:], ", "))
	flag.Var(&settings.RunMode, "mode", runModeHelp)
	flag.IntVar(&settings.Threads, "threads", runtime.NumCPU(), "Number of worker `threads`.")
	flag.IntVar(&settings.Workers, "workers", runtime.NumCPU()*2, "Number of `workers`.")
	flag.Var(&settings.ExcludePaths, "exclude", "List of `paths` to exclude from search.")
	flag.BoolVar(&settings.ParseHTML, "html", true, "Parse HTML documents for links to follow.")
	flag.BoolVar(&settings.AllowHTTPSUpgrade, "allow-upgrade", false, "Allow HTTP->HTTPS upgrades.")
	sleepTimeValue := DurationFlag{&settings.SleepTime}
	flag.Var(sleepTimeValue, "sleep", "Time (as `duration`) to sleep between requests.")
	flag.StringVar(&settings.LogfilePath, "logfile", "", "Logfile `filename` (defaults to stderr)")
	flag.StringVar(&settings.WordlistPath, "wordlist", "", "Wordlist `filename` to use (default built-in)")
	flag.Var(&settings.Extensions, "extensions", "List of `extensions` to mangle with.")
	flag.BoolVar(&settings.Mangle, "mangle", true, "Mangle by adding extensions.")
	flag.Var(&settings.Header, "header", "Headers to send with each request.")
	flag.Var(&settings.OptionalHeader, "optional-header", "Headers to try sending one at a time.")
	flag.Var(&settings.Proxies, "proxy", "Proxy or `proxies` to use.")
	timeoutValue := DurationFlag{&settings.Timeout}
	flag.Var(timeoutValue, "timeout", "Network connection timeout (`duration`).")
	if len(outputFormats) > 1 {
		formatHelp := fmt.Sprintf("Output `format`.  Options: [%s]", strings.Join(outputFormats, ", "))
		flag.StringVar(&settings.OutputFormat, "format", outputFormats[0], formatHelp)
	}
	flag.StringVar(&settings.OutputPath, "outfile", "", "Output `file`, defaults to stdout.")
	loglevelHelp := fmt.Sprintf("Log `level`.  Options: [%s]", strings.Join(logging.LogLevelStrings[:], ", "))
	flag.StringVar(&settings.LogLevel, "loglevel", settings.LogLevel, loglevelHelp)
	flag.StringVar(&settings.UserAgent, "user-agent", DefaultUserAgent, "`User-Agent` for requests")
	flag.BoolVar(&settings.IncludeRedirects, "include-redirects", false, "Include redirects in reports.")
	flag.Var(&settings.SpiderCodes, "spider-codes", "HTTP Response Codes to Continue Spidering On.")
	robotsModeHelp := fmt.Sprintf("Robots `mode`.  Options: [%s]", strings.Join(robotsModeStrings[:], ", "))
	flag.Var(&settings.RobotsMode, "robots-mode", robotsModeHelp)
	flag.StringVar(&settings.HTTPUsername, "http-username", "", "Username to be used for HTTP Auth")
	flag.StringVar(&settings.HTTPPassword, "http-password", "", "Password to be used for HTTP Auth")
	flag.BoolVar(&settings.ProgressBar, "progress", true, "Display a progress bar on stderr.")

	// Debugging flags
	flag.BoolVar(&settings.DebugCPUProf, "debug-cpuprof", false, "[DEBUG] CPU Profiling")

	settings.flagsSet = true
}

// Load settings from the first file found in searchPaths
func (settings *ScanSettings) LoadFromDefaultConfigFiles() {
	for _, path := range defaultConfigPaths {
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				continue
			}
			settings.LoadFromConfigFile(path)
			return
		}
	}
}

// Load from the specified file
func (settings *ScanSettings) LoadFromConfigFile(path string) {
	settings.InitFlags()
	// TODO: load
	settings.configPath = path
}

// Parse command line flags into settings
func (settings *ScanSettings) ParseFlags() {
	settings.InitFlags()
	flag.Parse()
	for i := 0; i < flag.NArg(); i++ {
		settings.BaseURLs = append(settings.BaseURLs, flag.Arg(i))
	}
}

// Validate settings
func (settings *ScanSettings) Validate() error {
	flagError := func(str string) error {
		os.Stderr.WriteString("Usage:\n")
		flag.PrintDefaults()
		return errors.New(str)
	}
	if len(settings.BaseURLs) == 0 {
		return flagError("URL is required.")
	}
	return nil
}

// Printable config
func (settings *ScanSettings) String() string {
	flags := make([]string, 0)

	flag.VisitAll(func(f *flag.Flag) {
		flags = append(flags, fmt.Sprintf("-%s=%s", f.Name, f.Value.String()))
	})

	return strings.Join(flags, " ")
}

// Convert BaseURL strings to URLs
func (settings *ScanSettings) GetScopes() ([]*url.URL, error) {
	scopes := make([]*url.URL, len(settings.BaseURLs))
	for i, baseURL := range settings.BaseURLs {
		parsed, err := url.Parse(baseURL)
		scopes[i] = parsed
		if err != nil {
			return nil, fmt.Errorf("Unable to parse BaseURL (%s): %s", baseURL, err.Error())
		}
		if scopes[i].Path == "" {
			scopes[i].Path = "/"
		}
		logging.Logf(logging.LogDebug, "Added BaseURL: %s", scopes[i].String())
	}
	return scopes, nil
}

// Init output formats
func SetOutputFormats(formats []string) {
	outputFormats = formats
}
