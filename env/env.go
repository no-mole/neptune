package env

import (
	"fmt"
	"os"

	"github.com/go-ini/ini"
)

const (
	ModeDev  = "dev"
	ModeTest = "test"
	ModeGrey = "grey"
	ModeProd = "prod"
)

type Env struct {
	Mode  string `yaml:"mode"`
	Debug bool   `yaml:"debug"`
}

func IsAvailableEnvMode(name string) bool {
	return name == ModeDev || name == ModeTest || name == ModeGrey || name == ModeProd
}

var env Env

const (
	defaultMode  = ModeProd
	defaultDebug = false
	envMode      = "MODE"
	envDebug     = "DEBUG"
)

func init() {
	env = Env{Mode: defaultMode, Debug: defaultDebug}
}

func LoadEnv(rootDir string) Env {
	globalEnv := &env
	filePath := fmt.Sprintf("%s/%s", rootDir, ".env")
	cfg, err := ini.Load(filePath)
	if err == nil {
		if mode := cfg.Section("").Key("mode").String(); IsAvailableEnvMode(mode) {
			globalEnv.Mode = mode
		}
		if debug := cfg.Section("").Key("debug").String(); debug == "1" || debug == "true" {
			globalEnv.Debug = true
		}
	}
	mode := os.Getenv(envMode)
	if IsAvailableEnvMode(mode) {
		globalEnv.Mode = mode
	}
	debug := os.Getenv(envDebug)
	if debug == "1" || debug == "true" {
		globalEnv.Debug = true
	}
	return env
}

func GetEnvMode() string {
	return env.Mode
}

func GetEnvDebug() bool {
	return env.Debug
}
