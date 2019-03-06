package bitcoinrpc

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/btcsuite/btcd/chaincfg"
	rpcclient "github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
)

// Bitcoinrpc speaks to btcd and btcwallet
type Bitcoinrpc struct {
	BitcoinUser string
	BitcoinPass string
	Client      *rpcclient.Client
}

// InitRpcConnection creates websocket connection to btcd
func (b *Bitcoinrpc) InitRpcConnection() {
	// must have certificate in workdir. Certificate needed to communicate to btcwallet over websockets.
	certs, err := ioutil.ReadFile("rpc.cert")
	if err != nil {
		log.Fatal(err)
	}

	// @TODO - have config options in env file
	connCfg := &rpcclient.ConnConfig{
		Host:         "btcwallet:18332",
		Endpoint:     "ws",
		User:         b.BitcoinUser,
		Pass:         b.BitcoinPass,
		Certificates: certs,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatalf("could not connect to bitcoin node : %v", err.Error())
	}
	log.Println("connected to bitcoin node")

	b.Client = client
}

// AddAddress adds a bitcoin address to the node to watch
func (b *Bitcoinrpc) AddAddress(address string) error {
	log.Printf("Now watching for address %s", address)

	err := b.Client.ImportAddressRescan(address, "", false)

	if err != nil {
		log.Fatalf("could not import address %s : %v", address, err.Error())
		return fmt.Errorf("could not import address %s", address)
	}

	return nil
}

// GetReceivedAmount returns the amount owned by the address
func (b *Bitcoinrpc) GetReceivedAmount(address string) (int32, error) {
	log.Printf("Getting amount for address %s", address)
	encodedAddress, err := btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	if err != nil {
		log.Fatalf("could not decode address %s : %v", address, err.Error())
		return 0, errors.New("could not get balance")
	}

	// @TODO option to set conf policy
	amount, err := b.Client.GetReceivedByAddressMinConf(encodedAddress, 0)
	if err != nil {
		log.Printf("could not fetch amount for address %s : %v", address, err.Error())
		return 0, errors.New("could not get balance")
	}

	// convert to sats, floats are not nice
	satoshis := amount.ToUnit(btcutil.AmountSatoshi)

	log.Printf("%s owns %v satoshis", address, satoshis)
	return int32(satoshis), nil
}

// IsValidAddress validates a given address
func (b *Bitcoinrpc) IsValidAddress(address string) bool {
	_, err := btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	if err != nil {
		return false
	}
	return true
}
