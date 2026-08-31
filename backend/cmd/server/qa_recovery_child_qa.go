//go:build qa && !release && linux

package main

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"sync"

	"golang.org/x/sys/unix"
)

const (
	qaRecoveryChildControlFD             = 3
	qaRecoveryChildBootTokenBytes        = 32
	qaRecoveryChildRecordMaxBytes        = 128
	qaRecoveryChildControlFDEnvironment  = "HUB_QA_CHILD_CONTROL_FD"
	qaRecoveryChildGenerationEnvironment = "HUB_QA_CHILD_GENERATION"
	qaRecoveryChildModeEnvironment       = "HUB_QA_CHILD_MODE"
)

var (
	errQARecoveryChildInput           = errors.New("qa recovery child input is invalid")
	errQARecoveryChildDescriptor      = errors.New("qa recovery child descriptor is invalid")
	errQARecoveryChildEntropy         = errors.New("qa recovery child entropy is unavailable")
	errQARecoveryChildAnnounce        = errors.New("qa recovery child announcement failed")
	errQARecoveryChildRestart          = errors.New("qa recovery child restart request failed")
	errQARecoveryChildRestartForbidden = errors.New("qa recovery child restart is forbidden")
)

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
	mu               sync.Mutex
	mode             qaRecoveryChildMode
	pid              int
	generation       uint64
	controlFD        int
	bootToken        [qaRecoveryChildBootTokenBytes]byte
	evidence         qaRecoveryChildEvidence
	announceAttempted bool
	announced        bool
	restartAttempted bool
	restartRequested bool
}

func newQARecoveryChild(lookup func(string) (string, bool)) (*qaRecoveryChild, error) {
	if lookup == nil {
		return nil, errQARecoveryChildInput
	}

	fdText, fdSet := lookup(qaRecoveryChildControlFDEnvironment)
	generationText, generationSet := lookup(qaRecoveryChildGenerationEnvironment)
	modeText, modeSet := lookup(qaRecoveryChildModeEnvironment)
	if !fdSet && !generationSet && !modeSet {
		return &qaRecoveryChild{mode: qaRecoveryChildModeUnsupervised}, nil
	}
	if !fdSet || !generationSet || !modeSet || fdText != strconv.Itoa(qaRecoveryChildControlFD) {
		return nil, errQARecoveryChildInput
	}

	generation, valid := parseQARecoveryChildGeneration(generationText)
	if !valid {
		return nil, errQARecoveryChildInput
	}

	mode, valid := parseQARecoveryChildMode(modeText, generation)
	if !valid {
		return nil, errQARecoveryChildInput
	}

	pid := os.Getpid()
	if pid <= 1 {
		return nil, errQARecoveryChildInput
	}

	var stat unix.Stat_t
	if err := unix.Fstat(qaRecoveryChildControlFD, &stat); err != nil || stat.Mode&unix.S_IFMT != unix.S_IFIFO {
		return nil, errQARecoveryChildDescriptor
	}
	flags, err := unix.FcntlInt(uintptr(qaRecoveryChildControlFD), unix.F_GETFL, 0)
	if err != nil || flags&unix.O_ACCMODE != unix.O_WRONLY {
		return nil, errQARecoveryChildDescriptor
	}
	if _, err := unix.FcntlInt(uintptr(qaRecoveryChildControlFD), unix.F_SETFD, unix.FD_CLOEXEC); err != nil {
		return nil, errQARecoveryChildDescriptor
	}

	child := &qaRecoveryChild{
		mode:      mode,
		pid:       pid,
		generation: generation,
		controlFD: qaRecoveryChildControlFD,
	}
	if _, err := cryptorand.Read(child.bootToken[:]); err != nil {
		zeroQARecoveryChildBytes(child.bootToken[:])
		return nil, errQARecoveryChildEntropy
	}
	child.evidence = qaRecoveryChildEvidence{
		pid:        child.pid,
		generation: child.generation,
		bootSHA256: sha256.Sum256(child.bootToken[:]),
	}
	return child, nil
}

func parseQARecoveryChildGeneration(value string) (uint64, bool) {
	if value == "" {
		return 0, false
	}
	for index := range value {
		if value[index] < '0' || value[index] > '9' {
			return 0, false
		}
	}
	generation, err := strconv.ParseUint(value, 10, 64)
	return generation, err == nil && generation > 0 && strconv.FormatUint(generation, 10) == value
}

func parseQARecoveryChildMode(value string, generation uint64) (qaRecoveryChildMode, bool) {
	switch value {
	case "normal":
		return qaRecoveryChildModeNormal, generation == 1
	case "recovery":
		return qaRecoveryChildModeRecovery, generation == 2
	default:
		return qaRecoveryChildModeUnsupervised, false
	}
}

