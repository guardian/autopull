package downloadmanager

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/guardian/autopull/communicator"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type DownloadManager interface {
	Init() error
	Shutdown(wait bool)
	DownloadThread()
	PerformDownload(incomingEntry *communicator.ArchiveEntryDownloadSynopsis, linkInfo *communicator.DownloadManagerItemResponse) error
	Enqueue(incomingEntry communicator.ArchiveEntryDownloadSynopsis)
}

//NOTE: anything in here must be threadsafe, and is considered immutable for that reason
type DownloadManagerImpl struct {
	DownloadThreadCount int
	LongLivedToken      string
	Communicator        *communicator.Communicator
	incomingChannel     chan communicator.ArchiveEntryDownloadSynopsis
	errorChannel        chan error
	BasePath            string
	CanClobber          bool
	waitGroup           *sync.WaitGroup
}

func NewDownloadManager(comm *communicator.Communicator, longLivedToken string, threadCount int, bufferSize int, basePath string, canClobber bool) DownloadManager {
	var properBasePath string
	if strings.HasSuffix(basePath, "/") {
		r := regexp.MustCompile("/+$")
		properBasePath = r.ReplaceAllString(basePath, "")
	} else {
		properBasePath = basePath
	}
	return &DownloadManagerImpl{
		DownloadThreadCount: threadCount,
		LongLivedToken:      longLivedToken,
		Communicator:        comm,
		incomingChannel:     make(chan communicator.ArchiveEntryDownloadSynopsis, bufferSize),
		errorChannel:        make(chan error, bufferSize),
		BasePath:            properBasePath,
		CanClobber:          canClobber,
		waitGroup:           &sync.WaitGroup{},
	}
}

func (d *DownloadManagerImpl) Init() error {
	log.Printf("DEBUG DownloadManager.Init initialising %d download routines", d.DownloadThreadCount)
	for i := 0; i < d.DownloadThreadCount; i += 1 {
		go d.DownloadThread()
	}
	return nil
}

func (d *DownloadManagerImpl) Shutdown(wait bool) {
	for i := 0; i < d.DownloadThreadCount; i += 1 {
		d.incomingChannel <- communicator.ArchiveEntryDownloadSynopsis{}
	}
	if wait {
		d.waitGroup.Wait()
	}
}

func (d *DownloadManagerImpl) Enqueue(incomingEntry communicator.ArchiveEntryDownloadSynopsis) {
	d.incomingChannel <- incomingEntry
}

func (d *DownloadManagerImpl) DownloadThread() {
	log.Print("DEBUG DownloadManager.DownloadThread initialising")
	d.waitGroup.Add(1)
	for {
		select {
		case incomingEntry := <-d.incomingChannel:
			log.Printf("got %s", spew.Sdump(incomingEntry))
			if incomingEntry.EntryId == "" {
				log.Printf("INFO DownloadManager.DownloadThread terminating")
				d.waitGroup.Done()
				return
			}
			log.Printf("INFO DownloadManager.DownloadThread getting download link for %s", incomingEntry.EntryId)
			linkInfoPtr, linkInfoErr := d.Communicator.GetItemLink(d.LongLivedToken, incomingEntry.EntryId, 0)
			if linkInfoErr != nil {
				if linkInfoErr != nil {
					log.Printf("ERROR DownloadManager.DownloadThread could not get download link: %s", linkInfoErr)
					continue
				}
			}

			switch linkInfoPtr.RestoreStatus {
			case "RS_PENDING":
				fallthrough
			case "RS_UNDERWAY":
				fallthrough
			case "RS_ERROR":
				log.Printf("ERROR DownloadManager.DownloadThread %s is not available to download, restore status is %s", incomingEntry.Path, linkInfoPtr.RestoreStatus)
			case "RS_UNNEEDED":
				fallthrough
			case "RS_ALREADY":
				fallthrough
			case "RS_SUCCESS":
				log.Printf("INFO DownloadManager.DownloadThread %s is available to download", incomingEntry.Path)
				dlErr := d.PerformDownload(&incomingEntry, linkInfoPtr)
				if dlErr != nil {
					log.Printf("ERROR DownloadManager.DownloadThread could not download content for %s: %s", incomingEntry.Path, dlErr)
				}
			}
		}
	}
}

func verifyFile(pathTarget string, canClobber bool) error {
	_, statErr := os.Stat(pathTarget)
	if statErr == nil {
		log.Printf("WARN DownloadManager.PerformDownload a file already exists at %s", pathTarget)
		if !canClobber {
			log.Printf("WARN DownloadManager.PerformDownload not overwriting an existing file. If you want to overwrite, specify this in the config")
			return errors.New("file already exists")
		}
	} else {
		if !os.IsNotExist(statErr) {
			log.Printf("ERROR DownloadManager.PerformDownload could not check for existence of file: %s", statErr)
			return statErr
		}
	}
	return nil
}

