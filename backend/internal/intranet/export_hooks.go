//go:build hubtest

package intranet

import (
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

func ExportMapServiceError(err error) error {
	return mapServiceError(err)
}

func ExportToDirectoryUserResponse(u *ent.User) DirectoryUserResponse {
	return toDirectoryUserResponse(u)
}

func ExportDisplayUserName(u *ent.User) string {
	return displayUserName(u)
}

func ExportToCoffeePartner(u *ent.User) CoffeePartnerResponse {
	return toCoffeePartner(u)
}

func ExportPairKey(a, b uuid.UUID) string {
	return pairKey(a, b)
}

// ExportBuildCoffeeMatches — FR-INTR-10 matching для hubtest.
func ExportBuildCoffeeMatches(ids []uuid.UUID, avoid map[string]struct{}) (pairs [][2]uuid.UUID, triples [][3]uuid.UUID) {
	return buildCoffeeMatches(ids, avoid)
}
