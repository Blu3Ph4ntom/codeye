package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/codeye/codeye/internal/fonts"
	"github.com/spf13/cobra"
)

const fontName = "JetBrainsMonoNerdFont-Regular.ttf"
const markerName = ".nf_installed"

// markerPath returns the path to the marker file written after font install.
func markerPath() string {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(cfg, "codeye", markerName)
}

// nfInstalled reports whether codeye font install has been run.
func nfInstalled() bool {
	mp := markerPath()
	if mp == "" {
		return false
	}
	_, err := os.Stat(mp)
	return err == nil
}

// fontInstallDir returns the OS-appropriate fonts directory for the current user.
func fontInstallDir() (string, error) {
	switch runtime.GOOS {
	case "linux":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share", "fonts"), nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Fonts"), nil
	case "windows":
		local := os.Getenv("LOCALAPPDATA")
		if local == "" {
			return "", fmt.Errorf("%%LOCALAPPDATA%% not set")
		}
		return filepath.Join(local, "Microsoft", "Windows", "Fonts"), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func fontCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "font",
		Short: "Manage the bundled Nerd Font",
	}
	cmd.AddCommand(fontInstallCmd(), fontStatusCmd())
	return cmd
}

func fontInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install JetBrainsMono Nerd Font to your user fonts directory",
		Long: `Extracts the bundled JetBrainsMono Nerd Font (v3.3.0) from the binary
and installs it to your user fonts directory. No internet connection required.

After installation, set "JetBrainsMono Nerd Font" as your terminal font,
then run codeye with --nf for VS Code-style file icons.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := fontInstallDir()
			if err != nil {
				return fmt.Errorf("cannot determine font directory: %w", err)
			}

			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("cannot create font directory: %w", err)
			}

			dest := filepath.Join(dir, fontName)
			if err := os.WriteFile(dest, fonts.JetBrainsMonoRegular, 0o644); err != nil {
				return fmt.Errorf("cannot write font file: %w", err)
			}
			fmt.Printf("✓ installed %s\n  → %s\n\n", fontName, dest)

			// Refresh font cache on Linux
			if runtime.GOOS == "linux" {
				fmt.Println("  refreshing font cache (fc-cache)...")
				out, err := exec.Command("fc-cache", "-f", dir).CombinedOutput()
				if err != nil {
					// non-fatal: font still works, user can refresh manually
					fmt.Printf("  warning: fc-cache failed: %v\n  %s\n", err, out)
				} else {
					fmt.Println("  ✓ font cache refreshed")
				}
			}

			// Write marker so codeye can auto-detect on next run
			mp := markerPath()
			if mp != "" {
				_ = os.MkdirAll(filepath.Dir(mp), 0o755)
				_ = os.WriteFile(mp, []byte("1"), 0o644)
			}

			fmt.Println(`
Next step: set your terminal font to "JetBrainsMono Nerd Font"

  • GNOME Terminal → Edit → Preferences → Text → Custom font
  • Kitty:          font_family JetBrainsMono Nerd Font
  • Alacritty:      family: "JetBrainsMono Nerd Font"
  • WezTerm:        font = wezterm.font("JetBrainsMono Nerd Font")
  • iTerm2:         Preferences → Profiles → Text → Font
  • VS Code:        "editor.fontFamily": "JetBrainsMono Nerd Font"

Then run:  codeye --nf`)
			return nil
		},
	}
}

func fontStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show font installation status",
		Run: func(cmd *cobra.Command, args []string) {
			dir, _ := fontInstallDir()
			dest := ""
			installed := false
			if dir != "" {
				dest = filepath.Join(dir, fontName)
				if _, err := os.Stat(dest); err == nil {
					installed = true
				}
			}

			if installed {
				fmt.Printf("✓ font installed: %s\n", dest)
				if nfInstalled() {
					fmt.Println("✓ marker set — --nf auto-activated when font is detected")
				}
				fmt.Println("\nTest Nerd Font icons:  codeye --nf")
			} else {
				fmt.Println("✗ font not installed")
				fmt.Println("\nInstall with:  codeye font install")
			}
		},
	}
}
