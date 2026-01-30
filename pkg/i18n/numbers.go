// Package i18n provides internationalization for the CLI.
package i18n

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// NumberFormat defines locale-specific number formatting rules.
type NumberFormat struct {
	ThousandsSep string // "," for en, "." for de
	DecimalSep   string // "." for en, "," for de
	PercentFmt   string // "%s%%" for en, "%s %%" for de (space before %)
}

// Default number formats by language.
var numberFormats = map[string]NumberFormat{
	"en": {ThousandsSep: ",", DecimalSep: ".", PercentFmt: "%s%%"},
	"de": {ThousandsSep: ".", DecimalSep: ",", PercentFmt: "%s %%"},
	"fr": {ThousandsSep: " ", DecimalSep: ",", PercentFmt: "%s %%"},
	"es": {ThousandsSep: ".", DecimalSep: ",", PercentFmt: "%s%%"},
	"zh": {ThousandsSep: ",", DecimalSep: ".", PercentFmt: "%s%%"},
}

// getNumberFormat returns the number format for the current language.
func getNumberFormat() NumberFormat {
	lang := currentLangForGrammar()
	// Extract base language (en-GB → en)
	if idx := strings.IndexAny(lang, "-_"); idx > 0 {
		lang = lang[:idx]
	}
	if fmt, ok := numberFormats[lang]; ok {
		return fmt
	}
	return numberFormats["en"] // fallback
}

// FormatNumber formats an integer with locale-specific thousands separators.
//
//	FormatNumber(1234567) // "1,234,567" (en) or "1.234.567" (de)
func FormatNumber(n int64) string {
	nf := getNumberFormat()
	return formatIntWithSep(n, nf.ThousandsSep)
}

// FormatDecimal formats a float with locale-specific separators.
// Uses up to 2 decimal places, trimming trailing zeros.
//
//	FormatDecimal(1234.5)  // "1,234.5" (en) or "1.234,5" (de)
//	FormatDecimal(1234.00) // "1,234" (en) or "1.234" (de)
func FormatDecimal(f float64) string {
	return FormatDecimalN(f, 2)
}

// FormatDecimalN formats a float with N decimal places.
//
//	FormatDecimalN(1234.5678, 3) // "1,234.568" (en)
func FormatDecimalN(f float64, decimals int) string {
	nf := getNumberFormat()

	// Split into integer and fractional parts
	intPart := int64(f)
	fracPart := math.Abs(f - float64(intPart))

	// Format integer part with thousands separator
	intStr := formatIntWithSep(intPart, nf.ThousandsSep)

	// Format fractional part
	if decimals <= 0 || fracPart == 0 {
		return intStr
	}

	// Round and format fractional part
	multiplier := math.Pow(10, float64(decimals))
	fracInt := int64(math.Round(fracPart * multiplier))

	if fracInt == 0 {
		return intStr
	}

	// Format with leading zeros, then trim trailing zeros
	fracStr := fmt.Sprintf("%0*d", decimals, fracInt)
	fracStr = strings.TrimRight(fracStr, "0")

	return intStr + nf.DecimalSep + fracStr
}

// FormatPercent formats a decimal as a percentage.
//
//	FormatPercent(0.85)   // "85%" (en) or "85 %" (de)
//	FormatPercent(0.333)  // "33.3%" (en)
//	FormatPercent(1.5)    // "150%" (en)
func FormatPercent(f float64) string {
	nf := getNumberFormat()
	pct := f * 100

	// Format the number part
	var numStr string
	if pct == float64(int64(pct)) {
		numStr = strconv.FormatInt(int64(pct), 10)
	} else {
		numStr = FormatDecimalN(pct, 1)
	}

	return fmt.Sprintf(nf.PercentFmt, numStr)
}

// FormatBytes formats bytes as human-readable size.
//
//	FormatBytes(1536)      // "1.5 KB"
//	FormatBytes(1536000)   // "1.5 MB"
//	FormatBytes(1536000000) // "1.4 GB"
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	nf := getNumberFormat()

	var value float64
	var unit string

	switch {
	case bytes >= TB:
		value = float64(bytes) / TB
		unit = "TB"
	case bytes >= GB:
		value = float64(bytes) / GB
		unit = "GB"
	case bytes >= MB:
		value = float64(bytes) / MB
		unit = "MB"
	case bytes >= KB:
		value = float64(bytes) / KB
		unit = "KB"
	default:
		return fmt.Sprintf("%d B", bytes)
	}

	// Format with 1 decimal place, trim .0
	intPart := int64(value)
	fracPart := value - float64(intPart)

	if fracPart < 0.05 {
		return fmt.Sprintf("%d %s", intPart, unit)
	}

	fracDigit := int(math.Round(fracPart * 10))
	if fracDigit == 10 {
		return fmt.Sprintf("%d %s", intPart+1, unit)
	}

	return fmt.Sprintf("%d%s%d %s", intPart, nf.DecimalSep, fracDigit, unit)
}

// FormatOrdinal formats a number as an ordinal.
//
//	FormatOrdinal(1)  // "1st" (en) or "1." (de)
//	FormatOrdinal(2)  // "2nd" (en) or "2." (de)
//	FormatOrdinal(3)  // "3rd" (en) or "3." (de)
//	FormatOrdinal(11) // "11th" (en) or "11." (de)
func FormatOrdinal(n int) string {
	lang := currentLangForGrammar()
	// Extract base language
	if idx := strings.IndexAny(lang, "-_"); idx > 0 {
		lang = lang[:idx]
	}

	// Most languages just use number + period
	switch lang {
	case "en":
		return formatEnglishOrdinal(n)
	default:
		return fmt.Sprintf("%d.", n)
	}
}

// formatEnglishOrdinal returns English ordinal suffix.
func formatEnglishOrdinal(n int) string {
	abs := n
	if abs < 0 {
		abs = -abs
	}

	// Special cases for 11, 12, 13
	if abs%100 >= 11 && abs%100 <= 13 {
		return fmt.Sprintf("%dth", n)
	}

	switch abs % 10 {
	case 1:
		return fmt.Sprintf("%dst", n)
	case 2:
		return fmt.Sprintf("%dnd", n)
	case 3:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}

// formatIntWithSep formats an integer with thousands separator.
func formatIntWithSep(n int64, sep string) string {
	if sep == "" {
		return strconv.FormatInt(n, 10)
	}

	negative := n < 0
	if negative {
		n = -n
	}

	str := strconv.FormatInt(n, 10)
	if len(str) <= 3 {
		if negative {
			return "-" + str
		}
		return str
	}

	// Insert separators from right to left
	var result strings.Builder
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(sep)
		}
		result.WriteRune(c)
	}

	if negative {
		return "-" + result.String()
	}
	return result.String()
}
