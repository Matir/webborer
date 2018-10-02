package results

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"os"
	"sort"
)

// Check results for broken links.
type LinkCheckResultsManager struct {
	baseResultsManager
	writer     io.Writer
	fp         *os.File
	format     string
	resMap     map[string]*Result
	missing    int
	writerImpl linkCheckWriter
}

func (rm *LinkCheckResultsManager) init() error {
	rm.resMap = make(map[string]*Result)
	switch rm.format {
	case "text":
		rm.format = "csv"
		fallthrough
	case "csv":
		rm.writerImpl = newLinkCheckCSVWriter(rm.writer)
	case "html":
		rm.writerImpl = newLinkCheckHTMLWriter(rm.writer)
	default:
		return fmt.Errorf("Invalid format: %s", rm.format)
	}
	return nil
}

func (rm *LinkCheckResultsManager) Run(resChan <-chan *Result) {
	rm.start()
	go func() {
		defer func() {
			rm.writerImpl.flush()
			if rm.fp != nil {
				rm.fp.Close()
			}
			rm.done()
		}()
		var keys []string
		for res := range resChan {
			key := res.URL.String()
			rm.resMap[key] = res
			keys = append(keys, key)
		}
		sort.Strings(keys)
		rm.writerImpl.writeHeader()
		count := 0
		for _, resKey := range keys {
			groupWritten := false
			res := rm.resMap[resKey]
			for k, t := range res.Links {
				if rm.linkIsBroken(k) {
					if !groupWritten {
						groupWritten = true
						rm.writerImpl.writeGroup(k)
					}
					rm.writerImpl.writeBrokenLink(resKey, k, LinkTypes[t])
					count++
				}
			}
		}
		rm.writerImpl.writeFooter(count)
	}()
}

// Check if an HTTP code is broken, consider all 400/500s
func codeIsBroken(code int) bool {
	return code >= 400
}

func (rm *LinkCheckResultsManager) linkIsBroken(url string) bool {
	if r, ok := rm.resMap[url]; !ok {
		rm.missing++
		return false
	} else {
		return codeIsBroken(r.Code)
	}
}

type linkCheckWriter interface {
	writeHeader()
	writeFooter(count int)
	writeGroup(src string)
	writeBrokenLink(src, dst, ltype string)
	flush()
}

// Write link check output in CSV
type linkCheckCSVWriter struct {
	csvWriter *csv.Writer
}

func newLinkCheckCSVWriter(writer io.Writer) *linkCheckCSVWriter {
	return &linkCheckCSVWriter{csv.NewWriter(writer)}
}

func (w *linkCheckCSVWriter) writeHeader() {
	w.csvWriter.Write([]string{"Source URL", "Destination URL", "Type"})
}

func (w *linkCheckCSVWriter) writeFooter(count int) {
	return
}

func (w *linkCheckCSVWriter) writeGroup(src string) {
	return
}

func (w *linkCheckCSVWriter) writeBrokenLink(src, dst, ltype string) {
	w.csvWriter.Write([]string{src, dst, ltype})
}

func (w *linkCheckCSVWriter) flush() {
	w.csvWriter.Flush()
}

// Write link check output in HTML
type linkCheckHTMLWriter struct {
	writer io.Writer
}

func newLinkCheckHTMLWriter(writer io.Writer) *linkCheckHTMLWriter {
	return &linkCheckHTMLWriter{writer}
}

func (w *linkCheckHTMLWriter) writeHeader() {
}

func (w *linkCheckHTMLWriter) writeFooter(count int) {
}

func (w *linkCheckHTMLWriter) writeGroup(src string) {
}

func (w *linkCheckHTMLWriter) writeBrokenLink(src, dst, ltype string) {
}

func (w *linkCheckHTMLWriter) flush() {
}
