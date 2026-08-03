/*
Package reportartifacts owns the object-storage half of a user data disclosure: where the
generated report lives, that it is ciphertext at rest, and that it can be destroyed.

Everything a person has ever done in this system ends up in one of these objects, so the
plaintext exists only in memory — on the way out of the aggregator, and on the way back in when
the subject asks for it. Nothing hands out a URL to the object itself, because the bytes behind
that URL are useless to whoever follows it: a signed URL would deliver base64 ciphertext and a
key the recipient does not have. Read-and-decrypt is the only delivery path, which is the shape
platform-go's dataprivacy package settled on as well (its Download refuses an encrypted artifact
with ErrArtifactEncrypted, and Open is the call that always works).
*/
package reportartifacts

//go:generate go tool github.com/matryer/moq -out store_mock.go -pkg reportartifacts -rm -fmt goimports . Store:StoreMock
