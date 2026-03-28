package cmd

import (
	"fmt"
	"os"

	"github.com/ErikHellman/coop-cli/internal/api"
	"github.com/ErikHellman/coop-cli/internal/auth"
	"github.com/spf13/cobra"
)

const defaultStoreID = "251300"

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
	rootCmd.PersistentFlags().StringVarP(&storeID, "store", "s", defaultStoreID, "Coop store ID (default: Stora Coop Boländerna)")
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
	if s := os.Getenv("COOP_STORE"); storeID == defaultStoreID && s != "" {
		return s
	}
	return storeID
}

func authenticatedClient() (*api.Client, error) {
	e, p, err := getCredentials()
	if err != nil {
		return nil, err
	}

	session, err := auth.Login(e, p)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return api.NewClient(session, getStoreID()), nil
}
