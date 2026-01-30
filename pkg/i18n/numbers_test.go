package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatNumber(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"zero", 0, "0"},
		{"small", 123, "123"},
		{"thousands", 1234, "1,234"},
		{"millions", 1234567, "1,234,567"},
		{"negative", -1234567, "-1,234,567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDecimal(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"integer", 1234.0, "1,234"},
		{"one decimal", 1234.5, "1,234.5"},
		{"two decimals", 1234.56, "1,234.56"},
		{"trailing zeros", 1234.50, "1,234.5"},
		{"small", 0.5, "0.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDecimal(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatPercent(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"whole", 0.85, "85%"},
		{"decimal", 0.333, "33.3%"},
		{"over 100", 1.5, "150%"},
		{"zero", 0.0, "0%"},
		{"one", 1.0, "100%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPercent(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatBytes(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"bytes", 500, "500 B"},
		{"KB", 1536, "1.5 KB"},
		{"MB", 1572864, "1.5 MB"},
		{"GB", 1610612736, "1.5 GB"},
		{"exact KB", 1024, "1 KB"},
		{"exact MB", 1048576, "1 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatOrdinal(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"1st", 1, "1st"},
		{"2nd", 2, "2nd"},
		{"3rd", 3, "3rd"},
		{"4th", 4, "4th"},
		{"11th", 11, "11th"},
		{"12th", 12, "12th"},
		{"13th", 13, "13th"},
		{"21st", 21, "21st"},
		{"22nd", 22, "22nd"},
		{"23rd", 23, "23rd"},
		{"100th", 100, "100th"},
		{"101st", 101, "101st"},
		{"111th", 111, "111th"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatOrdinal(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestI18nNumberNamespace(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	t.Run("i18n.numeric.number", func(t *testing.T) {
		result := svc.T("i18n.numeric.number", 1234567)
		assert.Equal(t, "1,234,567", result)
	})

	t.Run("i18n.numeric.decimal", func(t *testing.T) {
		result := svc.T("i18n.numeric.decimal", 1234.56)
		assert.Equal(t, "1,234.56", result)
	})

	t.Run("i18n.numeric.percent", func(t *testing.T) {
		result := svc.T("i18n.numeric.percent", 0.85)
		assert.Equal(t, "85%", result)
	})

	t.Run("i18n.numeric.bytes", func(t *testing.T) {
		result := svc.T("i18n.numeric.bytes", 1572864)
		assert.Equal(t, "1.5 MB", result)
	})

	t.Run("i18n.numeric.ordinal", func(t *testing.T) {
		result := svc.T("i18n.numeric.ordinal", 3)
		assert.Equal(t, "3rd", result)
	})
}
