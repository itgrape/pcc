package pccclient

import (
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/goccy/go-yaml"
)

func (c *Client) Load(target any) error {
	return c.LoadContext(context.Background(), target)
}

func (c *Client) LoadContext(ctx context.Context, target any) error {
	if c.namespace == "" {
		return ErrNamespaceRequired
	}
	if c.defaultConfig == "" {
		return ErrConfigRequired
	}
	return c.LoadConfig(ctx, c.namespace, c.defaultConfig, target)
}

func (c *Client) LoadConfig(
	ctx context.Context,
	namespace string,
	configName string,
	target any,
) error {
	if err := validateTarget(target); err != nil {
		return err
	}
	config, err := c.GetConfig(ctx, namespace, configName)
	if err != nil {
		return err
	}
	return config.Decode(target)
}

func (c *Client) LoadField(path string, target any) error {
	return c.LoadFieldContext(context.Background(), path, target)
}

func (c *Client) LoadFieldContext(ctx context.Context, path string, target any) error {
	if c.namespace == "" {
		return ErrNamespaceRequired
	}
	if c.defaultConfig == "" {
		return ErrConfigRequired
	}
	return c.LoadConfigField(ctx, c.namespace, c.defaultConfig, path, target)
}

func (c *Client) LoadConfigField(
	ctx context.Context,
	namespace string,
	configName string,
	path string,
	target any,
) error {
	if err := validateTarget(target); err != nil {
		return err
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("%w: empty path", ErrFieldNotFound)
	}

	config, err := c.GetConfig(ctx, namespace, configName)
	if err != nil {
		return err
	}
	return config.DecodeField(path, target)
}

func (config *Config) Decode(target any) error {
	if err := validateTarget(target); err != nil {
		return err
	}
	if err := decodeConfig(config, target); err != nil {
		return &DecodeError{ConfigName: config.Name, Format: config.Format, Err: err}
	}
	return nil
}

func (config *Config) DecodeField(path string, target any) error {
	if err := validateTarget(target); err != nil {
		return err
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("%w: empty path", ErrFieldNotFound)
	}
	document, err := decodeDocument(config)
	if err != nil {
		return &DecodeError{ConfigName: config.Name, Format: config.Format, Err: err}
	}
	value, err := lookupPath(document, path)
	if err != nil {
		return err
	}
	if err := assignValue(target, value); err != nil {
		return &DecodeError{ConfigName: config.Name, Format: config.Format, Err: err}
	}
	return nil
}

func (c *Client) String(path string) (string, error) {
	var value string
	err := c.LoadField(path, &value)
	return value, err
}

func (c *Client) Int(path string) (int, error) {
	var value int
	err := c.LoadField(path, &value)
	return value, err
}

func (c *Client) Bool(path string) (bool, error) {
	var value bool
	err := c.LoadField(path, &value)
	return value, err
}

func validateTarget(target any) error {
	if target == nil {
		return ErrTargetRequired
	}
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return ErrTargetRequired
	}
	return nil
}

func decodeConfig(config *Config, target any) error {
	if config.Format == FormatText {
		return assignValue(target, config.Content)
	}
	document, err := decodeDocument(config)
	if err != nil {
		return err
	}
	return assignValue(target, document)
}

func decodeDocument(config *Config) (any, error) {
	switch config.Format {
	case FormatJSON:
		decoder := json.NewDecoder(strings.NewReader(config.Content))
		decoder.UseNumber()
		var document any
		if err := decoder.Decode(&document); err != nil {
			return nil, err
		}
		return document, nil
	case FormatYAML:
		var document any
		if err := yaml.Unmarshal([]byte(config.Content), &document); err != nil {
			return nil, err
		}
		return document, nil
	case FormatProperties:
		return parseProperties(config.Content)
	case FormatText:
		return config.Content, nil
	default:
		return nil, fmt.Errorf("%w: %d", ErrUnsupportedFormat, config.Format)
	}
}

