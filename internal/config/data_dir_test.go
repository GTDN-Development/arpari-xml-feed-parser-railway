package config

import "testing"

func TestDataDirFromEnvDefaultsToLocalData(t *testing.T) {
	got := dataDirFromEnv(func(string) string {
		return ""
	})

	if got != DefaultDataDir {
		t.Fatalf("expected default data dir %q, got %q", DefaultDataDir, got)
	}
}

func TestDataDirFromEnvUsesRailwayVolumeMountPath(t *testing.T) {
	got := dataDirFromEnv(func(key string) string {
		if key == "RAILWAY_VOLUME_MOUNT_PATH" {
			return "/data"
		}
		return ""
	})

	if got != "/data" {
		t.Fatalf("expected Railway volume mount path, got %q", got)
	}
}

func TestDataDirFromEnvPrefersExplicitDataDir(t *testing.T) {
	got := dataDirFromEnv(func(key string) string {
		switch key {
		case "DATA_DIR":
			return "/tmp/arpari-data"
		case "RAILWAY_VOLUME_MOUNT_PATH":
			return "/data"
		default:
			return ""
		}
	})

	if got != "/tmp/arpari-data" {
		t.Fatalf("expected explicit data dir, got %q", got)
	}
}
