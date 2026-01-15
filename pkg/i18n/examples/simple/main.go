package main

import (
	"fmt"
	"log"

	"github.com/snider/i18n/i18n"
)

func main() {
	// Create a new i18n service.
	service, err := i18n.New()
	if err != nil {
		log.Fatalf("failed to create i18n service: %v", err)
	}

	// Set the language to French.
	err = service.SetLanguage("fr")
	if err != nil {
		log.Fatalf("failed to set language: %v", err)
	}

	// Translate a message.
	searchMessage := service.Translate("app.ui.search")
	fmt.Println(searchMessage)

	// Set the language to Spanish.
	err = service.SetLanguage("es")
	if err != nil {
		log.Fatalf("failed to set language: %v", err)
	}

	// Translate the same message again.
	searchMessage = service.Translate("app.ui.search")
	fmt.Println(searchMessage)

	// Translate with arguments.
	// Note: You would need to add "greeting": "Hello {{.Name}}" to your locale files to make this work properly.
	// Since we are using embedded locales in the library, we can't easily modify them here without rebuilding the lib.
	// However, we can demonstrate the API usage.
	greeting := service.Translate("greeting", map[string]string{"Name": "World"})
	fmt.Println(greeting)
}
