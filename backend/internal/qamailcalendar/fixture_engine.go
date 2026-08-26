//go:build qa && !release

package qamailcalendar

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"

	"github.com/valerii/hub/backend/internal/mail"
)

var (
	errInboundRequestFixtureEngineInput        = errors.New("qa mail calendar fixture engine input is invalid")
	errInboundRequestFixtureEngineOperation    = errors.New("qa mail calendar fixture engine operation failed")
	errInboundRequestFixtureEngineCompensation = errors.New("qa mail calendar fixture engine compensation failed")
	errInboundRequestFixtureEngineDestroy      = errors.New("qa mail calendar fixture engine cleanup failed")
)

type inboundRequestEngineFixture struct {
	accountID string
	folderID  string
	messageID string
	version   string
}

func seedInboundRequestFixtureEngine(
	ctx context.Context,
	engine *mail.FakeMailEngine,
	plan inboundRequestFixturePlan,
) (inboundRequestEngineFixture, error) {
	if ctx == nil || engine == nil {
		return inboundRequestEngineFixture{}, errInboundRequestFixtureEngineInput
	}
	canonicalPlan, err := newInboundRequestFixturePlan(plan.nonce)
	if err != nil || plan != canonicalPlan {
		return inboundRequestEngineFixture{}, errInvalidFixtureNonce
	}

	accountHash := hex.EncodeToString(plan.accountControlHash[:])
	account, err := engine.EnsureAccount(ctx, mail.AccountSpec{
		ProvisioningKey:   plan.accountProvisioningKey,
		LocalPart:         "owner",
		Domain:            plan.domain,
		DesiredGeneration: 1,
		DesiredHash:       accountHash,
		Enabled:           true,
		Primary: mail.AccountIdentitySpec{
			LocalPart: "owner",
			Domain:    plan.domain,
			Enabled:   true,
		},
	})
	if err != nil {
		return inboundRequestEngineFixture{}, errInboundRequestFixtureEngineOperation
	}

	compensate := func() error {
		if !engine.DeleteAccount(account.AccountID) {
			return errors.Join(errInboundRequestFixtureEngineOperation, errInboundRequestFixtureEngineCompensation)
		}
		return errInboundRequestFixtureEngineOperation
	}
	if account.AccountID == "" ||
		account.State != mail.EngineAccountActive ||
		account.AppliedGeneration != 1 ||
		account.AppliedHash != accountHash ||
		account.PrimaryIdentityID == "" ||
		len(account.AliasIdentityIDs) != 0 {
		return inboundRequestEngineFixture{}, compensate()
	}

	folder, err := engine.CreateFolder(ctx, account.AccountID, mail.FolderSpec{Name: plan.folderName})
	if err != nil {
		return inboundRequestEngineFixture{}, compensate()
	}
	if folder.AccountID != account.AccountID ||
		folder.FolderID == "" ||
		folder.ParentID != "" ||
		folder.Name != plan.folderName ||
		folder.Role != mail.EngineFolderRoleCustom {
		return inboundRequestEngineFixture{}, compensate()
	}

	method := mail.MailCalendarMethodRequest
	receivedAt := plan.fixedAt
	locatorHash := append([]byte(nil), plan.calendarLocatorHash[:]...)
	messageID := "qa-message-" + plan.nonce
	version := "qa-version-" + plan.nonce
	message := mail.EngineMessage{
		Projection: mail.EngineMessageProjection{
			AccountID:               account.AccountID,
			MessageID:               messageID,
			Version:                 version,
			InternetMessageID:       "<qa-inbound-request-" + plan.nonce + "@example.test>",
			From:                    mail.EngineAddress{Email: plan.organizerAddress},
			To:                      []mail.EngineAddress{{Email: plan.ownerAddress}},
			CC:                      []mail.EngineAddress{},
			BCC:                     []mail.EngineAddress{},
			ReplyTo:                 []mail.EngineAddress{},
			References:              []string{},
			Subject:                 plan.subject,
			Snippet:                 plan.snippet,
			ReceivedAt:              &receivedAt,
			CalendarMethod:          &method,
			AuthenticationClass:     mail.MailAuthenticationClassTrustedInternal,
			CalendarPartLocatorHash: locatorHash,
			SizeBytes:               int64(len(plan.snippet)),
			Priority:                mail.EnginePriorityNormal,
			FolderIDs:               []string{folder.FolderID},
			Flags:                   mail.EngineMessageFlags{},
			ReservedKeywords:        []mail.EngineKeywordMarker{},
			Attachments:             []mail.EngineProjectionAttachment{},
			Origin:                  mail.EngineProjectionOriginInboundOrUnknown,
		},
		Body: mail.EngineMessageBody{
			Format:    mail.EngineBodyPlain,
			PlainBody: plan.snippet,
		},
	}
	engine.PutMessage(message)

	readMessage, err := engine.GetMessage(ctx, mail.EngineMessageReadRequest{
		AccountID:       account.AccountID,
		MessageID:       messageID,
		MaxContentBytes: mail.MaxEngineMessageBytes,
	})
	if err != nil ||
		readMessage.Projection.AccountID != account.AccountID ||
		readMessage.Projection.MessageID != messageID ||
		readMessage.Projection.Version != version ||
		readMessage.Projection.InternetMessageID != message.Projection.InternetMessageID ||
		readMessage.Projection.InReplyTo != "" ||
		len(readMessage.Projection.References) != 0 ||
		readMessage.Projection.From != message.Projection.From ||
		len(readMessage.Projection.To) != 1 ||
		readMessage.Projection.To[0] != message.Projection.To[0] ||
		len(readMessage.Projection.CC) != 0 ||
		len(readMessage.Projection.BCC) != 0 ||
		len(readMessage.Projection.ReplyTo) != 0 ||
		readMessage.Projection.Subject != plan.subject ||
		readMessage.Projection.Snippet != plan.snippet ||
		readMessage.Projection.SentAt != nil ||
		readMessage.Projection.ReceivedAt == nil ||
		!readMessage.Projection.ReceivedAt.Equal(plan.fixedAt) ||
		readMessage.Projection.CalendarMethod == nil ||
		*readMessage.Projection.CalendarMethod != method ||
		readMessage.Projection.AuthenticationClass != mail.MailAuthenticationClassTrustedInternal ||
		!bytes.Equal(readMessage.Projection.CalendarPartLocatorHash, plan.calendarLocatorHash[:]) ||
		readMessage.Projection.SizeBytes != int64(len(plan.snippet)) ||
		readMessage.Projection.Priority != mail.EnginePriorityNormal ||
		len(readMessage.Projection.FolderIDs) != 1 ||
		readMessage.Projection.FolderIDs[0] != folder.FolderID ||
		readMessage.Projection.Flags != (mail.EngineMessageFlags{}) ||
		len(readMessage.Projection.ReservedKeywords) != 0 ||
		len(readMessage.Projection.Attachments) != 0 ||
		readMessage.Projection.Origin != mail.EngineProjectionOriginInboundOrUnknown ||
		readMessage.Body.Format != mail.EngineBodyPlain ||
		readMessage.Body.PlainBody != plan.snippet ||
		readMessage.Body.RichBody != "" {
		return inboundRequestEngineFixture{}, compensate()
	}

	calendarText := []byte(plan.calendarText)
	engine.PutCalendarPart(mail.FakeCalendarPart{
		AccountID:   account.AccountID,
		MessageID:   messageID,
		Version:     version,
		LocatorHash: plan.calendarLocatorHash,
		Method:      method,
		ITIP:        calendarText,
	})

	calendarPart, err := engine.ReadCalendarPart(ctx, mail.EngineCalendarPartReadRequest{
		AccountID:   account.AccountID,
		MessageID:   messageID,
		Version:     version,
		LocatorHash: plan.calendarLocatorHash,
		Method:      method,
		MaxBytes:    mail.MaxCalendarPartBytes,
	})
	if err != nil ||
		calendarPart.LocatorHash != plan.calendarLocatorHash ||
		!bytes.Equal(calendarPart.ITIP, []byte(plan.calendarText)) {
		return inboundRequestEngineFixture{}, compensate()
	}

	return inboundRequestEngineFixture{
		accountID: account.AccountID,
		folderID:  folder.FolderID,
		messageID: messageID,
		version:   version,
	}, nil
}

func destroyInboundRequestFixtureEngine(
	engine *mail.FakeMailEngine,
	fixture inboundRequestEngineFixture,
) error {
	if engine == nil || fixture.accountID == "" {
		return errInboundRequestFixtureEngineInput
	}
	if !engine.DeleteAccount(fixture.accountID) {
		return errInboundRequestFixtureEngineDestroy
	}
	return nil
}
