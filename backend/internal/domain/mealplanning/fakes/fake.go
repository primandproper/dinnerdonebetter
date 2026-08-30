package fakes

import (
	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/pointer"
)

const exampleQuantity = fake.DefaultPageSize

// The four builders below exist because this domain is full of flattened ranges: a
// minimum and a maximum in two columns rather than one range type. BuildFakeRecord
// fills each of a pair independently, which produces a maximum below its minimum half
// the time — a range nothing can satisfy, and one every scaling calculation reads.

// BuildFakeOptionalFloat32MinMax returns a fake (*float32, *float32) pair for flattened Min/Max fields.
func BuildFakeOptionalFloat32MinMax() (minimum, maximum *float32) {
	m := float32(fake.BuildFakeNumber())

	return &m, pointer.To(float32(fake.BuildFakeNumber()) + m)
}

// BuildFakeOptionalUint32MinMax returns a fake (*uint32, *uint32) pair for flattened Min/Max fields.
func BuildFakeOptionalUint32MinMax() (minimum, maximum *uint32) {
	m := uint32(fake.BuildFakeNumber())

	return &m, pointer.To(uint32(fake.BuildFakeNumber()) + m)
}

// BuildFakeFloat32WithOptionalMax returns a (float32, *float32) pair: required min + optional max.
func BuildFakeFloat32WithOptionalMax() (minimum float32, maximum *float32) {
	minimum = float32(fake.BuildFakeNumber())

	return minimum, pointer.To(float32(fake.BuildFakeNumber()) + minimum)
}

// BuildFakeUint32WithOptionalMax returns a (uint32, *uint32) pair: required min + optional max.
func BuildFakeUint32WithOptionalMax() (minimum uint32, maximum *uint32) {
	minimum = uint32(fake.BuildFakeNumber())

	return minimum, pointer.To(uint32(fake.BuildFakeNumber()) + minimum)
}

// BuildFakeUint16WithOptionalMax returns a (uint16, *uint16) pair: required min + optional max.
func BuildFakeUint16WithOptionalMax() (minimum uint16, maximum *uint16) {
	minimum = uint16(fake.BuildFakeNumber())

	return minimum, pointer.To(uint16(fake.BuildFakeNumber()) + minimum)
}
