package fakes

import (
	"fmt"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// buildFakeTOTPToken builds a token of the length the domain validates: six digits.
func buildFakeTOTPToken() string {
	return fmt.Sprintf("%d%s", gofakeit.Number(0, 9), gofakeit.Zip())
}
