package main

import (
	"encoding/json"
	"github.com/zeebo/errs/v2"
	"os"
	"time"
)

func load(name string) (*LoadedTrace, error) {
	l := LoadedTrace{
		AllTraces: make([]*TreeSpan, 0),
	}
	raw, err := os.ReadFile(name)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	t := Trace{}
	err = json.Unmarshal(raw, &t)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	for n, _ := range t.Data[0].Processes {
		t.Data[0].Processes[n].ID = n
	}
	var spans = map[string]*TreeSpan{}
	leaf := map[string]bool{}
	for _, s := range t.Data[0].Spans {
		leaf[s.SpanID] = false
		process := t.Data[0].Processes[s.ProcessID]
		spans[s.SpanID] = &TreeSpan{
			TraceID:       s.TraceID,
			SpanID:        s.SpanID,
			OperationName: s.OperationName,
			Duration:      s.Duration,
			StartTime:     time.UnixMicro(s.StartTime),
			Process:       process,
			Children:      make([]*TreeSpan, 0),
			Tags:          s.Tags,
		}
		l.AllTraces = append(l.AllTraces, spans[s.SpanID])

		if l.TraceID == "" {
			l.TraceID = s.TraceID
		}

	}

	orphans := &TreeSpan{
		Children: make([]*TreeSpan, 0),
	}

	for _, s := range t.Data[0].Spans {
		for _, ref := range s.References {
			switch ref.RefType {
			case "CHILD_OF":
				leaf[s.SpanID] = true
				root, found := spans[ref.SpanID]
				if !found {
					orphans.Children = append(orphans.Children, spans[s.SpanID])
				} else {
					root.Children = append(root.Children, spans[s.SpanID])
				}
			default:
				panic("unsupported " + ref.RefType)
			}
		}
	}

	root := &TreeSpan{
		Children: make([]*TreeSpan, 0),
	}
	for id, f := range leaf {
		if !f {
			root.Children = append(root.Children, spans[id])
		}
	}

	root.Recalculate()
	root.Duration = root.Timebox
	l.Tree = &TreeTrace{
		Spans:     spans,
		RootSpans: root,
	}
	l.Orphans = orphans

	return &l, nil

}
