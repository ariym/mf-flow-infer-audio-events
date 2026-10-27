# mf-flow-infer-audio-events

Detect audio regions using auditok energy detection or its optional WebRTC validator. Detection parameters, channel selection, pause controls, plot output and clip export are exposed through options. Regions include timing, sample metadata, source provenance and effective settings. No semantic labels or confidence values are invented.

Version 2 is a macOS ARM64 flow using the mf protocol-v2 target/binding envelope (legacy protocol-v1 calls remain supported). Dependencies: auditok@2. No mf-hub server or client database implementation is linked into this executable. The supplied client database path is checked for existence but never opened or written.

## Run

Install and explicitly prepare the dependency programs. Edit `example-request.json` with existing absolute media, program, database and output-directory paths.

```sh
bin/infer_audio_events describe
bin/infer_audio_events preflight < example-request.json
bin/infer_audio_events < example-request.json
```

`options` configures the final program operation. `program_options.<name>` configures a particular dependency; `describe` on that program exposes its complete option schema. Multiple audio streams require media objects with explicit `stream_index`; optional start/end times are in seconds. Media is local to the executing machine. Every selected stream is fully decoded during preflight before any inference starts. Missing files, corrupt media, missing models, unsupported options or incompatible profiles abort the entire job.

Results are published atomically at `<input.output_dir>/<job_id>/result.json` with versioned tags, complete stage responses, source/model provenance, and hashed artifacts. Intermediate writes stay in a staging directory. SIGINT/SIGTERM terminates subprocess groups and removes active staging output. Stable job IDs and content/configuration fingerprints make identical retries return the saved result; conflicting retries fail. Artifact checks detect corrupted prior results. Uncatchable kills can leave unpublished `.mf-*` staging directories, which may be cleaned when the matching advisory lock is no longer held.

## Build

```sh
go test ./...
go vet ./...
python3 scripts/release.py --repo ariym/mf-flow-infer-audio-events --version 2 --platform darwin/arm64
```

The source manifest is `flow.json`; generated `dist/manifest.json` has real checksums and release URLs. Only validated darwin/arm64 artifacts are published. Source is based on mf-hub's canonical flow template and its reusable `engine/` package. No installation adapter or client tag storage is implemented here.

## Output requests

Version 2 declares `audio_events@1` in `execution.produces` and requires resolved audio. Once its archive and immutable hub manifest are published, `mf request audio_events video.mp4` can discover it and queue missing audio extraction automatically. V1 only pins the schema and cannot be discovered as a producer. Auditok detects energy-based sound regions; it does not label yawns or other semantic event categories.
