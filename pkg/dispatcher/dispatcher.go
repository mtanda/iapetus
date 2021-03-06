package dispatcher

import (
	"github.com/kobtea/iapetus/pkg/model"
	"github.com/kobtea/iapetus/pkg/util"
	"net/http"
	"time"
)

type Input struct {
	Query    string
	Matchers []string
	time     time.Time
	start    time.Time
	end      time.Time
}

func NewInput(r *http.Request) (Input, error) {
	var in Input
	r.ParseForm()
	if v, ok := r.Form["query"]; ok {
		in.Query = v[0]
	}
	if v, ok := r.Form["match[]"]; ok {
		in.Matchers = v
	}
	if v, ok := r.Form["time"]; ok {
		t, err := util.ParseTime(v[0])
		if err != nil {
			return Input{}, err
		}
		in.time = t
	}
	if v, ok := r.Form["start"]; ok {
		t, err := util.ParseTime(v[0])
		if err != nil {
			return Input{}, err
		}
		in.start = t
	}
	if v, ok := r.Form["end"]; ok {
		t, err := util.ParseTime(v[0])
		if err != nil {
			return Input{}, err
		}
		in.end = t
	}
	return in, nil
}

func NewDispatcher(cluster model.Cluster) *Dispatcher {
	return &Dispatcher{
		Cluster: cluster,
	}
}

type Dispatcher struct {
	Cluster model.Cluster
}

func (d Dispatcher) resolveNode(name string) *model.Node {
	for _, n := range d.Cluster.Nodes {
		if n.Name == name {
			return &n
		}
	}
	return nil
}

func (d Dispatcher) FindNode(in Input) *model.Node {
	for _, rule := range d.Cluster.Rules {
		if !rule.Range.IsZero() {
			if !in.start.IsZero() || !in.end.IsZero() {
				if rule.Range.Satisfy(in.start, in.end) {
					return d.resolveNode(rule.Target)
				}
			}
		}
		if !rule.Time.IsZero() {
			if !in.time.IsZero() {
				if rule.Time.Satisfy(in.time) {
					return d.resolveNode(rule.Target)
				}
			}
		}
		if !rule.Start.IsZero() {
			if !in.start.IsZero() {
				if rule.Start.Satisfy(in.start) {
					return d.resolveNode(rule.Target)
				}
			}
		}
		if !rule.End.IsZero() {
			if !in.end.IsZero() {
				if rule.End.Satisfy(in.end) {
					return d.resolveNode(rule.Target)
				}
			}
		}
	}
	return d.defaultNode()
}

func (d Dispatcher) defaultNode() *model.Node {
	for _, r := range d.Cluster.Rules {
		if r.Default {
			return d.resolveNode(r.Target)
		}
	}
	if len(d.Cluster.Nodes) > 0 {
		return &d.Cluster.Nodes[0]
	}
	return nil
}
