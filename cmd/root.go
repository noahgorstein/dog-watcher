package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/noahgorstein/dog-watcher/tui"
	"github.com/noahgorstein/go-stardog/stardog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultConfigFilename = ".dog-watcher"
	envPrefix             = "DOG_WATCHER"
)

// getHTTPClient obtains the *http.Client with the appropriate auth
// that can then be used to create a stardog.Client
func getHTTPClient(token, username, password string) *http.Client {

	if token != "" {
		t := &stardog.BearerAuthTransport{
			BearerToken: token,
		}
		return t.Client()
	}

	t := &stardog.BasicAuthTransport{
		Username: username,
		Password: password,
	}
	return t.Client()
}

func NewRootCommand() *cobra.Command {
	token := ""
	username := ""
	password := ""
	endpoint := ""

	rootCmd := &cobra.Command{
		Version: "0.2.0",
		Use:     "dog-watcher",
		Short:   "a TUI to manage procceses in Stardog",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {

			httpClient := getHTTPClient(token, username, password)
			client, err := stardog.NewClient(endpoint, httpClient)
			if err != nil {
				fmt.Printf("unable to create client: %v\n", err.Error())
				os.Exit(1)
			}

			isAlive, _, err := client.ServerAdmin.IsAlive(context.Background())
			if err != nil || !*isAlive {
				fmt.Println("stardog server is not alive")
				if err != nil {
					fmt.Printf("err: %v\n", err.Error())

				}
				os.Exit(1)
			}

			bubble := tui.NewModel(client, endpoint)
			p := tea.NewProgram(bubble, tea.WithAltScreen())

			if err := p.Start(); err != nil {
				fmt.Println("Error running program:", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.Flags().StringVarP(&username, "username", "u", "admin", "username")
	rootCmd.Flags().StringVarP(&password, "password", "p", "admin", "password")
	rootCmd.Flags().StringVarP(&endpoint, "server", "s", "http://localhost:5820", "server")
	rootCmd.Flags().StringVarP(&token, "token", "t", "", "token")

	return rootCmd
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	v.SetConfigName(defaultConfigFilename)

	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	v.AddConfigPath(home)

	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	return nil
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix))
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
