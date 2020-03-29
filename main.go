package main

import (
	"flag"
	"github.com/davecgh/go-spew/spew"
	"github.com/guardian/autopull/communicator"
	"github.com/guardian/autopull/config"
	"log"
	"net/url"
	"os"
)

func main() {
	log.Printf("autopull v0.1 Andy Gallagher. https://github.com/guardian/autopull")
	configFilePtr := flag.String("config","autopull.yaml","Path to a yaml config file")
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

	if len(os.Args)<2 {
		log.Printf("ERROR main You must specify a download token as the first positional argument")
		os.Exit(1)
	}

	downloadToken := os.Args[1]
	if downloadToken == "" {
		log.Printf("ERROR main You must specify a download token as the first positional argument")
		os.Exit(1)
	}

	log.Printf("INFO main Download token is %s", downloadToken)

	comm := communicator.Communicator{VaultDoorUri:*parsedUrl}

	downloadInfo, redeemErr := comm.RedeemToken(downloadToken,1)
	if redeemErr != nil {
		log.Printf("ERROR main could not redeem download token: %s", redeemErr)
		os.Exit(5)
	}

	spew.Dump(downloadInfo)
}