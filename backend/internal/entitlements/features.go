package entitlements

import (
	appmetering "github.com/primandproper/dinnerdonebetter/backend/internal/metering"

	platformentitlements "github.com/primandproper/platform-go/v13/entitlements"
)

const (
	// UploadedMediaBytesFeature gates the bytes an account may push through the media upload
	// endpoint, and is the quota half of the meter internal/metering counts them with.
	//
	// The key and the meter name are the same string today and are deliberately not the same
	// declaration. A meter is named by whoever set billing up and travels into provider-side
	// idempotency keys; a feature is named by whoever writes the gate. They agree until the
	// first time one of them is renamed.
	UploadedMediaBytesFeature = "uploaded_media_bytes"
)

// Features are the entitlement features this application gates on.
//
// They are Go rather than configuration, and the asymmetry with Plans is the point: a quota
// feature names a meter this application registered and a boolean one names a permission its
// handlers check, so neither means anything the code has not already been written for. A
// catalog that discovered its features from configuration would let a typo silently create a
// feature nobody gates on.
//
// Adding one here is not enough to make it do anything: something has to ask about it, and for
// a quota feature the meter it names has to be registered and recorded against.
func Features() []platformentitlements.Feature {
	return []platformentitlements.Feature{
		{
			Key:   UploadedMediaBytesFeature,
			Kind:  platformentitlements.KindQuota,
			Meter: appmetering.UploadedMediaBytesMeter,
		},
	}
}
