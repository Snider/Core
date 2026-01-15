module example.com/simple

go 1.24.3

require github.com/snider/i18n/i18n v0.0.0

require (
	github.com/nicksnyder/go-i18n/v2 v2.4.0 // indirect
	golang.org/x/text v0.16.0 // indirect
)

replace github.com/snider/i18n/i18n => ../../i18n
