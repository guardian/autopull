package communicator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type DownloadManagerItemResponse struct {
	Status        string  `json:"status"`
	RestoreStatus string  `json:"restoreStatus"`
	DownloadLink  url.URL `json:"downloadLink"`
}

func ParseDownloadManagerItemResponse(from []byte) (*DownloadManagerItemResponse, error) {
	var contentMap map[string]interface{}
	unmarshalErr := json.Unmarshal(from, &contentMap)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	status, statusIsOk := contentMap["status"].(string)
	if !statusIsOk {
		return nil, errors.New("status field was not correctly formatted")
	}
	restoreStatus, restoreStatusIsOk := contentMap["restoreStatus"].(string)
	if !restoreStatusIsOk {
		return nil, errors.New("restoreStatus field was not correctly formatted")
	}
	downloadLinkStr, dlLinkIsOk := contentMap["downloadLink"].(string)
	if !dlLinkIsOk {
		return nil, errors.New("downloadLink field was not correctly formatted")
	}
	downloadLinkPtr, urlParseErr := url.Parse(downloadLinkStr)
	if urlParseErr != nil {
		return nil, urlParseErr
	}

	return &DownloadManagerItemResponse{
		Status:        status,
		RestoreStatus: restoreStatus,
		DownloadLink:  *downloadLinkPtr,
	}, nil
}

/**
gets the download link for the given item or an error
*/
func (comm *Communicator) GetItemLink(longLivedToken string, fileId string, attempt int) (*DownloadManagerItemResponse, error) {
	serverBase := comm.GetActiveUrl()
	url := fmt.Sprintf("%s/api/bulk/%s/get/%s", serverBase.String(), longLivedToken, fileId)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("ERROR communicator.GetItemLink could not establish connection: %s", err)
		return nil, err
	}

	bodyContent, readErr := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		log.Printf("ERROR communicator.GetItemLink could not read server response: %s", err)
		return nil, err
	}

	switch resp.StatusCode {
	case 200:
		rtn, parseErr := ParseDownloadManagerItemResponse(bodyContent)
		if parseErr != nil {
			log.Printf("ERROR communicator.GetItemLink offending content was %s", string(bodyContent))
			log.Printf("ERROR communicator.GetItemLink could not understand server response: %s", parseErr)
			return nil, parseErr
		}
		return rtn, nil
	case 502:
		fallthrough
	case 503:
		fallthrough
	case 504:
		if attempt > 10 {
			log.Printf("ERROR communicator.GetItemLink could not contact server after %d attempts, giving up", attempt)
			return nil, errors.New("server not responding")
		}
		log.Printf("ERROR communcator.GetItemLink could not contact server on attemt %d. Retrying after a delay...", attempt)
		time.Sleep(5 * time.Second)
		return comm.GetItemLink(longLivedToken, fileId, attempt+1)
	default:
		log.Printf("ERROR communicator.GetItemLink server returned an error %d: %s", resp.StatusCode, string(bodyContent))
		return nil, errors.New("server returned an error")
	}
}
