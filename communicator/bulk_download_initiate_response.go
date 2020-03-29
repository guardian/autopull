package communicator

type LightboxEntry struct {
	Id             string `json:"id"`
	Description    string `json:"description"`
	UserEmail      string `json:"userEmail"`
	AddedAtString  string `json:"addedAt"`
	ErrorCount     int    `json:"errorCount"`
	AvailCount     int    `json:"availCount"`
	RestoringCount int    `json:"restoringCount"`
}

type ArchiveEntryDownloadSynopsis struct {
	EntryId  string `json:"entryId"`
	Path     string `json:"path"`
	FileSize int64  `json:"fileSize"`
}

type BulkDownloadInitiateResponse struct {
	Status         string                         `json:"status"`
	Metadata       LightboxEntry                  `json:"metadata"`
	RetrievalToken string                         `json:"retrievalToken"`
	Entries        []ArchiveEntryDownloadSynopsis `json:"entries"`
}

/**
returns the total count and total size of all the file entries
*/
func (r *BulkDownloadInitiateResponse) TotalUpEntries() (int64, int64) {
	var count int64 = 0
	var totalBytes int64 = 0

	for _, ent := range r.Entries {
		count += 1
		totalBytes += ent.FileSize
	}

	return count, totalBytes
}
