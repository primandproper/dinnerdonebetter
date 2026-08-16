package integration

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"
)

// A virtual authenticator, because the alternative is not testing the ceremonies.
//
// Everything the passkey path does sits either side of a signature made by a security key: the
// ceremony is begun, a device signs the challenge, and the ceremony is finished by verifying
// that signature. Without a device, the only assertions available are that a malformed request
// is rejected — which leaves the half worth testing, that a challenge is stored durably and
// answerable exactly once, observable through nothing.
//
// So this is a real ES256 authenticator: a P-256 key, a COSE public key, an authenticator data
// structure, and a signature over the bytes the specification says to sign. It produces the two
// response payloads a browser would POST. What it does not do is attestation — it registers
// with the "none" format, which is what a passkey deployment asks for anyway.
const (
	// Authenticator data flags, from the specification. The two this authenticator does not
	// set — backup eligible and backup state — are deliberately absent in both ceremonies: a
	// credential registered with one answer and asserted with the other is refused for an
	// inconsistency that is real, and would look here like a flaky test.
	flagUserPresent            = 0x01
	flagUserVerified           = 0x04
	flagAttestedCredentialData = 0x40

	// The COSE labels for an ES256 key on P-256: kty=EC2(2), alg=ES256(-7), crv=P-256(1).
	coseKeyType   = 2
	coseAlgorithm = -7
	coseCurve     = 1

	// aaguidLength is fixed by the attested credential data layout: sixteen zero bytes here,
	// which is what an authenticator that declines to identify its model reports.
	aaguidLength = 16

	// coordinateLength is the width of one P-256 coordinate.
	coordinateLength = 32

	// The relying party this device signs for, and the origin it claims, taken from the
	// same constants the rendered testing config is generated from. Either one disagreeing
	// with the server's is a ceremony that fails verification, so they are not spelled
	// twice.
	passkeyTestRPID   = branding.LocalDevRPID
	passkeyTestOrigin = branding.LocalDevConsumerWebAppURL
)

// virtualAuthenticator is one passkey on one device.
type virtualAuthenticator struct {
	key          *ecdsa.PrivateKey
	credentialID []byte
	signCount    uint32
}

// newVirtualAuthenticator mints a device with one credential.
func newVirtualAuthenticator(t *testing.T) *virtualAuthenticator {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	credentialID := make([]byte, 32)
	_, err = rand.Read(credentialID)
	require.NoError(t, err)

	return &virtualAuthenticator{key: key, credentialID: credentialID, signCount: 1}
}

// register produces the attestation response a browser would POST to finish a registration
// ceremony for challenge.
func (a *virtualAuthenticator) register(t *testing.T, challenge string) []byte {
	t.Helper()

	clientData := a.clientData(t, "webauthn.create", challenge)
	authData := a.authenticatorData(t, flagUserPresent|flagUserVerified|flagAttestedCredentialData, a.attestedCredentialData(t))

	attestation, err := cbor.Marshal(map[string]any{
		"fmt":      "none",
		"attStmt":  map[string]any{},
		"authData": authData,
	})
	require.NoError(t, err)

	return marshalAuthenticatorResponse(t, a.credentialID, map[string]string{
		"clientDataJSON":    encodeAuthenticatorBytes(clientData),
		"attestationObject": encodeAuthenticatorBytes(attestation),
	})
}

// assert produces the assertion response a browser would POST to finish a login ceremony for
// challenge. userHandle is echoed back by the authenticator, and is what a discoverable login
// identifies the user by.
//
// The sign count advances on every call, which is what the server writes back and compares the
// next login against.
func (a *virtualAuthenticator) assert(t *testing.T, challenge string, userHandle []byte) []byte {
	t.Helper()

	a.signCount++

	clientData := a.clientData(t, "webauthn.get", challenge)
	authData := a.authenticatorData(t, flagUserPresent|flagUserVerified, nil)

	// The signature covers the authenticator data followed by the hash of the client data,
	// which is what ties one signature to one challenge from one origin.
	clientDataHash := sha256.Sum256(clientData)
	signed := sha256.Sum256(append(append([]byte{}, authData...), clientDataHash[:]...))

	signature, err := ecdsa.SignASN1(rand.Reader, a.key, signed[:])
	require.NoError(t, err)

	response := map[string]string{
		"clientDataJSON":    encodeAuthenticatorBytes(clientData),
		"authenticatorData": encodeAuthenticatorBytes(authData),
		"signature":         encodeAuthenticatorBytes(signature),
	}

	if len(userHandle) > 0 {
		response["userHandle"] = encodeAuthenticatorBytes(userHandle)
	}

	return marshalAuthenticatorResponse(t, a.credentialID, response)
}

// clientData renders the collected client data for one ceremony step.
func (a *virtualAuthenticator) clientData(t *testing.T, ceremony, challenge string) []byte {
	t.Helper()

	data, err := json.Marshal(map[string]any{
		"type":        ceremony,
		"challenge":   challenge,
		"origin":      passkeyTestOrigin,
		"crossOrigin": false,
	})
	require.NoError(t, err)

	return data
}

// authenticatorData renders the authenticator data structure: the relying party's hash, the
// flags, the counter, and — for a registration — the credential itself.
func (a *virtualAuthenticator) authenticatorData(t *testing.T, flags byte, attested []byte) []byte {
	t.Helper()

	rpIDHash := sha256.Sum256([]byte(passkeyTestRPID))

	data := make([]byte, 0, sha256.Size+1+4+len(attested))
	data = append(data, rpIDHash[:]...)
	data = append(data, flags)
	data = binary.BigEndian.AppendUint32(data, a.signCount)

	return append(data, attested...)
}

// attestedCredentialData renders the credential a registration announces: the authenticator's
// AAGUID, the credential ID, and the public key.
func (a *virtualAuthenticator) attestedCredentialData(t *testing.T) []byte {
	t.Helper()

	data := make([]byte, aaguidLength)
	data = binary.BigEndian.AppendUint16(data, uint16(len(a.credentialID)))

	data = append(data, a.credentialID...)

	return append(data, a.coseKey(t)...)
}

// coseKey renders the public key in the COSE encoding the specification requires,
// deterministically so that the same key renders the same bytes.
func (a *virtualAuthenticator) coseKey(t *testing.T) []byte {
	t.Helper()

	encoder, err := cbor.CanonicalEncOptions().EncMode()
	require.NoError(t, err)

	// The uncompressed point is a leading 0x04 followed by the two coordinates, which is
	// where the COSE encoding's x and y come from.
	point, err := a.key.PublicKey.Bytes()
	require.NoError(t, err)
	require.Len(t, point, 1+2*coordinateLength)

	key, err := encoder.Marshal(map[int]any{
		1:  coseKeyType,
		3:  coseAlgorithm,
		-1: coseCurve,
		-2: point[1 : 1+coordinateLength],
		-3: point[1+coordinateLength:],
	})
	require.NoError(t, err)

	return key
}

// marshalAuthenticatorResponse wraps one ceremony's response in the credential envelope the
// browser's WebAuthn API produces.
func marshalAuthenticatorResponse(t *testing.T, credentialID []byte, response map[string]string) []byte {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"id":       encodeAuthenticatorBytes(credentialID),
		"rawId":    encodeAuthenticatorBytes(credentialID),
		"type":     "public-key",
		"response": response,
	})
	require.NoError(t, err)

	return body
}

// encodeAuthenticatorBytes renders bytes the way every field of a WebAuthn response is rendered.
func encodeAuthenticatorBytes(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}
