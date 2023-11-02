package types

// Person represents a person
type Person struct {
	ID          string                 `json:"id,omitempty"`
	ResourceID  string                 `json:"_rid,omitempty"`
	Timestamp   int                    `json:"_ts,omitempty"`
	Self        string                 `json:"_self,omitempty"`
	ETag        string                 `json:"_etag,omitempty"`
	Attachments string                 `json:"_attachments,omitempty"`
	LSN         int                    `json:"_lsn,omitempty"`
	Metadata    map[string]interface{} `json:"_metadata,omitempty"`

	Surname    string `json:"surname,omitempty"`
	UpdateTime int    `json:"updateTime,omitempty"`
}

// People represents people
type People struct {
	Count      int       `json:"_count,omitempty"`
	ResourceID string    `json:"_rid,omitempty"`
	People     []*Person `json:"Documents,omitempty"`
}
