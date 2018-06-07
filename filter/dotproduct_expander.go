package filter

import (
	"github.com/Matir/webborer/task"
	"github.com/Matir/webborer/workqueue"
)

type DotProductExpander struct {
	Hostlist []string
	adder    workqueue.QueueAddCount
}

func NewDotProductExpander(hostlist []string) *DotProductExpander {
	return &DotProductExpander{Hostlist: hostlist}
}

func (dp *DotProductExpander) SetAddCount(adder workqueue.QueueAddCount) {
	dp.adder = adder
}

func (dp *DotProductExpander) Expand(inchan <-chan *task.Task) <-chan *task.Task {
	outChan := make(chan *task.Task, cap(inchan))
	go func() {
		defer close(outChan)
		for it := range inchan {
			outChan <- it
			for _, host := range dp.Hostlist {
				newIt := it.Copy()
				newIt.Host = host
				dp.adder(1)
				outChan <- newIt
			}
		}
	}()
	return outChan
}
