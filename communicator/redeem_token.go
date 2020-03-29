package communicator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

/**
redeems the short-lived token and returns a pointer to the decoded response, or returns an error
 */
func (comm *Communicator) RedeemToken(token string, attempt int) (*BulkDownloadInitiateResponse, error) {
	client := http.Client{}
	url := fmt.Sprintf("%s/api/bulk/%s", comm.VaultDoorUri.String(), token)
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
		return &info, nil
	case 502:
		fallthrough
	case 503:
		fallthrough
	case 504:
		if attempt>10 {
			log.Printf("ERROR communicator.RedeemToken Server is not available after %d attempts, giving up.", attempt)
			return nil, errors.New("server was not available")
		}
		log.Printf("ERROR communicator.RedeemToken Server is not available on attempt %d. Retrying after a delay...", attempt)
		time.Sleep(5*time.Second)
		return comm.RedeemToken(token, attempt)
	default:
		log.Printf("ERROR communicator.RedeemToken Server returned %d: %s", resp.StatusCode, string(bodyContent))
		return nil, errors.New("invalid server response")
	}
}
