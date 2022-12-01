package mapping

import (
	"io"

	"github.com/zeromicro/go-zero/core/internal/encoding"
)

// UnmarshalYamlBytes unmarshals content into v.
func UnmarshalYamlBytes(content []byte, v interface{}, opts ...UnmarshalOption) error {
	b, err := encoding.YamlToJson(content)
	if err != nil {
		return err
	}

	return UnmarshalJsonBytes(b, v, opts...)
}

// UnmarshalYamlReader unmarshals content from reader into v.
func UnmarshalYamlReader(reader io.Reader, v interface{}, opts ...UnmarshalOption) error {
	b, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	return UnmarshalYamlBytes(b, v, opts...)
}
