package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed skill.md
var skillContent []byte

var claudeSkillsCmd = &cobra.Command{
	Use:   "claude-skills",
	Short: "Install Claude Code skills for coop-cli",
	Long:  "Installs a Claude Code skill that teaches Claude how to use coop-cli commands.",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to find home directory: %w", err)
		}

		skillDir := filepath.Join(home, ".claude", "skills", "coop-cli")
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return fmt.Errorf("failed to create skill directory: %w", err)
		}

		skillPath := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(skillPath, skillContent, 0644); err != nil {
			return fmt.Errorf("failed to write skill file: %w", err)
		}

		fmt.Printf("Installed Claude Code skill to %s\n", skillPath)
		fmt.Println("Claude Code can now help you use coop-cli. Try asking: \"Search for milk on Coop\"")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(claudeSkillsCmd)
}
