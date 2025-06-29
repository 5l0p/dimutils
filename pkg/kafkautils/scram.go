package kafkautils

import (
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"github.com/xdg-go/scram"
)

// HashGeneratorFcn represents a hash generator function
type HashGeneratorFcn func() hash.Hash

// SHA256 hash generator
var SHA256 HashGeneratorFcn = sha256.New

// SHA512 hash generator  
var SHA512 HashGeneratorFcn = sha512.New

// XDGSCRAMClient implements SCRAM authentication
type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

// Begin starts the SCRAM authentication process
func (x *XDGSCRAMClient) Begin(userName, password, authzID string) (err error) {
	x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.ClientConversation = x.Client.NewConversation()
	return nil
}

// Step performs a step in the SCRAM authentication
func (x *XDGSCRAMClient) Step(challenge string) (response string, err error) {
	response, err = x.ClientConversation.Step(challenge)
	return
}

// Done checks if SCRAM authentication is complete
func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}