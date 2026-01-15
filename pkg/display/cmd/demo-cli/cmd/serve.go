package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the HTTP server",
	Long:  `Starts the HTTP server to serve the frontend and the API.`,
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/api/v1/demo", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, world!")
		})

		fs := http.FileServer(http.Dir("./ui/dist/display/browser"))
		http.Handle("/", fs)

		log.Println("Listening on :8080...")
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
