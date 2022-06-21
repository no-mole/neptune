package dispatcher

import (
	"strconv"
	"strings"
)

type Config struct {
	Fields map[string]string `json:"fields" yaml:"fields"`
}

func (c *Config) GetString(key string, defaultValue string) string {
	if value, ok := c.Fields[key]; ok && value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) GetInt(key string, defaultValue int) int {
	value := c.GetString(key, "")
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func (c *Config) GetBool(key string, defaultValue bool) bool {
	value := c.GetString(key, "")
	if value == "" {
		return defaultValue
	}
	if value == "true" {
		return true
	}
	return false
}

func (c *Config) GetStringArray(key string, defaultValue []string) []string {
	value := c.GetString(key, "")
	if value == "" {
		return defaultValue
	}
	return strings.Split(value, ",")
}

func (c *Config) GetInt64Array(key string, defaultValue []int) []int {
	value := c.GetStringArray(key, []string{})
	if len(value) == 0 {
		return defaultValue
	}
	ret := make([]int, len(value))
	for k, v := range value {
		ret[k], _ = strconv.Atoi(v)
	}
	return ret
}
