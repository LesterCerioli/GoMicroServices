package conf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/zeromicro/go-zero/core/internal/types"
	"github.com/zeromicro/go-zero/core/jsonx"
	"github.com/zeromicro/go-zero/core/mapping"
	"gopkg.in/yaml.v2"
)

const distanceBetweenUpperAndLower = 32

var loaders = map[string]func([]byte, interface{}) error{
	".json": LoadFromJsonBytes,
	".toml": LoadFromTomlBytes,
	".yaml": LoadFromYamlBytes,
	".yml":  LoadFromYamlBytes,
}

// Load loads config into v from file, .json, .yaml and .yml are acceptable.
func Load(file string, v interface{}, opts ...Option) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	loader, ok := loaders[strings.ToLower(path.Ext(file))]
	if !ok {
		return fmt.Errorf("unrecognized file type: %s", file)
	}

	var opt options
	for _, o := range opts {
		o(&opt)
	}

	if opt.env {
		return loader([]byte(os.ExpandEnv(string(content))), v)
	}

	return loader(content, v)
}

// LoadConfig loads config into v from file, .json, .yaml and .yml are acceptable.
// Deprecated: use Load instead.
func LoadConfig(file string, v interface{}, opts ...Option) error {
	return Load(file, v, opts...)
}

// LoadFromJsonBytes loads config into v from content json bytes.
func LoadFromJsonBytes(content []byte, v interface{}) error {
	var m map[string]interface{}
	if err := jsonx.Unmarshal(content, &m); err != nil {
		return err
	}

	return mapping.UnmarshalJsonMap(toCamelCaseKeyMap(m), v, mapping.WithCanonicalKeyFunc(toCamelCase))
}

// LoadConfigFromJsonBytes loads config into v from content json bytes.
// Deprecated: use LoadFromJsonBytes instead.
func LoadConfigFromJsonBytes(content []byte, v interface{}) error {
	return LoadFromJsonBytes(content, v)
}

// LoadFromTomlBytes loads config into v from content toml bytes.
func LoadFromTomlBytes(content []byte, v interface{}) error {
	var val interface{}
	if err := toml.NewDecoder(bytes.NewReader(content)).Decode(&val); err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(val); err != nil {
		return err
	}

	return LoadFromJsonBytes(buf.Bytes(), v)
}

// LoadFromYamlBytes loads config into v from content yaml bytes.
func LoadFromYamlBytes(content []byte, v interface{}) error {
	var res interface{}
	if err := yaml.Unmarshal(content, &res); err != nil {
		return err
	}

	res = types.ToStringKeyMap(res)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(res); err != nil {
		return err
	}

	return LoadFromJsonBytes(buf.Bytes(), v)
}

// LoadConfigFromYamlBytes loads config into v from content yaml bytes.
// Deprecated: use LoadFromYamlBytes instead.
func LoadConfigFromYamlBytes(content []byte, v interface{}) error {
	return LoadFromYamlBytes(content, v)
}

// MustLoad loads config into v from path, exits on error.
func MustLoad(path string, v interface{}, opts ...Option) {
	if err := Load(path, v, opts...); err != nil {
		log.Fatalf("error: config file %s, %s", path, err.Error())
	}
}

func toCamelCase(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))
	var capNext bool
	boundary := true
	for _, v := range s {
		isCap := v >= 'A' && v <= 'Z'
		isLow := v >= 'a' && v <= 'z'
		if boundary && (isCap || isLow) {
			if capNext {
				if isLow {
					v -= distanceBetweenUpperAndLower
				}
			} else {
				if isCap {
					v += distanceBetweenUpperAndLower
				}
			}
			boundary = false
		}
		if isCap || isLow {
			buf.WriteRune(v)
			capNext = false
		} else if v == ' ' || v == '\t' {
			buf.WriteRune(v)
			capNext = false
			boundary = true
		} else if v == '_' {
			capNext = true
			boundary = true
		} else {
			buf.WriteRune(v)
			capNext = true
		}
	}

	return buf.String()
}

func toCamelCaseKeyMap(m map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range m {
		vv, ok := v.(map[string]interface{})
		if ok {
			res[toCamelCase(k)] = toCamelCaseKeyMap(vv)
		} else {
			res[toCamelCase(k)] = v
		}
	}

	return res
}
