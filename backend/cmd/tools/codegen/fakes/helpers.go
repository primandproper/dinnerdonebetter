package main

import "strings"

// helper is one primitive a fakes package may need.
//
// These were duplicated verbatim across all thirteen fakes packages — the same seed, the same
// BuildFakeID, the same off-by-a-comment buildFakeNumber — with the differences between copies
// being accidents rather than decisions. Generating them from one source is most of the reason a
// fakes package no longer has a hand-written file at all.
type helper struct {
	// name is the identifier a package has to mention to get this helper.
	name string
	// source is the declaration, verbatim.
	source string
	// always emits the helper whether or not anything names it.
	always bool
}

// helpers is the catalog, in emission order.
//
// Order is fixed rather than sorted so the generated file reads top-down: the seed, the sizes, the
// primitives, then the compound builders that use them.
var helpers = []helper{
	{
		name:   "init",
		always: true,
		source: `func init() {
	if err := fake.Seed(time.Now().UnixNano()); err != nil {
		panic(err)
	}
}`,
	},
	{
		name: "exampleQuantity",
		source: `// exampleQuantity is how many elements a faked list has.
const exampleQuantity = 3`,
	},
	{
		name: "BuildFakeID",
		source: `// BuildFakeID builds a fake ID.
func BuildFakeID() string {
	return identifiers.New()
}`,
	},
	{
		name: "BuildFakeTime",
		source: `// BuildFakeTime builds a fake time.
func BuildFakeTime() time.Time {
	return fake.Date().Add(0).Truncate(time.Second).UTC()
}`,
	},
	{
		name: "buildFakeNumber",
		source: `// buildFakeNumber builds a fake number comfortably above zero and below every integer
// width the domains use, so that it survives a conversion to any of them and no validator
// mistakes it for unset.
func buildFakeNumber() float64 {
	return math.Round(float64((fake.Number(101, math.MaxInt8-1) * 100) / 100))
}`,
	},
	{
		name: "buildUniqueString",
		source: `// buildUniqueString builds a fake string.
func buildUniqueString() string {
	return fake.LoremIpsumSentence(7)
}`,
	},
	{
		name: "buildFakePassword",
		source: `// buildFakePassword builds a fake password.
func buildFakePassword() string {
	return fake.Password(true, true, true, true, false, 32)
}`,
	},
	{
		name: "buildFakeTOTPToken",
		source: `// buildFakeTOTPToken builds a fake TOTP token.
func buildFakeTOTPToken() string {
	return fmt.Sprintf("%d%s", fake.Number(0, 9), fake.Zip())
}`,
	},
	{
		name: "BuildFakeOptionalFloat32MinMax",
		source: `// BuildFakeOptionalFloat32MinMax returns a fake (*float32, *float32) pair for flattened
// Min/Max fields.
//
// The pair is built together because the maximum has to exceed the minimum: two independent
// fakes for the two fields fail the range validation the domain applies to them.
func BuildFakeOptionalFloat32MinMax() (minimum, maximum *float32) {
	m := float32(buildFakeNumber())
	maximum = pointer.To(float32(buildFakeNumber()) + m)

	return &m, maximum
}`,
	},
	{
		name: "BuildFakeOptionalUint32MinMax",
		source: `// BuildFakeOptionalUint32MinMax returns a fake (*uint32, *uint32) pair for flattened
// Min/Max fields.
func BuildFakeOptionalUint32MinMax() (minimum, maximum *uint32) {
	m := uint32(buildFakeNumber())
	maximum = pointer.To(uint32(buildFakeNumber()) + m)

	return &m, maximum
}`,
	},
	{
		name: "BuildFakeFloat32WithOptionalMax",
		source: `// BuildFakeFloat32WithOptionalMax returns a (float32, *float32) pair: required min plus
// optional max.
func BuildFakeFloat32WithOptionalMax() (minimum float32, maximum *float32) {
	minimum = float32(buildFakeNumber())
	maximum = pointer.To(float32(buildFakeNumber()) + minimum)

	return minimum, maximum
}`,
	},
	{
		name: "BuildFakeUint32WithOptionalMax",
		source: `// BuildFakeUint32WithOptionalMax returns a (uint32, *uint32) pair: required min plus
// optional max.
func BuildFakeUint32WithOptionalMax() (minimum uint32, maximum *uint32) {
	minimum = uint32(buildFakeNumber())
	maximum = pointer.To(uint32(buildFakeNumber()) + minimum)

	return minimum, maximum
}`,
	},
	{
		name: "BuildFakeUint16WithOptionalMax",
		source: `// BuildFakeUint16WithOptionalMax returns a (uint16, *uint16) pair: required min plus
// optional max.
func BuildFakeUint16WithOptionalMax() (minimum uint16, maximum *uint16) {
	minimum = uint16(buildFakeNumber())
	maximum = pointer.To(uint16(buildFakeNumber()) + minimum)

	return minimum, maximum
}`,
	},
}

// renderHelpers emits the helpers a package needs, and only those.
//
// A helper one helper away from a needed one is needed too, so inclusion runs to a fixed point:
// asking for BuildFakeUint16WithOptionalMax asks for buildFakeNumber whether or not anything else
// mentions it.
func renderHelpers(used map[string]struct{}) string {
	wanted := map[string]bool{}

	for changed := true; changed; {
		changed = false

		for _, h := range helpers {
			if wanted[h.name] {
				continue
			}

			_, mentioned := used[h.name]
			if !h.always && !mentioned && !mentionedBy(h.name, wanted) {
				continue
			}

			wanted[h.name] = true
			changed = true
		}
	}

	var sb strings.Builder

	for _, h := range helpers {
		if wanted[h.name] {
			sb.WriteString(h.source)
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

// mentionedBy reports whether any already-wanted helper's source names this one.
func mentionedBy(name string, wanted map[string]bool) bool {
	for _, h := range helpers {
		if wanted[h.name] && h.name != name && strings.Contains(h.source, name) {
			return true
		}
	}

	return false
}
