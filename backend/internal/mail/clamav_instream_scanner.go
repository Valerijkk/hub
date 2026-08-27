package mail

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	clamAVINSTREAMAddressMaxBytes  = 255
	clamAVINSTREAMHostMaxBytes     = 253
	clamAVINSTREAMDialTimeout      = 3 * time.Second
	clamAVINSTREAMIOTimeout        = 45 * time.Second
	clamAVINSTREAMChunkBytes       = 32 * 1024
	clamAVINSTREAMVerdictMaxBytes  = 1024
	clamAVINSTREAMCommand          = "zINSTREAM\x00"
	clamAVINSTREAMPingCommand      = "zPING\x00"
	clamAVINSTREAMPingResponse     = "PONG\x00"
	clamAVINSTREAMCleanResponse    = "stream: OK"
	clamAVINSTREAMRejectedSuffix   = " FOUND"
	clamAVINSTREAMResponsePrefix   = "stream: "
)

// ClamAVINSTREAMScanner opens one isolated ClamAV INSTREAM connection per attachment scan.
type ClamAVINSTREAMScanner struct {
	address string
}

var _ MailboxAttachmentMalwareScanner = (*ClamAVINSTREAMScanner)(nil)

// NewClamAVINSTREAMScanner constructs a scanner for one bounded host:port endpoint.
func NewClamAVINSTREAMScanner(address string) (*ClamAVINSTREAMScanner, error) {
	normalized, ok := clamAVINSTREAMAddress(address)
	if !ok {
		return nil, mailboxAttachmentScanError(MailboxAttachmentScanErrorInvalid)
	}
	return &ClamAVINSTREAMScanner{address: normalized}, nil
}

