package config

import "os"

const DefaultDataDir = "data"

func DataDir() string {
	return dataDirFromEnv(os.Getenv)
}

func dataDirFromEnv(getenv func(string) string) string {
	if dataDir := getenv("DATA_DIR"); dataDir != "" {
		return dataDir
	}
	if dataDir := getenv("RAILWAY_VOLUME_MOUNT_PATH"); dataDir != "" {
		return dataDir
	}
	return DefaultDataDir
}
