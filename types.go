package main

import (
	"github.com/pkg/errors"
	"time"
)

type Trace struct {
	Data []Data `json:"data"`
}

type Data struct {
	TraceID   string              `json:"traceID"`
	Spans     []Span              `json:"spans"`
	Processes map[string]*Process `json:"processes"`
}

type Process struct {
	ID          string
	ServiceName string `json:"serviceName"`
	Tags        []Tag  `json:"tags"`
}
type Span struct {
	TraceID       string      `json:"traceID"`
	SpanID        string      `json:"spanID"`
	OperationName string      `json:"operationName"`
	Duration      int         `json:"duration"`
	ProcessID     string      `json:"processID"`
	References    []Reference `json:"references"`
	StartTime     int64       `json:"startTime"`
	Tags          []Tag       `json:"tags"`
}

type Tag struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type Reference struct {
	RefType string `json:"refType"`
	TraceID string `json:"traceID"`
	SpanID  string `json:"spanID"`
}

type TreeSpan struct {
	TraceID       string
	SpanID        string
	OperationName string
	StartTime     time.Time
	Duration      int
	Timebox       int
	MaxChildEnd   time.Time
	Process       *Process
	Children      []*TreeSpan
	Tags          []Tag
}

func (s *TreeSpan) HasTag(key string, name string) bool {
	for _, t := range s.Tags {
		if t.Key == key && t.Value == name {
			return true
		}
	}
	return false
}

type TreeTrace struct {
	Spans     map[string]*TreeSpan
	RootSpans *TreeSpan
}

func (s *TreeSpan) Recalculate() (time.Time, time.Time) {
	start := s.StartTime
	end := s.StartTime.Add(time.Duration(s.Duration) * time.Microsecond)

	for _, c := range s.Children {
		from, to := c.Recalculate()
		//if c.HasTag("status", "canceled") {
		//	continue
		//}
		if c.Process == nil || s.Process == nil || c.Process.ServiceName == s.Process.ServiceName {
			if from.Before(start) {
				start = from
			}
			if to.After(end) {
				end = to
			}
		}
	}
	s.Timebox = int(end.Sub(start).Microseconds())
	return start, end
}

func (s *TreeSpan) Walk(f func(parent *TreeSpan, span *TreeSpan) (bool, error)) error {
	for _, c := range s.Children {
		cont, err := f(s, c)
		if err != nil {
			return errors.WithStack(err)
		}
		if cont {
			err = c.Walk(f)
		}
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

type LoadedTrace struct {
	TraceID   string
	Tree      *TreeTrace
	Orphans   *TreeSpan
	AllTraces []*TreeSpan
}
