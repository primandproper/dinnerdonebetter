package fakes

import (
	"math"
	"time"

	"github.com/primandproper/platform-go/v11/identifiers"
	"github.com/primandproper/platform-go/v11/pointer"

	fake "github.com/brianvoe/gofakeit/v7"
)

func init() {
	if err := fake.Seed(time.Now().UnixNano()); err != nil {
		panic(err)
	}
}

const (
	exampleQuantity = 3
)

// BuildFakeID builds a fake ID.
func BuildFakeID() string {
	return identifiers.New()
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

// BuildFakeOptionalFloat32MinMax returns a fake (*float32, *float32) pair for flattened Min/Max fields.
func BuildFakeOptionalFloat32MinMax() (minimum, maximum *float32) {
	m := float32(buildFakeNumber())
	maximum = pointer.To(float32(buildFakeNumber()) + m)
	return &m, maximum
}

// BuildFakeOptionalUint32MinMax returns a fake (*uint32, *uint32) pair for flattened Min/Max fields.
func BuildFakeOptionalUint32MinMax() (minimum, maximum *uint32) {
	m := uint32(buildFakeNumber())
	maximum = pointer.To(uint32(buildFakeNumber()) + m)
	return &m, maximum
}

// BuildFakeFloat32WithOptionalMax returns a (float32, *float32) pair: required min + optional max.
func BuildFakeFloat32WithOptionalMax() (minimum float32, maximum *float32) {
	minimum = float32(buildFakeNumber())
	maximum = pointer.To(float32(buildFakeNumber()) + minimum)
	return minimum, maximum
}

// BuildFakeUint32WithOptionalMax returns a (uint32, *uint32) pair: required min + optional max.
func BuildFakeUint32WithOptionalMax() (minimum uint32, maximum *uint32) {
	minimum = uint32(buildFakeNumber())
	maximum = pointer.To(uint32(buildFakeNumber()) + minimum)
	return minimum, maximum
}

// BuildFakeUint16WithOptionalMax returns a (uint16, *uint16) pair: required min + optional max.
func BuildFakeUint16WithOptionalMax() (minimum uint16, maximum *uint16) {
	minimum = uint16(buildFakeNumber())
	maximum = pointer.To(uint16(buildFakeNumber()) + minimum)
	return minimum, maximum
}
