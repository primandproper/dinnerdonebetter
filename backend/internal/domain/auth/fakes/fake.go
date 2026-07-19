package fakes

import (
	"fmt"
	"math"
	"time"

	"github.com/primandproper/platform-go/v5/identifiers"

	fake "github.com/brianvoe/gofakeit/v7"
)

func init() {
	if err := fake.Seed(time.Now().UnixNano()); err != nil {
		panic(err)
	}
}

// BuildFakeID builds a fake ID.
func BuildFakeID() string {
	return identifiers.New()
}

func buildFakeTOTPToken() string {
	return fmt.Sprintf("%d%s", fake.Number(0, 9), fake.Zip())
}

// BuildFakeTime builds a fake time.
func BuildFakeTime() time.Time {
	return fake.Date().Add(0).Truncate(time.Second).UTC()
}

func buildFakeNumber() float64 {
	return math.Round(float64((fake.Number(101, math.MaxInt8-1) * 100) / 100))
}

// buildUniqueString builds a fake string.
func buildUniqueString() string {
	return fake.LoremIpsumSentence(7)
}

// buildFakePassword builds a fake password.
func buildFakePassword() string {
	return fake.Password(true, true, true, true, false, 32)
}
