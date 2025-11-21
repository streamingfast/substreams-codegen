package chains

import (
	"regexp"

	registry "github.com/pinax-network/graph-networks-libs/packages/golang/lib"
	networks "github.com/streamingfast/firehose-networks"
)

var solanaNetworkRegexp = regexp.MustCompile(`^solana`)
var nearNetworkRegexp = regexp.MustCompile(`^near`)
var tronNetworkRegexp = regexp.MustCompile(`^tron`)
var stellarNetworkRegexp = regexp.MustCompile(`^stellar`)
var monadNetworkRegexp = regexp.MustCompile(`^monad`)

func SolanaNetworks() []*registry.Network {
	return excludeSolanaAccounts(networks.GetSubstreamsRegistry().Search(solanaNetworkRegexp))
}

func NearNetworks() []*registry.Network {
	return networks.GetSubstreamsRegistry().Search(nearNetworkRegexp)
}

func TronNetworks() []*registry.Network {
	return networks.GetSubstreamsRegistry().Search(tronNetworkRegexp)
}

func StellarNetworks() []*registry.Network {
	return networks.GetSubstreamsRegistry().Search(stellarNetworkRegexp)
}

func MonadNetworks() []*registry.Network {
	return networks.GetSubstreamsRegistry().Search(monadNetworkRegexp)
}

func excludeSolanaAccounts(networks []*registry.Network) []*registry.Network {
	var filteredNetworks []*registry.Network
	for _, net := range networks {
		if net.ID == "solana-accounts" {
			continue
		}
		filteredNetworks = append(filteredNetworks, net)
	}
	return filteredNetworks
}