func parseProperties(content string) (map[string]any, error) {
	document := make(map[string]any)
	for lineNumber, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		separator := strings.IndexAny(line, "=:")
		if separator < 1 {
			return nil, fmt.Errorf("invalid properties line %d", lineNumber+1)
		}
		key := strings.TrimSpace(line[:separator])
		value := parseScalar(strings.TrimSpace(line[separator+1:]))
		setNested(document, strings.Split(key, "."), value)
	}
	return document, nil
}

func parseScalar(value string) any {
	if unquoted, err := strconv.Unquote(value); err == nil {
		return unquoted
	}
	if boolean, err := strconv.ParseBool(value); err == nil {
		return boolean
	}
	if integer, err := strconv.ParseInt(value, 10, 64); err == nil {
		return integer
	}
	if number, err := strconv.ParseFloat(value, 64); err == nil {
		return number
	}
	return value
}

func setNested(document map[string]any, path []string, value any) {
	current := document
	for _, part := range path[:len(path)-1] {
		next, ok := current[part].(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[part] = next
		}
		current = next
	}
	current[path[len(path)-1]] = value
}

func lookupPath(document any, path string) (any, error) {
	current := document
	for _, segment := range strings.Split(path, ".") {
		name, indexes, err := parseSegment(segment)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrFieldNotFound, path)
		}
		if name != "" {
			object, ok := current.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%w: %s", ErrFieldNotFound, path)
			}
			current, ok = object[name]
			if !ok {
				return nil, fmt.Errorf("%w: %s", ErrFieldNotFound, path)
			}
		}
		for _, index := range indexes {
			items, ok := current.([]any)
			if !ok || index < 0 || index >= len(items) {
				return nil, fmt.Errorf("%w: %s", ErrFieldNotFound, path)
			}
			current = items[index]
		}
	}
	return current, nil
}

func parseSegment(segment string) (string, []int, error) {
	name := segment
	var indexes []int
	if bracket := strings.IndexByte(segment, '['); bracket >= 0 {
		name = segment[:bracket]
		rest := segment[bracket:]
		for rest != "" {
			if rest[0] != '[' {
				return "", nil, ErrFieldNotFound
			}
			end := strings.IndexByte(rest, ']')
			if end < 2 {
				return "", nil, ErrFieldNotFound
			}
			index, err := strconv.Atoi(rest[1:end])
			if err != nil {
				return "", nil, err
			}
			indexes = append(indexes, index)
			rest = rest[end+1:]
		}
	}
	return name, indexes, nil
}

func assignValue(target any, value any) error {
	if text, ok := target.(encoding.TextUnmarshaler); ok {
		return text.UnmarshalText([]byte(fmt.Sprint(value)))
	}

	destination := reflect.ValueOf(target).Elem()
	if destination.Type() == reflect.TypeFor[time.Duration]() {
		duration, err := time.ParseDuration(fmt.Sprint(value))
		if err != nil {
			return err
		}
		destination.SetInt(int64(duration))
		return nil
	}

	switch destination.Kind() {
	case reflect.String:
		stringValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot assign %T to string", value)
		}
		destination.SetString(stringValue)
		return nil
	case reflect.Bool:
		boolean, err := strconv.ParseBool(fmt.Sprint(value))
		if err != nil {
			return err
		}
		destination.SetBool(boolean)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		integer, err := strconv.ParseInt(fmt.Sprint(value), 10, destination.Type().Bits())
		if err != nil {
			return err
		}
		destination.SetInt(integer)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		integer, err := strconv.ParseUint(fmt.Sprint(value), 10, destination.Type().Bits())
		if err != nil {
			return err
		}
		destination.SetUint(integer)
		return nil
	case reflect.Float32, reflect.Float64:
		number, err := strconv.ParseFloat(fmt.Sprint(value), destination.Type().Bits())
		if err != nil {
			return err
		}
		destination.SetFloat(number)
		return nil
	case reflect.Slice:
		if destination.Type().Elem().Kind() == reflect.Uint8 {
			text, ok := value.(string)
			if !ok {
				return fmt.Errorf("cannot assign %T to bytes", value)
			}
			destination.SetBytes([]byte(text))
			return nil
		}
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           target,
		TagName:          "json",
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.TextUnmarshallerHookFunc(),
		),
	})
	if err != nil {
		return err
	}
	return decoder.Decode(value)
}
