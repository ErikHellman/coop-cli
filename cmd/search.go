package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/hellman/coop-cli/internal/api"
	"github.com/hellman/coop-cli/internal/auth"
	"github.com/hellman/coop-cli/internal/models"
	"github.com/spf13/cobra"
)

var searchLimit int

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for products",
	Long:  "Search for products on Coop.se and display their details.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		e, p, err := getCredentials()
		if err != nil {
			return err
		}

		session, err := auth.Login(e, p)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		client := api.NewClient(session, getStoreID())

		result, err := client.SearchProducts(query, searchLimit)
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
		category := models.CategoryPath(p.NavCategories)
		comparePrice := fmt.Sprintf("%.2f %s", p.ComparativePriceData.B2CPrice, p.ComparativePriceUnit.Text)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.2f kr\t%s\t%s\n",
			p.ID,
			p.Name,
			p.ManufacturerName,
			p.PackageSizeInfo,
			p.SalesPriceData.B2CPrice,
			comparePrice,
			category,
		)
	}
	w.Flush()
}
