package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ariym/mf-flow-infer-audio-events/engine"
	"strconv"
	"strings"
)

func executeTargets(ctx context.Context, req engine.Request, spec engine.Spec, preflight bool) (map[string]any, error) {
	if len(req.Targets) == 0 {
		return nil, fmt.Errorf("protocol v2 requires targets")
	}
	requests := make([]engine.Request, len(req.Targets))
	for i, target := range req.Targets {
		path := ""
		for _, b := range req.Bindings {
			if b.TargetIndex == i && b.Name == "audio" && b.Kind == "audio" {
				path = b.Path
			}
		}
		if path == "" {
			return nil, fmt.Errorf("target %d has no bound audio", i)
		}
		input := map[string]json.RawMessage{}
		for k, v := range req.Input {
			input[k] = v
		}
		delete(input, "audio_stream_index")
		if target.WholeFile {
			input["media"], _ = json.Marshal([]string{path})
		} else {
			start, err := timecodeSeconds(target.Start)
			if err != nil {
				return nil, err
			}
			end, err := timecodeSeconds(target.End)
			if err != nil || end <= start {
				return nil, fmt.Errorf("invalid target range")
			}
			input["media"], _ = json.Marshal([]any{map[string]any{"path": path, "stream_index": 0, "start": start, "end": end}})
		}
		child := req
		child.ProtocolVersion = 1
		child.JobID = fmt.Sprintf("%s-t%d", req.JobID, i)
		child.Input = input
		requests[i] = child
		if target.MediaPath == "" {
			return nil, fmt.Errorf("audio-event detection requires original media paths")
		}
		if _, err := engine.Execute(ctx, child, spec, true); err != nil {
			return nil, err
		}
	}
	if preflight {
		return map[string]any{"ready": true, "targets": len(requests)}, nil
	}
	tags := []any{}
	results := []any{}
	for i, child := range requests {
		result, err := engine.Execute(ctx, child, spec, false)
		if err != nil {
			return nil, err
		}
		for _, item := range result["tags"].([]any) {
			tag := item.(map[string]any)
			tag["target_index"] = i
			tag["media_path"] = req.Targets[i].MediaPath
			if value, ok := tag["value"].(map[string]any); ok {
				if media, ok := value["media"].(map[string]any); ok {
					for _, binding := range req.Bindings {
						if binding.TargetIndex == i && binding.Name == "audio" {
							media["source"] = map[string]any{"path": req.Targets[i].MediaPath, "sha256": binding.SourceSHA256, "stream_index": binding.StreamIndex}
						}
					}
				}
			}
			tags = append(tags, tag)
		}
		results = append(results, result)
	}
	return map[string]any{"tags": tags, "targets": results}, nil
}

func timecodeSeconds(value string) (float64, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid timecode")
	}
	h, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}
	m, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, err
	}
	s, err := strconv.ParseFloat(parts[2], 64)
	return h*3600 + m*60 + s, err
}
