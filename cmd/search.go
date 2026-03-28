package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ErikHellman/coop-cli/internal/models"
	"github.com/spf13/cobra"
)

var searchLimit int

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for products",
	Long:  "Search for products on Coop.se and display their details.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := authenticatedClient()
		if err != nil {
			return err
		}

		result, err := client.SearchProducts(args[0], searchLimit)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if result.Results.Count == 0 {
			fmt.Println("No products found.")
			return nil
		}

		fmt.Printf("Found %d products (showing %d):\n\n", result.Results.Count, len(result.Results.Items))
		printProducts(result.Results.Items)
		return nil
	},
}

func init() {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 10, "Maximum number of results to return")
	rootCmd.AddCommand(searchCmd)
}

func printProducts(products []models.Product) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tMANUFACTURER\tSIZE\tPRICE\tCOMPARE\tCATEGORY")
	fmt.Fprintln(w, "--\t----\t------------\t----\t-----\t-------\t--------")

	for _, p := range products {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.2f kr\t%.2f %s\t%s\n",
			p.ID,
			p.Name,
			p.ManufacturerName,
			p.PackageSizeInfo,
			p.SalesPriceData.B2CPrice,
			p.ComparativePriceData.B2CPrice,
			p.ComparativePriceUnit.Text,
			models.CategoryPath(p.NavCategories),
		)
	}
	w.Flush()
}
