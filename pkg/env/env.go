package env

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/0xveya/gns3util/pkg/utils/nwutils"
)

type EnvType string

const (
	EnvTypeString EnvType = "string"
	EnvTypeInt    EnvType = "int"
	EnvTypePort   EnvType = "port"
	EnvTypeURL    EnvType = "url"
	EnvTypeListen EnvType = "listen"
	EnvTypeSecret EnvType = "secret"
	EnvTypeBool   EnvType = "bool"
)

type FieldConfig struct {
	EnvVar   string
	Type     EnvType
	Required bool
	Default  string
	Validate func(string) bool
}

func LoadConfig(cfg any) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("config must be a non-nil pointer")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}

		cfg := parseFieldTag(envTag, field.Tag)
		if cfg == nil {
			continue
		}

		envVal := os.Getenv(cfg.EnvVar)
		if envVal == "" {
			if cfg.Default != "" {
				envVal = cfg.Default
			} else if cfg.Required {
				return fmt.Errorf("required env var %s not set", cfg.EnvVar)
			}
		}

		if envVal == "" {
			continue
		}

		if cfg.Validate != nil && !cfg.Validate(envVal) {
			if cfg.Default != "" {
				envVal = cfg.Default
			} else {
				return fmt.Errorf("invalid value for %s: %s", cfg.EnvVar, envVal)
			}
		}

		if err := setFieldValue(fieldVal, envVal, cfg); err != nil {
			return fmt.Errorf("failed to set %s: %w", field.Name, err)
		}
	}

	return nil
}

func parseFieldTag(envVar string, tag reflect.StructTag) *FieldConfig {
	typ := tag.Get("type")
	requiredStr := tag.Get("required")
	defaultVal := tag.Get("default")

	return &FieldConfig{
		EnvVar:   envVar,
		Type:     EnvType(typ),
		Required: requiredStr == "true",
		Default:  defaultVal,
		Validate: getValidator(EnvType(typ)),
	}
}

func getValidator(typ EnvType) func(string) bool {
	switch typ {
	case EnvTypePort:
		return func(s string) bool {
			_, ok := nwutils.ParsePort(s)
			return ok
		}
	case EnvTypeListen:
		return nwutils.IsValidListenAddr
	case EnvTypeURL:
		return func(s string) bool {
			_, ok := nwutils.ParseURL(s)
			return ok
		}
	case EnvTypeSecret:
		return func(s string) bool {
			return len(s) >= 8
		}
	case EnvTypeBool:
		return isValidBool
	default:
		return func(s string) bool {
			return s != ""
		}
	}
}

func setFieldValue(fieldVal reflect.Value, envVal string, cfg *FieldConfig) error {
	switch cfg.Type {
	case EnvTypeString, EnvTypeSecret, EnvTypeURL, EnvTypeListen:
		if fieldVal.Kind() != reflect.String {
			return fmt.Errorf("field type mismatch: expected string")
		}
		fieldVal.SetString(envVal)

	case EnvTypeInt, EnvTypePort:
		if fieldVal.Kind() != reflect.Int {
			return fmt.Errorf("field type mismatch: expected int")
		}
		port, ok := nwutils.ParsePort(envVal)
		if !ok {
			return fmt.Errorf("invalid port value %s: ", envVal)
		}
		fieldVal.SetInt(int64(port))
	case EnvTypeBool:
		if fieldVal.Kind() != reflect.Bool {
			return fmt.Errorf("field type mismatch: expected bool")
		}
		b, err := parseBool(envVal)
		if err != nil {
			return err
		}
		fieldVal.SetBool(b)
	default:
		if fieldVal.Kind() == reflect.String {
			fieldVal.SetString(envVal)
		}
	}

	return nil
}

func isValidBool(s string) bool {
	_, err := parseBool(s)
	return err == nil
}

func parseBool(s string) (bool, error) {
	s = strings.ToLower(s)
	switch s {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool value: %s", s)
	}
}
