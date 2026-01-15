package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

func setupRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/demo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, world!")
	})

	fs := http.FileServer(http.Dir("./ui/dist/i18n-element/browser"))
	mux.Handle("/", fs)
	return mux
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the i18n web server",
	Long: `Start the i18n web server.

The server provides a web interface for translating messages. The server
listens on port 8080 by default.`,
	Run: func(cmd *cobra.Command, args []string) {
		router := setupRouter()
		log.Println("Listening on :8080...")
		err := http.ListenAndServe(":8080", router)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
