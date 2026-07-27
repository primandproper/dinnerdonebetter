package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/primandproper/platform-go/v7/cryptography/encryption"
	"github.com/primandproper/platform-go/v7/cryptography/encryption/config"
	loggingnoop "github.com/primandproper/platform-go/v7/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v7/observability/tracing/noop"

	"github.com/spf13/cobra"
)

const secretKeyLengthBytes = 32

func main() {
	var secret, provider string

	root := &cobra.Command{
		Use:   "encryptor",
		Short: "Encrypt or decrypt payloads using AES or Salsa20",
	}
	root.PersistentFlags().StringVar(&secret, "secret", "", "hex-encoded 32-byte secret key (64 hex chars, required)")
	root.PersistentFlags().StringVar(&provider, "provider", config.ProviderSalsa20, "Encryption provider: aes or salsa20")

	if err := root.MarkPersistentFlagRequired("secret"); err != nil {
		log.Fatal(err)
	}

	root.AddCommand(encryptCmd(&secret, &provider))
	root.AddCommand(decryptCmd(&secret, &provider))

	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func encryptCmd(secret, provider *string) *cobra.Command {
	return &cobra.Command{
		Use:   "encrypt [payload]",
		Short: "Encrypt a payload",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runEncrypt(*secret, *provider, args[0])
		},
	}
}

func decryptCmd(secret, provider *string) *cobra.Command {
	return &cobra.Command{
		Use:   "decrypt [payload]",
		Short: "Decrypt a payload",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runDecrypt(*secret, *provider, args[0])
		},
	}
}

func runEncrypt(secret, provider, payload string) error {
	encDec, err := newEncryptorDecryptor(secret, provider)
	if err != nil {
		return err
	}

	ctx := context.Background()
	result, err := encDec.Encrypt(ctx, payload)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	fmt.Println(result)
	return nil
}

func runDecrypt(secret, provider, payload string) error {
	encDec, err := newEncryptorDecryptor(secret, provider)
	if err != nil {
		return err
	}

	ctx := context.Background()
	result, err := encDec.Decrypt(ctx, payload)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	fmt.Println(result)
	return nil
}

func newEncryptorDecryptor(key, provider string) (encryption.EncryptorDecryptor, error) {
	if len(key) != secretKeyLengthBytes {
		return nil, fmt.Errorf("secret must decode to 32 bytes (64 hex chars), got %d bytes", len(key))
	}

	encDec, err := config.NewEncryptorDecryptor(
		&config.Config{Provider: provider},
		tracingnoop.NewTracerProvider(),
		loggingnoop.NewLogger(),
		[]byte(key),
	)
	if err != nil {
		return nil, fmt.Errorf("create encryptor: %w", err)
	}

	return encDec, nil
}
