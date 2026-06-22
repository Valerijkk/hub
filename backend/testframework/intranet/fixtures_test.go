package intranet_test

import (
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/intranet"
)

type (
	Service                  = intranet.Service
	EntRepository            = intranet.EntRepository
	Repository               = intranet.Repository
	CoffeeNotifier           = intranet.CoffeeNotifier
	NewsPostResponse         = intranet.NewsPostResponse
	NewsListResponse         = intranet.NewsListResponse
	CreateNewsRequest        = intranet.CreateNewsRequest
	UpdateNewsRequest        = intranet.UpdateNewsRequest
	CreateKudoRequest        = intranet.CreateKudoRequest
	KudoResponse             = intranet.KudoResponse
	DirectoryUserResponse    = intranet.DirectoryUserResponse
	CoffeeEnrollmentResponse = intranet.CoffeeEnrollmentResponse
	CoffeePartnerResponse    = intranet.CoffeePartnerResponse
	CoffeeStatsResponse      = intranet.CoffeeStatsResponse
	CoffeeRoundResponse      = intranet.CoffeeRoundResponse
)

var (
	NewService       = intranet.NewService
	NewEntRepository = intranet.NewEntRepository

	ErrNotFound        = intranet.ErrNotFound
	ErrForbidden       = intranet.ErrForbidden
	ErrInvalidInput    = intranet.ErrInvalidInput
	ErrVersionConflict = intranet.ErrVersionConflict
	ErrSelfKudo        = intranet.ErrSelfKudo
	ErrInvalidKudo     = intranet.ErrInvalidKudo
)

func toDirectoryUserResponse(u *ent.User) intranet.DirectoryUserResponse {
	return intranet.ExportToDirectoryUserResponse(u)
}

func displayUserName(u *ent.User) string {
	return intranet.ExportDisplayUserName(u)
}

func toCoffeePartner(u *ent.User) CoffeePartnerResponse {
	return intranet.ExportToCoffeePartner(u)
}

func pairKey(a, b uuid.UUID) string {
	return intranet.ExportPairKey(a, b)
}
