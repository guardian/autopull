package communicator

type LightboxEntry struct {
	Id string `json:"id"`
	Description string `json:"description"`
	UserEmail string `json:"userEmail"`
	AddedAtString string `json:"addedAt"`
	ErrorCount int `json:"errorCount"`
	AvailCount int `json:"availCount"`
	RestoringCount int `json:"restoringCount"`
}

type ArchiveEntryDownloadSynopsis struct {
	EntryId string `json:"entryId"`
	Path string `json:"path"`
	FileSize int64 `json:"fileSize"`
}

type BulkDownloadInitiateResponse struct {
	Status string `json:"status"`
	Metadata LightboxEntry `json:"metadata"`
	RetrievalToken string `json:"retrievalToken"`
	Entries []ArchiveEntryDownloadSynopsis `json:"entries"`
}
