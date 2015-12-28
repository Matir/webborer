package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"os"
)

// This is the result emitted by the worker for each URL tested.
type Result struct {
	// URL of resource
	URL *url.URL
	// HTTP Status Code
	Code int
	// Error if one occurred
	Error error
	// Redirect URL
	Redir *url.URL
	// Content length
	Length int64
}

// ResultsManager provides an interface for reading results from a channel and
// writing them to some form of output.
type ResultsManager interface {
	// Run reads all of the Results in the given channel and writes them to an
	// appropriate output sync.  Run should start its own goroutine for the bulk
	// of the work.
	Run(<-chan Result)
	// Wait until the channel has been read and output done.
	Wait()
}

type baseResultsManager struct {
	finished chan bool
}

// PlainResultsManager is designed to output a very basic output that is good
// for human reading, but not so good for machine parsing.  This is the default
// output and provides a decent way to review results on-screen.
type PlainResultsManager struct {
	baseResultsManager
	writer io.Writer
	fp     *os.File
	redirs bool
}

// CSVResultsManager writes a CSV containing all of the results.
type CSVResultsManager struct {
	baseResultsManager
	writer *csv.Writer
	fp     *os.File
}

// HTMLResultsManager writes an HTML file containing the results.
type HTMLResultsManager struct {
	baseResultsManager
	writer  io.Writer
	fp      *os.File
	BaseURL string
}

// Available output formats as strings.
var OutputFormats = []string{"text", "csv", "html"}

// Construct a ResultsManager for the given settings in the ScanSettings.
// Returns an object satisfying the ResultsManager interface or an error.
func GetResultsManager(settings *ScanSettings) (ResultsManager, error) {
	var writer io.Writer
	var fp *os.File
	var err error

	format := settings.OutputFormat
	if settings.OutputPath == "" {
		writer = os.Stdout
	} else {
		if fp, err = os.Create(settings.OutputPath); err != nil {
			return nil, err
		} else {
			writer = fp
		}
	}
	switch {
	case format == "text":
		return &PlainResultsManager{writer: writer, fp: fp, redirs: settings.IncludeRedirects}, nil
	case format == "csv":
		return &CSVResultsManager{writer: csv.NewWriter(writer), fp: fp}, nil
	case format == "html":
		return &HTMLResultsManager{writer: writer, fp: fp, BaseURL: settings.BaseURL}, nil
	}
	return nil, fmt.Errorf("Invalid output type: %s", format)
}

func (b *baseResultsManager) start() {
	b.finished = make(chan bool)
}

func (b *baseResultsManager) done() {
	b.finished <- true
}

func (b *baseResultsManager) Wait() {
	<-b.finished
}

func (rm *PlainResultsManager) Run(res <-chan Result) {
	go func() {
		rm.start()
		defer func() {
			if rm.fp != nil {
				rm.fp.Close()
			}
			rm.done()
		}()

		for r := range res {
			if !FoundSomething(r.Code) {
				continue
			}
			if r.Redir == nil {
				fmt.Fprintf(rm.writer, "%d %s (%d bytes)\n", r.Code, r.URL.String(), r.Length)
			} else {
				fmt.Fprintf(rm.writer, "%d %s -> %s\n", r.Code, r.URL.String(), r.Redir.String())
			}
		}
	}()
}

func (rm *CSVResultsManager) Run(res <-chan Result) {
	go func() {
		rm.start()
		defer func() {
			rm.writer.Flush()
			if rm.fp != nil {
				rm.fp.Close()
			}
			rm.done()
		}()

		maybeString := func(u *url.URL) string {
			if u == nil {
				return ""
			}
			return u.String()
		}

		rm.writer.Write([]string{"code", "url", "content_length", "redirect_url"})

		for r := range res {
			if !FoundSomething(r.Code) {
				continue
			}
			record := []string{
				fmt.Sprintf("%d", r.Code),
				r.URL.String(),
				fmt.Sprintf("%d", r.Length),
				maybeString(r.Redir),
			}
			rm.writer.Write(record)
		}
	}()
}

func (rm *HTMLResultsManager) Run(res <-chan Result) {
	go func() {
		rm.start()
		rm.writeHeader()

		defer func() {
			rm.writeFooter()
			if rm.fp != nil {
				rm.fp.Close()
			}
			rm.done()
		}()

		for r := range res {
			if !FoundSomething(r.Code) {
				continue
			}
			if r.Redir != nil {
				continue
			}
			rm.writeResult(&r)
		}
	}()
}

func (rm *HTMLResultsManager) writeHeader() {
	header := `{{define "HEAD"}}<html><head><title>gobuster: {{.BaseURL}}</title></head><h2>Results for <a href="{{.BaseURL}}">{{.BaseURL}}</a></h2><table><tr><th>Code</th><th>URL</th><th>Size</th></tr>{{end}}`
	t, err := template.New("htmlResultsManager").Parse(header)
	if err != nil {
		Logf(LogWarning, "Error parsing a template: %s", err.Error())
	}
	data := struct {
		BaseURL string
	}{
		BaseURL: rm.BaseURL,
	}
	err = t.ExecuteTemplate(rm.writer, "HEAD", data)
	if err != nil {
		Logf(LogWarning, "Error writing template output: %s", err.Error())
	}
}

func (rm *HTMLResultsManager) writeFooter() {
	footer := `{{define "FOOTER"}}</table></html>{{end}}`
	t, err := template.New("htmlResultsManager").Parse(footer)
	if err != nil {
		Logf(LogWarning, "Error parsing a template: %s", err.Error())
	}
	err = t.ExecuteTemplate(rm.writer, "FOOTER", nil)
	if err != nil {
		Logf(LogWarning, "Error writing template output: %s", err.Error())
	}
}

func (rm *HTMLResultsManager) writeResult(res *Result) {
	// TODO: don't rebuild the template with each row
	tmpl := `{{define "ROW"}}<tr><td>{{.Code}}</td><td><a href="{{.URL.String}}">{{.URL.String}}</a></td><td>{{.Length}}</td></tr>{{end}}`
	t, err := template.New("htmlResultsManager").Parse(tmpl)
	if err != nil {
		Logf(LogWarning, "Error parsing a template: %s", err.Error())
	}
	err = t.ExecuteTemplate(rm.writer, "ROW", res)
	if err != nil {
		Logf(LogWarning, "Error writing template output: %s", err.Error())
	}
}
