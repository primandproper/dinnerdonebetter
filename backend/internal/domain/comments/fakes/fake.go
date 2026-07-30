package fakes

import (
	"time"

	"github.com/primandproper/platform-go/v8/identifiers"

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

// BuildFakeTime builds a fake time.
func BuildFakeTime() time.Time {
	return fake.Date().Add(0).Truncate(time.Second).UTC()
}

func buildUniqueString() string {
	return fake.LoremIpsumSentence(7)
}
