package dispatcher

import (
	"github.com/kobtea/iapetus/pkg/config"
	"github.com/kobtea/iapetus/pkg/util"
	"net/http"
	"time"
)

type Input struct {
	query string
	time  time.Time
	start time.Time
	end   time.Time
}

func NewInput(r *http.Request) (Input, error) {
	var in Input
	if v := r.FormValue("query"); v != "" {
		in.query = v
	}
	if v := r.FormValue("time"); v != "" {
		t, err := util.ParseTime(v)
		if err != nil {
			return Input{}, err
		}
		in.time = t
	}
	if v := r.FormValue("start"); v != "" {
		t, err := util.ParseTime(v)
		if err != nil {
			return Input{}, err
		}
		in.start = t
	}
	if v := r.FormValue("end"); v != "" {
		t, err := util.ParseTime(v)
		if err != nil {
			return Input{}, err
		}
		in.end = t
	}
	return in, nil
}

func NewDispatcher(cluster config.Cluster) *Dispatcher {
	return &Dispatcher{
		Cluster: cluster,
	}
}

type Dispatcher struct {
	Cluster config.Cluster
}

func (d Dispatcher) resolveNode(name string) *config.Node {
	for _, n := range d.Cluster.Nodes {
		if n.Name == name {
			return &n
		}
	}
	return nil
}

func (d Dispatcher) FindNode(in Input) *config.Node {
	for _, rule := range d.Cluster.Rules {
		if !rule.Range.IsZero() {
			if !in.start.IsZero() || !in.end.IsZero() {
				if rule.Range.Satisfy(in.start, in.end) {
					return d.resolveNode(rule.Target)
				}
			}
		}
	}
	return d.defaultNode()
}

func (d Dispatcher) defaultNode() *config.Node {
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
