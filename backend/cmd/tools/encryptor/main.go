package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/primandproper/platform-go/v13/cryptography/encryption"
	encryptioncfg "github.com/primandproper/platform-go/v13/cryptography/encryption/config"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"

	"github.com/spf13/cobra"
)

const (
	secretKeyLengthBytes = 32

	// defaultKeyID matches the CurrentKeyID the rendered configs ship with, so the tool opens
	// what the application wrote without being told the ID every time.
	defaultKeyID = "v1"
)

func main() {
	var secret, provider, keyID string

	root := &cobra.Command{
		Use:   "encryptor",
		Short: "Encrypt or decrypt payloads using the configured cipher",
	}
	root.PersistentFlags().StringVar(&secret, "secret", "", "hex-encoded 32-byte secret key (64 hex chars, required)")
	root.PersistentFlags().StringVar(&provider, "provider", encryptioncfg.ProviderAES, "Encryption provider")
	// The key ID is written into the ciphertext frame, and decryption looks the key up by it,
	// so reading a stored value back requires the ID it was written under.
	root.PersistentFlags().StringVar(&keyID, "key-id", defaultKeyID, "ID naming the key in the ring")

	if err := root.MarkPersistentFlagRequired("secret"); err != nil {
		log.Fatal(err)
	}

	root.AddCommand(encryptCmd(&secret, &provider, &keyID))
	root.AddCommand(decryptCmd(&secret, &provider, &keyID))

	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func encryptCmd(secret, provider, keyID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "encrypt [payload]",
		Short: "Encrypt a payload",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runEncrypt(*secret, *provider, *keyID, args[0])
		},
	}
}

func decryptCmd(secret, provider, keyID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "decrypt [payload]",
		Short: "Decrypt a payload",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runDecrypt(*secret, *provider, *keyID, args[0])
		},
	}
}

func runEncrypt(secret, provider, keyID, payload string) error {
	encDec, err := newEncryptorDecryptor(secret, provider, keyID)
	if err != nil {
		return err
	}

	ctx := context.Background()

	sealed, err := encDec.Encrypt(ctx, []byte(payload), nil)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	// base64 with the same alphabet the OAuth2 token columns use, so a value produced here
	// can be pasted into one of those columns and read back by the application.
	fmt.Println(base64.URLEncoding.EncodeToString(sealed))

	return nil
}

func runDecrypt(secret, provider, keyID, payload string) error {
	encDec, err := newEncryptorDecryptor(secret, provider, keyID)
	if err != nil {
		return err
	}

	raw, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	ctx := context.Background()

	opened, err := encDec.Decrypt(ctx, raw, nil)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	fmt.Println(string(opened))

	return nil
}

func newEncryptorDecryptor(key, provider, keyID string) (encryption.EncryptorDecryptor, error) {
	if len(key) != secretKeyLengthBytes {
		return nil, fmt.Errorf("secret must decode to 32 bytes (64 hex chars), got %d bytes", len(key))
	}

	encDec, err := encryptioncfg.NewKeyring(
		context.Background(),
		&encryptioncfg.Config{Provider: provider, CurrentKeyID: keyID},
		encryption.Keyset{encryption.KeyID(keyID): encryption.MasterKey(key)},
		encryptioncfg.WithLogger(loggingnoop.NewLogger()),
		encryptioncfg.WithTracerProvider(tracingnoop.NewTracerProvider()),
	)
	if err != nil {
		return nil, fmt.Errorf("create encryptor: %w", err)
	}

	return encDec, nil
}
