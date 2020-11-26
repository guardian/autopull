package main

import (
	"flag"
	"fmt"
	"github.com/guardian/autopull/communicator"
	"github.com/guardian/autopull/config"
	"github.com/guardian/autopull/downloadmanager"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func enqueueDownloads(entriesListPtr *[]communicator.ArchiveEntryDownloadSynopsis, mgr downloadmanager.DownloadManager) {
	for _, ent := range *entriesListPtr {
		mgr.Enqueue(ent)
	}
}

func ExitPause(noWait bool, exitCode int) {
	if !noWait {
		print("Press ENTER to close...")
		fmt.Scanln()
	}
	os.Exit(exitCode)
}

func main() {
	log.Printf("autopull v0.1 Andy Gallagher. https://github.com/guardian/autopull")
	exePath, pathErr := os.Executable()
	var myPath string
	if pathErr != nil {
		log.Printf("could not get executable's path: %s", pathErr)
		myPath, _ = os.Getwd()
	} else {
		myPath = filepath.Dir(exePath)
	}

	configFilePtr := flag.String("config", filepath.Join(myPath, "autopull.yaml"), "Path to a yaml config file")
	downloadPathPtr := flag.String("to", "", "Download path, overriding the default value in the config file")
	flag.Parse()

	if configFilePtr == nil {
		log.Printf("ERROR main You must specify a config file with the --config argument")
		ExitPause(false, 2)
	}

	configuration, configErr := config.LoadConfig(*configFilePtr)
	if configErr != nil {
		log.Printf("ERROR main could not load config: %s", configErr)
		var nowait bool
		if configuration == nil {
			nowait = false
		} else {
			nowait = configuration.NoWait
		}
		ExitPause(nowait, 3)
	}

	vaultdoorUrl, parseErr := url.Parse(configuration.VaultDoorUri)
	if parseErr != nil {
		log.Printf("ERROR main could not parse VaultDoor uri %s: %s", configuration.VaultDoorUri, parseErr)
		ExitPause(configuration.NoWait, 4)
	}

	archivehunterUrl, parseErr := url.Parse(configuration.ArchiveHunterUri)
	if parseErr != nil {
		log.Printf("ERROR main could not parse ArchiveHunter uri %s: %s", configuration.ArchiveHunterUri, parseErr)
		ExitPause(configuration.NoWait, 4)
	}

	if len(os.Args) < 2 {
		log.Printf("ERROR main You must specify a download token as the first positional argument")
		ExitPause(configuration.NoWait, 1)
	}

	if os.Args[1] == "" {
		log.Printf("ERROR main You must specify a download or custom uri as the first positional argument")
		ExitPause(configuration.NoWait, 1)
	}

	var downloadToken config.DownloadTokenUri
	if strings.Contains(os.Args[1], ":") {
		var parseErr error
		downloadToken, parseErr = config.ParseArchiveHunterUri(os.Args[1])
		if parseErr != nil {
			log.Printf("ERROR main provided URI was not properly formed: %s", parseErr)
			ExitPause(configuration.NoWait, 5)
		}
		if !downloadToken.ValidateVaultDoor() && !downloadToken.ValidateArchiveHunter() {
			log.Printf("ERROR main parsed custom URI but it is not valid for VaultDoor nor ArchiveHunter")
			ExitPause(configuration.NoWait, 5)
		}
	} else {
		downloadToken = config.DownloadTokenUri{
			Proto:   "archivehunter",
			Subtype: "vaultdownload",
			Token:   os.Args[1],
		}
	}

	log.Printf("INFO main Download token is %s", downloadToken)

	var commType communicator.CommunicatorType
	if downloadToken.ValidateVaultDoor() {
		commType = communicator.VaultDoor
	} else if downloadToken.ValidateArchiveHunter() {
		commType = communicator.ArchiveHunter
	}

	comm := communicator.Communicator{VaultDoorUri: *vaultdoorUrl, ArchiveHunterUri: *archivehunterUrl, Type: commType}

	downloadInfo, redeemErr := comm.RedeemToken(downloadToken, 1)
	if redeemErr != nil {
		log.Printf("ERROR main could not redeem download token: %s", redeemErr)
		ExitPause(configuration.NoWait, 5)
	}

	//spew.Dump(downloadInfo)

	totalFiles, totalBytes := downloadInfo.TotalUpEntries()
	log.Printf("INFO main Will try to download a total of %d files totalling %s", totalFiles, FormatByteSize(totalBytes, 0))

	threadCount := configuration.DownloadThreads
	if threadCount == 0 {
		threadCount = 5
	}

	dlQueueBufferSize := configuration.QueueBufferSize
	if dlQueueBufferSize == 0 {
		dlQueueBufferSize = 10
	}

	var downloadPath string
	if downloadPathPtr != nil && *downloadPathPtr != "" {
		downloadPath = *downloadPathPtr
	} else if configuration.DownloadPath != "" {
		downloadPath = configuration.DownloadPath
	} else {
		log.Printf("ERROR main no download path has been set. Try setting `download_path: yourpath` in the settings file")
		ExitPause(configuration.NoWait, 7)
	}

	mgr := downloadmanager.NewDownloadManager(&comm, downloadInfo.RetrievalToken, threadCount, dlQueueBufferSize, downloadPath, configuration.AllowOverwrite)

	initErr := mgr.Init()
	if initErr != nil {
		log.Printf("ERROR main Could not initialise download manager: %s", initErr)
		ExitPause(configuration.NoWait, 6)
	}

	time.Sleep(1 * time.Second)
	enqueueDownloads(&downloadInfo.Entries, mgr)

	log.Printf("DEBUG main enqueued items, waiting for download threads")
	time.Sleep(5 * time.Second)
	mgr.Shutdown(true)
	ExitPause(configuration.NoWait, 0)
}
