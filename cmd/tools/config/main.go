//go:build spheretools
// +build spheretools

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-sphere/sphere-bun-layout/internal/config"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "config",
		Short: "Config Tools",
		Long:  `Config Tools is a set of tools for config operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	genCmd = &cobra.Command{
		Use:   "gen",
		Short: "Generate config file",
		Long:  `Generate a config file with default values.`,
	}
	testCmd = &cobra.Command{
		Use:   "test",
		Short: "Test config file format",
		Long:  `Test config file format is correct.`,
	}
)

func main() {
	Execute()
}

func init() {
	{
		flag := testCmd.Flags()
		conf := flag.String("config", "config.json", "config file path")
		testCmd.RunE = func(cmd *cobra.Command, args []string) error {
			con, err := config.NewConfig(*conf)
			if err != nil {
				return err
			}
			bytes, err := json.MarshalIndent(con, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(bytes))
			return nil
		}
	}
	{

		flag := genCmd.Flags()
		output := flag.String("output", "config_gen.json", "output file path")
		genCmd.RunE = func(*cobra.Command, []string) error {
			conf := config.NewEmptyConfig()
			raw, err := json.MarshalIndent(conf, "", "  ")
			if err != nil {
				return err
			}
			return os.WriteFile(*output, raw, 0644)
		}
	}
	rootCmd.AddCommand(genCmd, testCmd)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
