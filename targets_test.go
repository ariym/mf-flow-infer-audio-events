package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ariym/mf-flow-infer-audio-events/engine"
)

func TestAudioEventTargetsPreserveRangeSourceAndEmptyResults(t *testing.T) {
	for _, empty := range []bool{false, true} {
		t.Run(map[bool]string{false: "events", true: "empty"}[empty], func(t *testing.T) {
			root := t.TempDir()
			db := filepath.Join(root, "mf.db")
			audio := filepath.Join(root, "audio.wav")
			output := filepath.Join(root, "results")
			program := filepath.Join(root, "program")
			for _, p := range []string{db, audio} {
				if err := os.WriteFile(p, []byte("fixture"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Mkdir(output, 0700); err != nil {
				t.Fatal(err)
			}
			code := `#!/usr/bin/env python3
import json,sys
if sys.argv[1]=='describe':print(json.dumps({'program_name':'demo','version':2}));sys.exit(0)
r=json.load(sys.stdin)
if sys.argv[1]=='preflight':print('{}');sys.exit(0)
m=r['media'][0]
assert isinstance(m,dict) and m['start']==2 and m['end']==3
rows=[] if r.get('options',{}).get('empty') else [{'start':2,'end':2.5,'duration':0.5,'sampling_rate':16000,'sample_width':2,'channels':1}]
value={'operation':'split','events':rows,'media':{'path':m['path']},'options':r.get('options',{}),'raw':{'regions':[]}}
print(json.dumps({'provenance':{},'result':{'items':[{'media_path':m['path'],'audio_events':value}]}}))
`
			if err := os.WriteFile(program, []byte(code), 0755); err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(output)
			options, _ := json.Marshal(map[string]bool{"empty": empty})
			req := engine.Request{ProtocolVersion: 2, JobID: "events-v2", DBPath: db, Input: map[string]json.RawMessage{"output_dir": raw, "options": options}, Programs: map[string]engine.Dependency{"demo": {Version: 2, Path: program}}, Targets: []engine.Target{{MediaFileID: 1, MediaPath: "/original/video.mp4", Start: "00:00:02.000000", End: "00:00:03.000000"}}, Bindings: []engine.Binding{{TargetIndex: 0, Name: "audio", Kind: "audio", Path: audio, SourceSHA256: "source-digest", StreamIndex: 4}}}
			spec := engine.Spec{Key: "test", Steps: []engine.Step{{Program: "demo", Operation: "split"}}, Programs: map[string]int{"demo": 2}, Tags: map[string]int{"audio_events": 1}}
			result, err := executeTargets(context.Background(), req, spec, false)
			if err != nil {
				t.Fatal(err)
			}
			tags := result["tags"].([]any)
			if len(tags) != 1 {
				t.Fatalf("expected event value, even with no detections: %+v", result)
			}
			tag := tags[0].(map[string]any)
			if tag["target_index"] != 0 || tag["media_path"] != "/original/video.mp4" {
				t.Fatal("lost target", tag)
			}
			value := tag["value"].(map[string]any)
			events := value["events"].([]any)
			if empty && len(events) != 0 || !empty && len(events) != 1 {
				t.Fatal("incorrect events", events)
			}
			media := value["media"].(map[string]any)
			source := media["source"].(map[string]any)
			if source["sha256"] != "source-digest" || source["stream_index"] != 4 {
				t.Fatal("lost source", source)
			}
		})
	}
}