func prepareDirectories(pathTarget string) error {
	pathParts := strings.Split(pathTarget, string(os.PathSeparator))
	if len(pathParts) == 0 {
		log.Printf("WARN file without any subdirectory: %s", pathTarget)
		return nil
	} else {
		dirname := strings.Join(pathParts[0:len(pathParts)-1], string(os.PathSeparator))
		log.Printf("DEBUG creating directory for %s", dirname)
		err := os.MkdirAll(dirname, 0755)
		if err != nil {
			if !os.IsExist(err) {
				log.Printf("ERROR DownloadManager.prepareDirectories could not create folder for download: %s", err)
				return err
			}
		}
		return nil
	}
}

func doDownload(pathTarget string, downloadUrl string, expectedSize int64) (bool, error) {
	file, openErr := os.OpenFile(pathTarget, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if openErr != nil {
		log.Printf("ERROR DownloadManager.PerformDownload could not open target file %s: %s", pathTarget, openErr)
		return false, openErr
	}
	defer file.Close()

	dlResponse, dlErr := http.Get(downloadUrl)
	if dlErr != nil {
		log.Printf("ERROR DownloadManager.PerformDownload could not initiate download: %s", dlErr)
		return true, dlErr
	}
	defer dlResponse.Body.Close()

	switch dlResponse.StatusCode {
	case 200:
		log.Printf("INFO DownloadManager.PerformDownload downloading %s to %s", downloadUrl, pathTarget)
		bytesCopied, copyErr := io.Copy(file, dlResponse.Body)
		if copyErr != nil {
			log.Printf("ERROR DownloadManager.PerformDownload download of %s failed: %s", pathTarget, copyErr)
			os.Remove(pathTarget)
			return false, copyErr
		}
		if bytesCopied < expectedSize {
			log.Printf("WARN DownloadManager.PerformDownload %s potential short download, expected %d got %d", pathTarget, expectedSize, bytesCopied)
		} else if bytesCopied > expectedSize {
			log.Printf("WARN DownloadManager.PerformDownload %s downloaded more bytes than expected??? Strange. Expected %d got %d", pathTarget, expectedSize, bytesCopied)
		}
		return false, nil
	case 404:
		errorContent, _ := ioutil.ReadAll(dlResponse.Body)
		log.Printf("ERROR DownloadManager.PerformDownload %s was not found. Server said %s", downloadUrl, string(errorContent))
		return false, errors.New("download not found")
	case 403:
		log.Printf("ERROR DownloadManager.PerformDownload server responded permission denied, maybe token expired? Try re-starting the download from your browser")
		return false, errors.New("server permission denied")
	case 502:
		fallthrough
	case 503:
		fallthrough
	case 504:
		return true, errors.New("server was not available")
	default:
		errorContent, _ := ioutil.ReadAll(dlResponse.Body)
		log.Printf("ERROR DownloadManager.PerformDownload server responded %d: %s", dlResponse.StatusCode, string(errorContent))
		return false, errors.New("server error")
	}
}

func makeAbsoluteUrl(baseUri url.URL, tailUri url.URL) (url.URL, error) {
	//tailUri, parseErr := url.Parse(tailUriString)
	//if parseErr!=nil {
	//	log.Printf("ERROR DownloadManager.makeAbsoluteUrl invalid tail uri %s", tailUriString)
	//	return url.URL{}, parseErr
	//}

	rtn := url.URL{
		Scheme:     baseUri.Scheme,
		Opaque:     baseUri.Opaque,
		Host:       baseUri.Host,
		Path:       tailUri.Path,
		RawPath:    tailUri.RawPath,
		ForceQuery: false,
		RawQuery:   tailUri.RawQuery,
		Fragment:   tailUri.Fragment,
	}
	return rtn, nil
}

func (d *DownloadManagerImpl) PerformDownload(incomingEntry *communicator.ArchiveEntryDownloadSynopsis, linkInfo *communicator.DownloadManagerItemResponse) error {
	pathParts := []string{
		d.BasePath,
		incomingEntry.Path,
	}

	removeDupSlash := regexp.MustCompile("/{2+}")
	pathTarget := removeDupSlash.ReplaceAllString(strings.Join(pathParts, string(os.PathSeparator)), "/")

	log.Printf("DEBUG DownloadManager.PerformDownload pathTarget is %s", pathTarget)

	downloadUri, _ := makeAbsoluteUrl(d.Communicator.VaultDoorUri, linkInfo.DownloadLink)
	//verify if a file already exists
	verifyErr := verifyFile(pathTarget, d.CanClobber)
	if verifyErr != nil {
		return verifyErr
	}

	//create directories if necessary
	dirErr := prepareDirectories(pathTarget)
	if dirErr != nil {
		return dirErr
	}

	//perform download, retrying on recoverable errors
	attempts := 0
	for {
		shouldRetry, dlErr := doDownload(pathTarget, downloadUri.String(), incomingEntry.FileSize)
		if dlErr == nil {
			log.Printf("INFO DownloadManager.PerformDownload completed download of %s", pathTarget)
			break
		} else {
			if shouldRetry {
				attempts += 1
				if attempts >= 10 {
					log.Printf("ERROR DownloadManager.PerformDownload giving up after %d attempts", attempts)
					return errors.New(fmt.Sprintf("gave up after %d attempts", attempts))
				}
				log.Printf("WARN DownloadManager.PerformDownload %s, retrying after a delay...", dlErr)
				time.Sleep(5 * time.Second)
			} else {
				break
			}
		}
	}
	return nil
}
