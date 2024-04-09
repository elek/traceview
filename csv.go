package main

import (
	"encoding/csv"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CSV struct {
	FilterFile string `arg:""`
	Path       string `arg:""`
}

func MatchName(name string) func(span *TreeSpan) bool {
	return func(span *TreeSpan) bool {
		return span.OperationName == name
	}
}
func (a CSV) Run() error {
	var filters []FilterAndName
	lines, err := os.ReadFile(a.FilterFile)
	if err != nil {
		return errors.WithStack(err)
	}
	csvHeaders := []string{"trace_id"}
	for _, line := range strings.Split(string(lines), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		name := line
		parts := strings.SplitN(line, " ", 2)
		if len(parts) > 1 {
			name = parts[0]
			line = parts[1]
		}

		csvHeaders = append(csvHeaders, name)

		selector := ""
		parts = strings.SplitN(line, "#", 2)
		if len(parts) > 1 {
			selector = "#" + parts[1]
		}
		parts = strings.SplitN(parts[0], "@", 2)
		if len(parts) > 1 {
			selector = "@" + parts[1]
		}
		target := parts[0]

		filters = append(filters, FilterAndName{
			Name:     name,
			Filter:   MatchName(target),
			Selector: selector,
		})

	}

	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	err = writer.Write(csvHeaders)
	if err != nil {
		return errors.WithStack(err)
	}

	stat, err := os.Stat(a.Path)
	if err != nil {
		return errors.WithStack(err)
	}
	if stat.IsDir() {
		entries, err := os.ReadDir(a.Path)
		if err != nil {
			return errors.WithStack(err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), "json") {
				err := a.processFile(filepath.Join(a.Path, e.Name()), filters, writer)
				if err != nil {
					return err
				}
			}
		}
	} else {
		err := a.processFile(a.Path, filters, writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a CSV) processFile(path string, filters []FilterAndName, writer *csv.Writer) error {

	trace, err := load(path)
	if err != nil {
		return errors.WithStack(err)
	}
	res := map[string]string{}
	for _, f := range filters {
		for _, span := range trace.AllTraces {
			if f.Filter(span) {
				if strings.HasPrefix(f.Selector, "#") {
					for _, t := range span.Process.Tags {
						if t.Key == f.Selector[1:] {
							res[f.Name] = fmt.Sprintf("%s", t.Value)
						}
					}
				} else {
					switch f.Selector {
					case "@time":
						res[f.Name] = fmt.Sprintf("%s", span.StartTime.Format(time.RFC3339))
					default:
						res[f.Name] = fmt.Sprintf("%d", span.Duration)
					}
				}
				break
			}
		}
	}

	csvLine := []string{trace.TraceID}
	for _, f := range filters {
		val, ok := res[f.Name]
		if !ok {
			csvLine = append(csvLine, "")
		} else {
			csvLine = append(csvLine, val)
		}
	}
	err = writer.Write(csvLine)
	if err != nil {
		return errors.WithStack(err)
	}
	writer.Flush()
	return nil
}

type FilterAndName struct {
	Name     string
	Filter   func(span *TreeSpan) bool
	Selector string
}
