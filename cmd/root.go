package cmd

import (
	"fmt"
	"os"

	"github.com/ErikHellman/coop-cli/internal/api"
	"github.com/ErikHellman/coop-cli/internal/auth"
	"github.com/spf13/cobra"
)

const repo = "ErikHellman/coop-cli"

var (
	Version string
	Commit  string
	Date    string
)

var (
	email    string
	password string
	storeID  string
)

func SetVersion(version, commit, date string) {
	Version = version
	Commit = commit
	Date = date
	rootCmd.Version = version
}

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
	rootCmd.PersistentFlags().StringVarP(&storeID, "store", "s", "", "Coop store ID (or set COOP_STORE, find with 'coop-cli stores')")
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

func getStoreID() (string, error) {
	s := storeID
	if s == "" {
		s = os.Getenv("COOP_STORE")
	}
	if s == "" {
		return "", fmt.Errorf("store ID required: use --store flag or COOP_STORE environment variable (find stores with 'coop-cli stores')")
	}
	return s, nil
}

func authenticatedClient() (*api.Client, error) {
	e, p, err := getCredentials()
	if err != nil {
		return nil, err
	}

	s, err := getStoreID()
	if err != nil {
		return nil, err
	}

	session, err := auth.Login(e, p)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return api.NewClient(session, s), nil
}
