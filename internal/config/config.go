// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"gopkg.in/yaml.v3"
)

type MysqlConf struct {
	DataSource string
}

type CycleConf struct {
	TotalPoints int64
}

type Config struct {
	rest.RestConf
	Mysql MysqlConf
	Cycle CycleConf
}

func Load(file string, v *Config) error {
	ext := strings.ToLower(filepath.Ext(file))
	if ext != ".yaml" && ext != ".yml" {
		return conf.Load(file, v)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var raw map[string]any
	if err = yaml.Unmarshal(content, &raw); err != nil {
		return err
	}

	normalizeLegacyCycle(raw)

	encoded, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	return conf.LoadFromJsonBytes(encoded, v)
}

func MustLoad(file string, v *Config) {
	if err := Load(file, v); err != nil {
		log.Fatalf("error: config file %s, %s", file, err.Error())
	}
}

func normalizeLegacyCycle(raw map[string]any) {
	value, ok := raw["Cycle"]
	if !ok {
		value, ok = raw["cycle"]
	}
	if !ok {
		return
	}

	if _, ok = value.(map[string]any); ok {
		return
	}

	points, ok := cycleTotalPoints(value)
	if !ok {
		return
	}

	raw["Cycle"] = map[string]any{
		"TotalPoints": points,
	}
}

func cycleTotalPoints(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int8:
		return int64(typed), true
	case int16:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case int64:
		return typed, true
	case uint:
		return int64(typed), true
	case uint8:
		return int64(typed), true
	case uint16:
		return int64(typed), true
	case uint32:
		return int64(typed), true
	case uint64:
		return int64(typed), true
	case float32:
		return int64(typed), true
	case float64:
		return int64(typed), true
	default:
		return 0, false
	}
}
