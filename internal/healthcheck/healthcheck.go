package healthcheck

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"
)

func Register(app *pocketbase.PocketBase) {
	var targetURL string

	cmd := &cobra.Command{
		Use:          "healthcheck",
		Short:        "Perform an HTTP health check on the running instance",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := &http.Client{Timeout: 3 * time.Second}
			resp, err := client.Get(targetURL)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unhealthy status code: %d", resp.StatusCode)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&targetURL, "url", "http://127.0.0.1:8090/api/health", "Health check endpoint URL")
	app.RootCmd.AddCommand(cmd)
}
