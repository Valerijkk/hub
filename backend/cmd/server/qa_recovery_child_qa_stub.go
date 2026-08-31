//go:build qa && !release && !linux

package main

import (
	"crypto/sha256"
	"errors"
	"sync"
)

var errQARecoveryChildUnsupportedPlatform = errors.New("qa recovery child is unsupported on this platform")

type qaRecoveryChildMode uint8

const (
	qaRecoveryChildModeUnsupervised qaRecoveryChildMode = iota
	qaRecoveryChildModeNormal
	qaRecoveryChildModeRecovery
)

type qaRecoveryChildEvidence struct {
	pid        int
	generation uint64
	bootSHA256 [sha256.Size]byte
}

type qaRecoveryChild struct {
	mu                sync.Mutex
	mode              qaRecoveryChildMode
	pid               int
	generation        uint64
	controlFD         int
	bootToken         [32]byte
	evidence          qaRecoveryChildEvidence
	announceAttempted bool
	announced         bool
	restartAttempted  bool
	restartRequested  bool
}

func newQARecoveryChild(_ func(string) (string, bool)) (*qaRecoveryChild, error) {
	return nil, errQARecoveryChildUnsupportedPlatform
}

func (child *qaRecoveryChild) supervised() bool {
	return false
}

func (child *qaRecoveryChild) unsupervised() bool {
	return false
}

func (child *qaRecoveryChild) normal() bool {
	return false
}

func (child *qaRecoveryChild) recovery() bool {
	return false
}

func (child *qaRecoveryChild) routeEvidence() (qaRecoveryChildEvidence, bool) {
	return qaRecoveryChildEvidence{}, false
}

func (child *qaRecoveryChild) announce() error {
	return errQARecoveryChildUnsupportedPlatform
}

func (child *qaRecoveryChild) requestRestart(_ func()) error {
	return errQARecoveryChildUnsupportedPlatform
}
