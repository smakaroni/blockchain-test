package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/smakaroni/maaad-blockchain-household/wallet"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

func walletCmd() *cobra.Command {
	var walletCmd = &cobra.Command{
		Use:   "wallet",
		Short: "Manages blockchain accounts and keys",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	walletCmd.AddCommand(walletNewAccount())
	walletCmd.AddCommand(walletPrintPrivKeyCmd())

	return walletCmd
}

func walletNewAccount() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "new-account",
		Short: "Creates a new account",
		Run: func(cmd *cobra.Command, args []string) {
			password := getPassPhrase("Please enter a password", true)
			dataDir := getDataDirFromCmd(cmd)

			acc, err := wallet.NewKeystoreAccount(dataDir, password)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("New account created: %s\n", acc.Hex())
			fmt.Printf("Saved in %s\n", wallet.GetKeystoreDirPath(dataDir))
		},
	}

	addDefaultReqFlags(cmd)

	return cmd
}

func walletPrintPrivKeyCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "pk-print",
		Short: "Unlocks keystore file and prints the keys",
		Run: func(cmd *cobra.Command, args []string) {
			ksFile, _ := cmd.Flags().GetString(flagKeystoreFile)
			password := getPassPhrase("Please enter a password to decrypt wallet", false)

			keyJson, err := ioutil.ReadFile(ksFile)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			key, err := keystore.DecryptKey(keyJson, password)
			if err != nil {
				fmt.Println(err.Error())
			}

			spew.Dump(key)
		},
	}

	addKeystoreFlag(cmd)

	return cmd
}

func getPassPhrase(s string, confirmation bool) string {
	password, err := prompt.Stdin.PromptPassword(s)
	if err != nil {
		utils.Fatalf("Failed to read password %v", err)
	}

	if confirmation {
		confirm, err := prompt.Stdin.PromptPassword("Reapeat password")
		if err != nil {
			utils.Fatalf("Failed to read password confirmation %v", err)
		}
		if password != confirm {
			utils.Fatalf("Password do not match")
		}
	}

	return password
}
