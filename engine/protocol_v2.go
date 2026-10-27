package engine

// Target freezes a submitted range independently of subsequent segment edits.
type Target struct {
	WholeFile   bool   `json:"whole_file,omitempty"`
	MediaPath   string `json:"media_path"`
	MediaFileID int64  `json:"media_file_id"`
	SegmentID   int64  `json:"segment_id,omitempty"`
	Start       string `json:"start"`
	End         string `json:"end"`
	StreamIndex *int   `json:"stream_index,omitempty"`
}
type Binding struct {
	TargetIndex  int    `json:"target_index"`
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	Key          string `json:"key"`
	Version      int    `json:"version"`
	TagID        int64  `json:"tag_id,omitempty"`
	SHA256       string `json:"sha256"`
	Path         string `json:"path,omitempty"`
	StreamIndex  int    `json:"stream_index,omitempty"`
	SourceSHA256 string `json:"source_sha256,omitempty"`
}
