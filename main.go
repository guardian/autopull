package main

import (
	"flag"
	"github.com/davecgh/go-spew/spew"
	"github.com/guardian/autopull/communicator"
	"github.com/guardian/autopull/config"
	"github.com/guardian/autopull/downloadmanager"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

func enqueueDownloads(entriesListPtr *[]communicator.ArchiveEntryDownloadSynopsis, mgr downloadmanager.DownloadManager) {
	for _, ent := range *entriesListPtr {
		mgr.Enqueue(ent)
	}
}

func main() {
	log.Printf("autopull v0.1 Andy Gallagher. https://github.com/guardian/autopull")
	pwd, _ := os.Getwd()
	configFilePtr := flag.String("config", "autopull.yaml", "Path to a yaml config file")
	downloadPathPtr := flag.String("to", pwd, "Download path")
	allowOverwritePtr := flag.Bool("overwrite", false, "allow overwriting of existing files")
	flag.Parse()

	if configFilePtr == nil {
		log.Printf("ERROR main You must specify a config file with the --config argument")
		os.Exit(2)
	}

	configuration, configErr := config.LoadConfig(*configFilePtr)
	if configErr != nil {
		log.Printf("ERROR main could not load config: %s", configErr)
		os.Exit(3)
	}

	parsedUrl, parseErr := url.Parse(configuration.VaultDoorUri)
	if parseErr != nil {
		log.Printf("ERROR main could not parse server uri %s: %s", configuration.VaultDoorUri, parseErr)
		os.Exit(4)
	}

	if len(os.Args) < 2 {
		log.Printf("ERROR main You must specify a download token as the first positional argument")
		os.Exit(1)
	}

	if os.Args[1] == "" {
		log.Printf("ERROR main You must specify a download or custom uri as the first positional argument")
		os.Exit(1)
	}

	var downloadToken config.DownloadTokenUri
	if strings.Contains(os.Args[1], ":") {
		var parseErr error
		downloadToken, parseErr = config.ParseArchiveHunterUri(os.Args[1])
		if parseErr != nil {
			log.Printf("ERROR main provided URI was not properly formed: %s", parseErr)
			os.Exit(5)
		}
		if !downloadToken.ValidateVaultDoor() {
			log.Printf("ERROR main parsed custom URI but it is not valid for VaultDoor")
			os.Exit(5)
		}
	} else {
		downloadToken = config.DownloadTokenUri{
			Proto:   "archivehunter",
			Subtype: "vaultdownload",
			Token:   os.Args[1],
		}
	}

	log.Printf("INFO main Download token is %s", downloadToken)

	comm := communicator.Communicator{VaultDoorUri: *parsedUrl}

	downloadInfo, redeemErr := comm.RedeemToken(downloadToken.Token, 1)
	if redeemErr != nil {
		log.Printf("ERROR main could not redeem download token: %s", redeemErr)
		os.Exit(5)
	}

	spew.Dump(downloadInfo)

	totalFiles, totalBytes := downloadInfo.TotalUpEntries()
	log.Printf("INFO main Will try to download a total of %d files totalling %s", totalFiles, FormatByteSize(totalBytes, 0))

	mgr := downloadmanager.NewDownloadManager(&comm, downloadInfo.RetrievalToken, 1, 10, *downloadPathPtr, *allowOverwritePtr)

	initErr := mgr.Init()
	if initErr != nil {
		log.Printf("ERROR main Could not initialise download manager: %s", initErr)
		os.Exit(6)
	}

	time.Sleep(1 * time.Second)
	enqueueDownloads(&downloadInfo.Entries, mgr)

	log.Printf("DEBUG main enqueued items, waiting for download threads")
	time.Sleep(5 * time.Second)
	mgr.Shutdown(true)
}
