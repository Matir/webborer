package filter

import (
	"github.com/Matir/webborer/task"
)

type DotProductExpander struct {
	Hostlist []string
}

func NewDotProductExpander(hostlist []string) *DotProductExpander {
	return &DotProductExpander{Hostlist: hostlist}
}

func (dp *DotProductExpander) Expand(inchan <-chan *task.Task) <-chan *task.Task {
	outChan := make(chan *task.Task, cap(inchan))
	go func() {
		defer close(outChan)
		for it := range inchan {
			for _, host := range dp.Hostlist {
				newIt := it.Copy()
				newIt.Host = host
				outChan <- newIt
			}
		}
	}()
	return outChan
}
