package results

import (
	"bufio"
	"fmt"
	"github.com/Matir/webborer/util"
	"os"
	"strings"
)

var neverImportant = []string{
	"etag",
	"cache-control",
}

type BaselineResult struct {
	Result

	// Which properties are significant
	PathSignificant    bool
	HeadersSignificant []string
	CodeSignificant    bool
}

type DiffResultsManager struct {
	baselines map[string]*BaselineResult
	done      chan interface{}
	keep      map[string][]*Result
	fp        *os.File
}

func NewDiffResultsManager(fp *os.File) *DiffResultsManager {
	return &DiffResultsManager{
		baselines: make(map[string]*BaselineResult),
		done:      make(chan interface{}),
		keep:      make(map[string][]*Result),
		fp:        fp,
	}
}

func NewBaselineResult(results ...Result) (*BaselineResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("Need at least one result.")
	}

	res := &BaselineResult{
		Result:             results[0],
		PathSignificant:    true,
		HeadersSignificant: make([]string, 0),
		CodeSignificant:    true,
	}

	for i := 0; i < len(results)-1; i++ {
		a, b := results[i], results[i+1]
		if a.Code != b.Code {
			res.CodeSignificant = false
		}
		if a.URL.Path != b.URL.Path {
			res.PathSignificant = false
		}
	}

	for k, _ := range res.Header {
		k = strings.ToLower(k)
		if util.StringSliceContains(neverImportant, k) {
			continue
		}
		matches := true
		baseline := results[0].Header[k][0]
		if len(results) > 0 {
			for _, r := range results[1:] {
				if r.Header[k][0] != baseline {
					matches = false
					break
				}
			}
		}
		if matches {
			res.HeadersSignificant = append(res.HeadersSignificant, k)
		}
	}

	return res, nil
}

func (b *BaselineResult) Matches(a *Result) bool {
	if b.PathSignificant && b.URL.Path != a.URL.Path {
		return false
	}
	if b.CodeSignificant && b.Code != a.Code {
		return false
	}
	return true
}

func (drm *DiffResultsManager) AddGroup(baselineResults ...Result) error {
	baseline, err := NewBaselineResult(baselineResults...)
	if err != nil {
		return err
	}

	drm.baselines[baseline.ResultGroup] = baseline
	return nil
}

func (drm *DiffResultsManager) Run(rChan <-chan *Result) {
	go func() {
		defer close(drm.done)
		for result := range rChan {
			if baseline, ok := drm.baselines[result.ResultGroup]; !ok {
				// No baseline!
				drm.Append(result)
			} else if !baseline.Matches(result) {
				drm.Append(result)
			}
		}
		// Write out results
		drm.WriteResults()
	}()
}

func (drm *DiffResultsManager) Wait() {
	<-drm.done
}

func (drm *DiffResultsManager) Append(result *Result) {
	if _, ok := drm.keep[result.ResultGroup]; !ok {
		drm.keep[result.ResultGroup] = make([]*Result, 0)
	}
	drm.keep[result.ResultGroup] = append(drm.keep[result.ResultGroup], result)
}

func (drm *DiffResultsManager) WriteResults() error {
	fp := bufio.NewWriter(drm.fp)
	defer func() {
		fp.Flush()
		drm.fp.Close()
	}()
	for groupName, group := range drm.keep {
		if _, err := fmt.Fprintf(fp, "Group: %s\n", groupName); err != nil {
			return err
		}
		for _, result := range group {
			if _, err := fmt.Fprintf(fp, "\t%s\t%s\t%d\n", result.URL.String(), result.Host, result.Code); err != nil {
				return err
			}
		}
		fmt.Fprintf(fp, "\n")
	}
	return nil
}