func (child *qaRecoveryChild) supervised() bool {
	if child == nil {
		return false
	}
	child.mu.Lock()
	defer child.mu.Unlock()
	return child.mode == qaRecoveryChildModeNormal || child.mode == qaRecoveryChildModeRecovery
}

func (child *qaRecoveryChild) unsupervised() bool {
	if child == nil {
		return false
	}
	child.mu.Lock()
	defer child.mu.Unlock()
	return child.mode == qaRecoveryChildModeUnsupervised
}

func (child *qaRecoveryChild) normal() bool {
	if child == nil {
		return false
	}
	child.mu.Lock()
	defer child.mu.Unlock()
	return child.mode == qaRecoveryChildModeNormal
}

func (child *qaRecoveryChild) recovery() bool {
	if child == nil {
		return false
	}
	child.mu.Lock()
	defer child.mu.Unlock()
	return child.mode == qaRecoveryChildModeRecovery
}

func (child *qaRecoveryChild) routeEvidence() (qaRecoveryChildEvidence, bool) {
	if child == nil {
		return qaRecoveryChildEvidence{}, false
	}
	child.mu.Lock()
	defer child.mu.Unlock()
	if child.mode == qaRecoveryChildModeUnsupervised || (child.announceAttempted && !child.announced) {
		return qaRecoveryChildEvidence{}, false
	}
	return child.evidence, true
}

func (child *qaRecoveryChild) announce() error {
	if child == nil {
		return errQARecoveryChildAnnounce
	}
	child.mu.Lock()
	defer child.mu.Unlock()
	if child.mode == qaRecoveryChildModeUnsupervised || child.announceAttempted || child.announced {
		return errQARecoveryChildAnnounce
	}
	child.announceAttempted = true

	record, valid := child.record("CHILD")
	defer zeroQARecoveryChildBytes(record)
	if !valid {
		zeroQARecoveryChildBytes(child.bootToken[:])
		return errQARecoveryChildAnnounce
	}
	defer zeroQARecoveryChildBytes(child.bootToken[:])

	written, err := unix.Write(child.controlFD, record)
	if err != nil || written != len(record) {
		return errQARecoveryChildAnnounce
	}
	child.announced = true
	return nil
}

func (child *qaRecoveryChild) requestRestart(rootCancel func()) error {
	if child == nil || rootCancel == nil {
		return errQARecoveryChildRestart
	}
	child.mu.Lock()
	if child.mode != qaRecoveryChildModeNormal || child.generation != 1 || !child.announced || child.restartAttempted || child.restartRequested {
		child.mu.Unlock()
		return errQARecoveryChildRestartForbidden
	}
	child.restartAttempted = true

	record, valid := child.record("RESTART")
	if !valid {
		child.mu.Unlock()
		return errQARecoveryChildRestart
	}
	defer zeroQARecoveryChildBytes(record)

	written, err := unix.Write(child.controlFD, record)
	if err != nil || written != len(record) {
		child.mu.Unlock()
		return errQARecoveryChildRestart
	}
	child.restartRequested = true
	child.mu.Unlock()
	rootCancel()
	return nil
}

func (child *qaRecoveryChild) record(kind string) ([]byte, bool) {
	if child == nil || child.pid <= 1 {
		return nil, false
	}
	generation := strconv.FormatUint(child.generation, 10)
	pid := strconv.Itoa(child.pid)

	switch kind {
	case "CHILD":
		mode, valid := child.mode.text()
		if !valid {
			return nil, false
		}
		encodedToken := make([]byte, hex.EncodedLen(len(child.bootToken)))
		defer zeroQARecoveryChildBytes(encodedToken)
		hex.Encode(encodedToken, child.bootToken[:])
		record := make([]byte, 0, qaRecoveryChildRecordMaxBytes)
		record = append(record, "CHILD 1 "...)
		record = append(record, pid...)
		record = append(record, ' ')
		record = append(record, generation...)
		record = append(record, ' ')
		record = append(record, mode...)
		record = append(record, ' ')
		record = append(record, encodedToken...)
		record = append(record, '\n')
		return record, len(record) <= qaRecoveryChildRecordMaxBytes
	case "RESTART":
		if child.mode != qaRecoveryChildModeNormal || child.generation != 1 {
			return nil, false
		}
		record := make([]byte, 0, qaRecoveryChildRecordMaxBytes)
		record = append(record, "RESTART 1 "...)
		record = append(record, pid...)
		record = append(record, ' ')
		record = append(record, generation...)
		record = append(record, '\n')
		return record, len(record) <= qaRecoveryChildRecordMaxBytes
	default:
		return nil, false
	}
}

func (mode qaRecoveryChildMode) text() (string, bool) {
	switch mode {
	case qaRecoveryChildModeNormal:
		return "normal", true
	case qaRecoveryChildModeRecovery:
		return "recovery", true
	default:
		return "", false
	}
}

func zeroQARecoveryChildBytes(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
