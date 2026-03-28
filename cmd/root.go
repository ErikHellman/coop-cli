package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	email    string
	password string
	storeID  string
)

var rootCmd = &cobra.Command{
	Use:   "coop-cli",
	Short: "CLI tool for Coop online grocery shopping",
	Long:  "A command-line tool for searching products and managing your shopping cart on Coop.se.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&email, "email", "e", "", "Coop account email (or set COOP_EMAIL)")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Coop account password (or set COOP_PASSWORD)")
	rootCmd.PersistentFlags().StringVarP(&storeID, "store", "s", "251300", "Coop store ID (default: Stora Coop Boländerna)")
}

func getCredentials() (string, string, error) {
	e := email
	if e == "" {
		e = os.Getenv("COOP_EMAIL")
	}
	p := password
	if p == "" {
		p = os.Getenv("COOP_PASSWORD")
	}
	if e == "" || p == "" {
		return "", "", fmt.Errorf("email and password required: use --email/--password flags or COOP_EMAIL/COOP_PASSWORD environment variables")
	}
	return e, p, nil
}

func getStoreID() string {
	s := storeID
	if s == "" {
		s = os.Getenv("COOP_STORE")
	}
	if s == "" {
		s = "251300"
	}
	return s
}
