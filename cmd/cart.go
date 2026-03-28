package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/ErikHellman/coop-cli/internal/models"
	"github.com/spf13/cobra"
)

var cartCmd = &cobra.Command{
	Use:   "cart",
	Short: "Manage your shopping cart",
	Long:  "Add, remove, list, or clear items in your Coop shopping cart.",
}

var cartListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all items in the cart",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := authenticatedClient()
		if err != nil {
			return err
		}

		cart, err := client.GetCart()
		if err != nil {
			return fmt.Errorf("failed to get cart: %w", err)
		}

		printCart(cart)
		return nil
	},
}

var cartAddCmd = &cobra.Command{
	Use:   "add <product-id> [quantity]",
	Short: "Add a product to the cart",
	Long:  "Add a product to the cart by its product ID. Default quantity is 1.",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		productID := args[0]
		quantity := 1
		if len(args) > 1 {
			var err error
			quantity, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid quantity: %s", args[1])
			}
		}

		client, err := authenticatedClient()
		if err != nil {
			return err
		}

		cart, err := client.AddToCart(productID, quantity)
		if err != nil {
			return fmt.Errorf("failed to add to cart: %w", err)
		}

		fmt.Printf("Added %s to cart.\n\n", productID)
		printCart(cart)
		return nil
	},
}

var cartRemoveCmd = &cobra.Command{
	Use:   "remove <product-id>",
	Short: "Remove a product from the cart",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		productID := args[0]

		client, err := authenticatedClient()
		if err != nil {
			return err
		}

		cart, err := client.RemoveFromCart(productID)
		if err != nil {
			return fmt.Errorf("failed to remove from cart: %w", err)
		}

		fmt.Printf("Removed %s from cart.\n\n", productID)
		printCart(cart)
		return nil
	},
}

var cartClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all items from the cart",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := authenticatedClient()
		if err != nil {
			return err
		}

		err = client.ClearCart()
		if err != nil {
			return fmt.Errorf("failed to clear cart: %w", err)
		}

		fmt.Println("Cart cleared.")
		return nil
	},
}

func init() {
	cartCmd.AddCommand(cartListCmd)
	cartCmd.AddCommand(cartAddCmd)
	cartCmd.AddCommand(cartRemoveCmd)
	cartCmd.AddCommand(cartClearCmd)
	rootCmd.AddCommand(cartCmd)
}

func printCart(cart *models.CartResponse) {
	if len(cart.Entries) == 0 {
		fmt.Printf("Cart is empty. (Store: %s)\n", cart.CoopStore.Name)
		return
	}

	fmt.Printf("Shopping cart at %s (%d items):\n\n", cart.CoopStore.Name, cart.TotalItems)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tPRODUCT\tQTY\tUNIT PRICE\tTOTAL")
	fmt.Fprintln(w, "--\t-------\t---\t----------\t-----")

	for _, entry := range cart.Entries {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
			entry.Product.Code,
			entry.Product.Name,
			entry.Quantity,
			entry.BasePrice.FormattedValue,
			entry.TotalPrice.FormattedValue,
		)
	}
	w.Flush()

	fmt.Printf("\nSubtotal: %s\n", cart.SubTotal.FormattedValue)
	fmt.Printf("Total:    %s\n", cart.TotalPrice.FormattedValue)
}
