package communicator

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/guardian/autopull/config"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

/**
consumes an NDJSON stream of ArchiveEntryDownloadSynopsis and yields them to the returned channels
*/
func asyncStreamingRetrieveContent(resp io.Reader) (chan *ArchiveEntryDownloadSynopsis, chan error) {
	outputCh := make(chan *ArchiveEntryDownloadSynopsis, 100)
	errCh := make(chan error, 100)

	go func() {
		scanner := bufio.NewScanner(resp)
		for scanner.Scan() {
			rawContent := scanner.Bytes()
			if len(rawContent) > 0 {
				var entry ArchiveEntryDownloadSynopsis
				unmarshalErr := json.Unmarshal(rawContent, &entry)
				if unmarshalErr != nil {
					errCh <- unmarshalErr
				} else {
					log.Printf("DEBUG asyncStreamingRetrieveContent got %v", entry)
					outputCh <- &entry
				}
			} else {
				log.Printf("INFO asyncStreamingRetrieveContent got zero-length record")
			}
		}

		if err := scanner.Err(); err != nil {
			errCh <- err
		}
		outputCh <- nil
		return
	}()

	return outputCh, errCh
}

func consumeDownloadStream(resp io.Reader) (*[]ArchiveEntryDownloadSynopsis, error) {
	output := make([]ArchiveEntryDownloadSynopsis, 0)

	contentCh, errCh := asyncStreamingRetrieveContent(resp)
	var lastError error

	for {
		select {
		case rec := <-contentCh:
			if rec == nil {
				log.Print("INFO consumeDownloadStream reached end of stream")
				if lastError != nil {
					return nil, lastError
				} else {
					return &output, nil
				}
			}
			output = append(output, *rec)
		case err := <-errCh:
			log.Print("WARNING consumeDownloadStream got an error: ", err)
			lastError = err
		}
	}
}

/**
the V2 api separates the token get and retrieval stages.  The BulkDownloadInitiate response has a nil entries list, which
we must populate with a subsequent call to summaryStream.
This function consumes the result of summaryStream and fills the 'entries' field for us.
*/
func (comm *Communicator) FetchDownloadSynopsisStreaming(partialResponse *BulkDownloadInitiateResponse, httpClient *http.Client, attempt int) (*BulkDownloadInitiateResponse, error) {
	log.Printf("DEBUG communicator.RedeemToken no download synopsis data, retrieving from stream...")

	url := fmt.Sprintf("%s/api/bulkv2/%s/summarystream", comm.ArchiveHunterUri.String(), partialResponse.RetrievalToken)
	resp, err := httpClient.Get(url)
	if err != nil {
		log.Printf("ERROR communicator.FetchDownloadSynopsisStreaming could not make connection to server: %s", err)
		return nil, err
	}
	defer resp.Body.Close()

	entriesPtr, retrieveErr := consumeDownloadStream(resp.Body)
	if retrieveErr != nil {
		return nil, retrieveErr
	} else {
		copiedResponse := *partialResponse
		copiedResponse.Entries = *entriesPtr
		log.Printf("DEBUG FetchDownloadSynopsisStreaming got final result %v", copiedResponse)
		return &copiedResponse, nil
	}
}

/**
redeems the short-lived token and returns a pointer to the decoded response, or returns an error
*/
func (comm *Communicator) RedeemToken(token config.DownloadTokenUri, attempt int) (*BulkDownloadInitiateResponse, error) {
	client := http.Client{}
	var url string
	if token.ValidateVaultDoor() {
		url = fmt.Sprintf("%s/api/bulk/%s", comm.VaultDoorUri.String(), token.Token)
	} else {
		url = fmt.Sprintf("%s/api/bulkv2/%s", comm.ArchiveHunterUri.String(), token.Token)
	}

	resp, err := client.Get(url)

	if err != nil {
		log.Printf("ERROR communicator.RedeemToken could not make connection to server: %s", err)
		return nil, err
	}

	bodyContent, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("ERROR communicator.RedeemToken could not read server response: %s", readErr)
		return nil, err
	}

	resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var info BulkDownloadInitiateResponse
		unmarshalErr := json.Unmarshal(bodyContent, &info)
		if unmarshalErr != nil {
			log.Printf("ERROR communicator.RedeemToken could not understand server response: %s", unmarshalErr)
			return nil, unmarshalErr
		}
		if info.Entries == nil {
			return comm.FetchDownloadSynopsisStreaming(&info, &client, 0)
		} else {
			return &info, nil
		}
	case 502:
		fallthrough
	case 503:
		fallthrough
	case 504:
		if attempt > 10 {
			log.Printf("ERROR communicator.RedeemToken Server is not available after %d attempts, giving up.", attempt)
			return nil, errors.New("server was not available")
		}
		log.Printf("ERROR communicator.RedeemToken Server is not available on attempt %d. Retrying after a delay...", attempt)
		time.Sleep(5 * time.Second)
		return comm.RedeemToken(token, attempt)
	default:
		log.Printf("ERROR communicator.RedeemToken Server returned %d: %s", resp.StatusCode, string(bodyContent))
		return nil, errors.New("invalid server response")
	}
}
