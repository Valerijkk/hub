package intranet_test
import (
	"github.com/valerii/hub/backend/internal/intranet"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
)

func TestDTO_DisplayAndDirectoryMapping(t *testing.T) {
	bday := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	u := &ent.User{
		ID:          uuid.New(),
		Email:       "alice@hub.local",
		Name:        "Alice Name",
		DisplayName: "",
		Department:  "Engineering",
		Birthday:    &bday,
	}
	dir := toDirectoryUserResponse(u)
	assert.Equal(t, "Alice Name", dir.DisplayName)
	require.NotNil(t, dir.Birthday)
	assert.Equal(t, "1990-05-15", *dir.Birthday)

	u2 := &ent.User{Email: "bob@hub.local"}
	assert.Equal(t, "bob@hub.local", displayUserName(u2))

	u3 := &ent.User{Name: "Bob", Email: "bob@hub.local"}
	assert.Equal(t, "Bob", displayUserName(u3))

	assert.Equal(t, intranet.CoffeePartnerResponse{}, toCoffeePartner(nil))
}
