package tck_test

import (
	"testing"

	arkjson "goark.dev/arkarta/json"
	"goark.dev/arkarta/json/sonic"
	"goark.dev/arkarta/json/tck"
)

func TestRunCodecWithStandardCodec(t *testing.T) {
	tck.RunCodec(t, tck.CodecFactory{
		New: func() arkjson.Codec {
			return arkjson.NewStandardCodec()
		},
		WithEscapeHTML: func(enabled bool) arkjson.Codec {
			return arkjson.NewStandardCodec(arkjson.WithEscapeHTML(enabled))
		},
		WithMaxBytes: func(maxBytes int64) arkjson.Codec {
			return arkjson.NewStandardCodec(arkjson.WithMaxBytes(maxBytes))
		},
		WithUnknownFieldGate: func(enabled bool) arkjson.Codec {
			return arkjson.NewStandardCodec(arkjson.WithDisallowUnknownFields(enabled))
		},
		WithUseNumber: func(enabled bool) arkjson.Codec {
			return arkjson.NewStandardCodec(arkjson.WithUseNumber(enabled))
		},
	})
}

func TestRunCodecWithSonicCodec(t *testing.T) {
	tck.RunCodec(t, tck.CodecFactory{
		New: func() arkjson.Codec {
			return sonic.NewCodec()
		},
		WithEscapeHTML: func(enabled bool) arkjson.Codec {
			return sonic.NewCodec(sonic.WithEscapeHTML(enabled))
		},
		WithMaxBytes: func(maxBytes int64) arkjson.Codec {
			return sonic.NewCodec(sonic.WithMaxBytes(maxBytes))
		},
		WithUnknownFieldGate: func(enabled bool) arkjson.Codec {
			return sonic.NewCodec(sonic.WithDisallowUnknownFields(enabled))
		},
		WithUseNumber: func(enabled bool) arkjson.Codec {
			return sonic.NewCodec(sonic.WithUseNumber(enabled))
		},
	})
}
