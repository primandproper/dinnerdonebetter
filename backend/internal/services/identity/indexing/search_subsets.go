package indexing

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"

	textsearch "github.com/primandproper/platform-go/v8/search/text"
)

type UserTextSearcher textsearch.Index[UserSearchSubset]

// UserSearchSubset represents the subset of values suitable to index for search.
type UserSearchSubset struct {
	_ struct{} `json:"-"`

	ID           string `json:"id,omitempty"`
	Username     string `json:"username,omitempty"`
	FirstName    string `json:"firstName,omitempty"`
	LastName     string `json:"lastName,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
}

// ConvertUserToUserSearchSubset converts a User to a UserSearchSubset.
func ConvertUserToUserSearchSubset(x *identity.User) *UserSearchSubset {
	return &UserSearchSubset{
		ID:           x.ID,
		Username:     x.Username,
		FirstName:    x.FirstName,
		LastName:     x.LastName,
		EmailAddress: x.EmailAddress,
	}
}
