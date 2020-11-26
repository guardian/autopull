package config

import (
	"errors"
	"strings"
)

type DownloadTokenUri struct {
	Proto   string //must be "archivehunter"
	Subtype string //expect "vaultdownload" for VaultDoor
	Token   string //long-lived token
}

func ParseArchiveHunterUri(content string) (DownloadTokenUri, error) {
	parts := strings.Split(content, ":")
	if len(parts) != 3 {
		return DownloadTokenUri{}, errors.New("not enough parts to split")
	}

	rtn := DownloadTokenUri{
		Proto:   parts[0],
		Subtype: parts[1],
		Token:   parts[2],
	}
	return rtn, nil
}

func (u DownloadTokenUri) ValidateVaultDoor() bool {
	if u.Proto != "archivehunter" {
		return false
	}
	if u.Subtype != "vaultdownload" {
		return false
	}
	return true
}

func (u DownloadTokenUri) ValidateArchiveHunter() bool {
	if u.Proto != "archivehunter" {
		return false
	}
	if u.Subtype != "bulkdownload" {
		return false
	}
	return true
}