// Begin starts a new ClamAV INSTREAM session on a fresh TCP connection.
func (scanner *ClamAVINSTREAMScanner) Begin(ctx context.Context) (MailboxAttachmentMalwareScanSession, error) {
	connection, deadline, err := scanner.connect(ctx)
	if err != nil {
		return nil, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	if !clamAVINSTREAMWriteFull(connection.conn, []byte(clamAVINSTREAMCommand)) {
		_ = connection.Close()
		return nil, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}

	session := &clamAVINSTREAMSession{
		connection: connection,
		deadline:   deadline,
	}
	session.contextStop = context.AfterFunc(ctx, func() {
		_ = connection.Close()
	})
	session.abortTimer = time.AfterFunc(clamAVINSTREAMIOTimeout, func() {
		session.abort()
	})
	return session, nil
}

// Ping checks the scanner endpoint without submitting attachment content.
func (scanner *ClamAVINSTREAMScanner) Ping(ctx context.Context) error {
	connection, _, err := scanner.connect(ctx)
	if err != nil {
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	stop := context.AfterFunc(ctx, func() {
		_ = connection.Close()
	})
	defer stop()

	if !clamAVINSTREAMWriteFull(connection.conn, []byte(clamAVINSTREAMPingCommand)) {
		_ = connection.Close()
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	response := make([]byte, len(clamAVINSTREAMPingResponse))
	if _, err := io.ReadFull(connection.conn, response); err != nil || string(response) != clamAVINSTREAMPingResponse {
		_ = connection.Close()
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	if ctx.Err() != nil {
		_ = connection.Close()
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	if err := connection.Close(); err != nil {
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	return nil
}

func (scanner *ClamAVINSTREAMScanner) connect(ctx context.Context) (*clamAVINSTREAMConnection, time.Time, error) {
	if scanner == nil || ctx == nil || scanner.address == "" || ctx.Err() != nil {
		return nil, time.Time{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorInvalid)
	}
	dialer := net.Dialer{Timeout: clamAVINSTREAMDialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", scanner.address)
	if err != nil || conn == nil {
		if conn != nil {
			_ = conn.Close()
		}
		return nil, time.Time{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	connection := &clamAVINSTREAMConnection{conn: conn}
	deadline := time.Now().Add(clamAVINSTREAMIOTimeout)
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
		deadline = contextDeadline
	}
	if !deadline.After(time.Now()) || conn.SetDeadline(deadline) != nil {
		_ = connection.Close()
		return nil, time.Time{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	return connection, deadline, nil
}

func clamAVINSTREAMAddress(address string) (string, bool) {
	if address == "" || len(address) > clamAVINSTREAMAddressMaxBytes || strings.TrimSpace(address) != address || clamAVINSTREAMContainsControl(address) {
		return "", false
	}
	host, portText, err := net.SplitHostPort(address)
	if err != nil || host == "" || len(host) > clamAVINSTREAMHostMaxBytes || strings.TrimSpace(host) != host || clamAVINSTREAMContainsControl(host) || strings.Contains(host, "://") || strings.ContainsAny(host, "/\\@?#") {
		return "", false
	}
	if portText == "" || len(portText) > 5 {
		return "", false
	}
	for _, value := range portText {
		if value < '0' || value > '9' {
			return "", false
		}
	}
	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil || port < 1 || port > 65535 {
		return "", false
	}
	return net.JoinHostPort(host, strconv.FormatUint(port, 10)), true
}

func clamAVINSTREAMContainsControl(value string) bool {
	for _, character := range value {
		if unicode.IsControl(character) || unicode.IsSpace(character) {
			return true
		}
	}
	return false
}

type clamAVINSTREAMConnection struct {
	conn     net.Conn
	closeOnce sync.Once
	closeErr  error
}

func (connection *clamAVINSTREAMConnection) Close() error {
	if connection == nil || connection.conn == nil {
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	connection.closeOnce.Do(func() {
		connection.closeErr = connection.conn.Close()
	})
	return connection.closeErr
}

type clamAVINSTREAMSession struct {
	mu          sync.Mutex
	connection  *clamAVINSTREAMConnection
	deadline    time.Time
	contextStop func() bool
	abortTimer  *time.Timer
	totalBytes  int64
	closed      bool
	failed      bool
	verdictRead bool
}

func (session *clamAVINSTREAMSession) Write(input []byte) (int, error) {
	if session == nil {
		return 0, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.failed || session.closed || session.verdictRead || session.connection == nil || session.connection.conn == nil {
		session.failLocked()
		return 0, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	if len(input) == 0 {
		return 0, nil
	}
	if int64(len(input)) > MaxEngineAttachmentBytes-session.totalBytes {
		session.failLocked()
		return 0, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	for remaining := input; len(remaining) > 0; {
		chunkBytes := len(remaining)
		if chunkBytes > clamAVINSTREAMChunkBytes {
			chunkBytes = clamAVINSTREAMChunkBytes
		}
		if !clamAVINSTREAMWriteFrame(session.connection.conn, remaining[:chunkBytes]) {
			session.failLocked()
			return 0, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
		}
		remaining = remaining[chunkBytes:]
	}
	session.totalBytes += int64(len(input))
	return len(input), nil
}

func (session *clamAVINSTREAMSession) Close() error {
	if session == nil {
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.failed || session.closed || session.verdictRead || session.connection == nil || session.connection.conn == nil {
		session.failLocked()
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	var terminator [4]byte
	if !clamAVINSTREAMWriteFull(session.connection.conn, terminator[:]) {
		session.failLocked()
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	session.closed = true
	return nil
}

func (session *clamAVINSTREAMSession) Verdict(ctx context.Context) (MailboxAttachmentMalwareVerdict, error) {
	if session == nil {
		return "", mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if ctx == nil || ctx.Err() != nil || session.failed || !session.closed || session.verdictRead || session.connection == nil || session.connection.conn == nil || !session.deadline.After(time.Now()) {
		session.failLocked()
		return "", mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	session.verdictRead = true
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(session.deadline) {
		if !contextDeadline.After(time.Now()) || session.connection.conn.SetDeadline(contextDeadline) != nil {
			session.failLocked()
			return "", mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
		}
	}
	stop := context.AfterFunc(ctx, func() {
		_ = session.connection.Close()
	})
	defer stop()

	response, ok := clamAVINSTREAMReadVerdict(session.connection.conn)
	if !ok {
		session.failLocked()
		return "", mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	if ctx.Err() != nil {
		session.failLocked()
		return "", mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	verdict, ok := clamAVINSTREAMParseVerdict(response)
	if !ok {
		session.failLocked()
		return "", mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	session.stopAbortLocked()
	session.stopContextLocked()
	if err := session.connection.Close(); err != nil {
		return "", mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	return verdict, nil
}

func (session *clamAVINSTREAMSession) failLocked() {
	if session == nil {
		return
	}
	session.failed = true
	session.stopAbortLocked()
	session.stopContextLocked()
	if session.connection != nil {
		_ = session.connection.Close()
	}
}

func (session *clamAVINSTREAMSession) abort() {
	if session == nil {
		return
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.failed || session.verdictRead {
		return
	}
	session.failLocked()
}

func (session *clamAVINSTREAMSession) stopAbortLocked() {
	if session == nil || session.abortTimer == nil {
		return
	}
	_ = session.abortTimer.Stop()
	session.abortTimer = nil
}

func (session *clamAVINSTREAMSession) stopContextLocked() {
	if session == nil || session.contextStop == nil {
		return
	}
	_ = session.contextStop()
	session.contextStop = nil
}

func clamAVINSTREAMWriteFrame(conn net.Conn, chunk []byte) bool {
	if conn == nil || len(chunk) == 0 || len(chunk) > clamAVINSTREAMChunkBytes {
		return false
	}
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(chunk)))
	return clamAVINSTREAMWriteFull(conn, length[:]) && clamAVINSTREAMWriteFull(conn, chunk)
}

func clamAVINSTREAMWriteFull(conn net.Conn, value []byte) bool {
	if conn == nil || len(value) == 0 {
		return false
	}
	for remaining := value; len(remaining) > 0; {
		written, err := conn.Write(remaining)
		if err != nil || written <= 0 || written > len(remaining) {
			return false
		}
		remaining = remaining[written:]
	}
	return true
}

func clamAVINSTREAMReadVerdict(conn net.Conn) (string, bool) {
	if conn == nil {
		return "", false
	}
	response := make([]byte, 0, clamAVINSTREAMVerdictMaxBytes)
	for len(response) < clamAVINSTREAMVerdictMaxBytes {
		var single [1]byte
		if _, err := io.ReadFull(conn, single[:]); err != nil {
			return "", false
		}
		if single[0] == 0 {
			return string(response), true
		}
		response = append(response, single[0])
	}
	return "", false
}

func clamAVINSTREAMParseVerdict(response string) (MailboxAttachmentMalwareVerdict, bool) {
	if response == clamAVINSTREAMCleanResponse {
		return MailboxAttachmentMalwareVerdictClean, true
	}
	if !strings.HasPrefix(response, clamAVINSTREAMResponsePrefix) || !strings.HasSuffix(response, clamAVINSTREAMRejectedSuffix) {
		return "", false
	}
	signature := strings.TrimSuffix(strings.TrimPrefix(response, clamAVINSTREAMResponsePrefix), clamAVINSTREAMRejectedSuffix)
	if strings.TrimSpace(signature) == "" || len(signature) > clamAVINSTREAMVerdictMaxBytes-len(clamAVINSTREAMResponsePrefix)-len(clamAVINSTREAMRejectedSuffix) {
		return "", false
	}
	return MailboxAttachmentMalwareVerdictRejected, true
}
