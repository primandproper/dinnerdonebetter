package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	notifications "github.com/primandproper/platform-go/v7/notifications/mobile"
	"github.com/primandproper/platform-go/v7/notifications/mobile/apns"
	"github.com/primandproper/platform-go/v7/notifications/mobile/fcm"
	loggingnoop "github.com/primandproper/platform-go/v7/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v7/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v7/observability/tracing/noop"

	"github.com/spf13/pflag"
)

const (
	platformIOS     = "ios"
	platformAndroid = "android"
	defaultTitle    = "Test Push"
	defaultBody     = "This is a test notification from the push_test tool."
)

var (
	platform        = pflag.String("platform", "ios", "Target platform: ios or android (required)")
	deviceToken     = pflag.String("device-token", "", "Device token to send to (required)")
	title           = pflag.String("title", defaultTitle, "Notification title")
	body            = pflag.String("body", defaultBody, "Notification body")
	authKeyPath     = pflag.String("auth-key-path", "", "Path to APNs .p8 auth key file (required for ios)")
	keyID           = pflag.String("key-id", "", "APNs key ID (required for ios)")
	teamID          = pflag.String("team-id", "", "APNs team ID (required for ios)")
	bundleID        = pflag.String("bundle-id", "", "App bundle ID (required for ios)")
	production      = pflag.Bool("production", false, "Use APNs production environment (default: sandbox)")
	credentialsPath = pflag.String("credentials-path", "", "Path to Firebase service account JSON (required for android, or use ADC)")
)

func main() {
	pflag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
	log.Println("Push notification sent successfully.")
}

func run() error {
	p := strings.ToLower(strings.TrimSpace(*platform))
	token := strings.TrimSpace(*deviceToken)

	if p == "" {
		return fmt.Errorf("--platform is required (ios or android)")
	}
	if token == "" {
		return fmt.Errorf("--device-token is required")
	}

	ctx := context.Background()
	logger := loggingnoop.NewLogger()
	tracerProvider := tracingnoop.NewTracerProvider()
	metricsProvider := metricsnoop.NewMetricsProvider()

	var sender notifications.PushNotificationSender

	switch p {
	case platformIOS:
		if *authKeyPath == "" || *keyID == "" || *teamID == "" || *bundleID == "" {
			return fmt.Errorf("for ios, --auth-key-path, --key-id, --team-id, and --bundle-id are required")
		}
		apnsCfg := &apns.Config{
			AuthKeyPath: *authKeyPath,
			KeyID:       *keyID,
			TeamID:      *teamID,
			BundleID:    *bundleID,
			Production:  *production,
		}
		apnsSender, err := apns.NewSender(apnsCfg, tracerProvider, logger, metricsProvider)
		if err != nil {
			return fmt.Errorf("creating APNs sender: %w", err)
		}
		sender = notifications.NewMultiPlatformPushSender(apnsSender, nil, logger, tracerProvider)

	case platformAndroid:
		fcmCfg := &fcm.Config{CredentialsPath: *credentialsPath}
		fcmSender, err := fcm.NewSender(ctx, fcmCfg, tracerProvider, logger, metricsProvider)
		if err != nil {
			return fmt.Errorf("creating FCM sender: %w", err)
		}
		sender = notifications.NewMultiPlatformPushSender(nil, fcmSender, logger, tracerProvider)

	default:
		return fmt.Errorf("invalid platform %q: must be ios or android", p)
	}

	msg := notifications.PushMessage{Title: *title, Body: *body}
	return sender.SendPush(ctx, p, token, msg)
}
