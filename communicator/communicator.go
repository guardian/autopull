package communicator

import "net/url"

type CommunicatorType int

const (
	VaultDoor CommunicatorType = iota
	ArchiveHunter
)

type Communicator struct {
	VaultDoorUri     url.URL
	ArchiveHunterUri url.URL
	Type             CommunicatorType
}

/**
return the URL relevant for the communicator tyoe
*/
func (comm *Communicator) GetActiveUrl() *url.URL {
	switch comm.Type {
	case VaultDoor:
		return &comm.VaultDoorUri
	case ArchiveHunter:
		return &comm.ArchiveHunterUri
	default:
		return nil
	}
}
