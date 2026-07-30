package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	platformconfig "github.com/primandproper/platform-go/v8/config"
)

const (
	// CeaseOperationEnvVarKey is the env var key used to indicate a function or job should just quit early.
	CeaseOperationEnvVarKey = "CEASE_OPERATION"
	// RunningInKubernetesEnvVarKey is the env var key we use to indicate we're running in Kubernetes.
	RunningInKubernetesEnvVarKey = "RUNNING_IN_KUBERNETES"
)

func ConditionallyCease() {
	if ShouldCeaseOperation() {
		slog.Info(fmt.Sprintf("%s is set to true, exiting", CeaseOperationEnvVarKey))
		os.Exit(0)
	}
}

// ShouldCeaseOperation returns whether a job should just quit without trying.
func ShouldCeaseOperation() bool {
	return strings.TrimSpace(strings.ToLower(os.Getenv(CeaseOperationEnvVarKey))) == "true"
}

// RunningInKubernetes returns whether the service is running in a Kubernetes cluster.
func RunningInKubernetes() bool {
	return os.Getenv(RunningInKubernetesEnvVarKey) != ""
}

// envVarOptions returns the platform config options shared by every loader in
// this package: the DINNER_DONE_BETTER_ prefix and a debug-logging OnSet hook.
func envVarOptions() []platformconfig.Option {
	return []platformconfig.Option{
		platformconfig.WithPrefix(EnvVarPrefix),
		platformconfig.WithOnSet(func(tag string, value any, isDefault bool) {
			slog.Debug("env var set",
				slog.String("tag", tag),
				slog.String("value", fmt.Sprintf("%+v", value)),
				slog.Bool("isDefault", isDefault),
			)
		}),
	}
}

func ApplyEnvironmentVariables(cfg any) error {
	return platformconfig.ApplyEnvironmentVariables(cfg, envVarOptions()...)
}
