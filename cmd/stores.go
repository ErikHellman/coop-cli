package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ErikHellman/coop-cli/internal/api"
	"github.com/spf13/cobra"
)

var storesCmd = &cobra.Command{
	Use:   "stores <query>",
	Short: "Search for Coop stores",
	Long:  "Search for Coop stores by name, city, or address. The store ID shown is the one to use with the --store flag.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stores, err := api.SearchStores(args[0])
		if err != nil {
			return fmt.Errorf("store search failed: %w", err)
		}

		if len(stores) == 0 {
			fmt.Println("No stores found.")
			return nil
		}

		fmt.Printf("Found %d stores:\n\n", len(stores))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "STORE ID\tNAME\tADDRESS\tPOSTAL CODE\tCITY\tPHONE\tOPEN TODAY")
		fmt.Fprintln(w, "--------\t----\t-------\t-----------\t----\t-----\t----------")

		for _, s := range stores {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				s.LedgerAccountNumber,
				s.Name,
				s.Address,
				s.PostalCode,
				s.City,
				s.Phone,
				s.OpeningHoursToday,
			)
		}
		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(storesCmd)
}
