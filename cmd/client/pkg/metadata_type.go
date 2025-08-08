package pkg

// metaData is the data associated with each blob.
type metaData struct {
	UUID      string `json:"uuid"`
	Hash      string `json:"hash"`
	TimeStamp string `json:"timestamp"`
}
