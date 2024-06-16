package sloths

import (
	"encoding/base64"
	"fmt"
	"github.com/gjermundgaraba/slothchain/cmd/slothchaind/cmd/lazycommandutils"
	"strings"
	"time"

	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdkflags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "sloths",
		Short:              "Transaction commands for sloths",
		DisableFlagParsing: true,
		RunE:               client.ValidateCmd,
	}

	cmd.AddCommand(TransferSlothCmd())

	return cmd
}

func TransferSlothCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer [from_address] [to_address] [nft-id]",
		Short: "Transfer sloth nfts between stargaze and slothchain using ICS721",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			from := args[0]
			to := args[1]
			nftID := args[2]

			node, _ := cmd.Flags().GetString(sdkflags.FlagNode)
			chainID, _ := cmd.Flags().GetString(sdkflags.FlagChainID)
			nftContract, _ := cmd.Flags().GetString(lazycommandutils.FlagNFTContract)
			ics721Contract, _ := cmd.Flags().GetString(lazycommandutils.FlagICS721Contract)
			ics721Channel, _ := cmd.Flags().GetString(lazycommandutils.FlagICS721Channel)
			waitForTx, _ := cmd.Flags().GetBool(lazycommandutils.FlagWaitForTx)

			mainnet, _ := cmd.Flags().GetBool(lazycommandutils.FlagMainnet)
			testnet, _ := cmd.Flags().GetBool(lazycommandutils.FlagTestnet)

			// Figure out if we are transferring from stargaze to slothchain or vice versa
			isStargaze := strings.HasPrefix(from, "stars")
			if isStargaze && !strings.HasPrefix(to, "lazy") {
				return fmt.Errorf("invalid addresses. Must transfer between stargaze and slothchain")
			}
			if !isStargaze && (!strings.HasPrefix(to, "stars") || !strings.HasPrefix(from, "lazy")) {
				return fmt.Errorf("invalid addresses. Must transfer between stargaze and slothchain")
			}

			if !mainnet && !testnet &&
				(node == "" || chainID == "" || nftContract == "" || ics721Contract == "" || ics721Channel == "") {
				return fmt.Errorf("missing required flags. Either set --mainnet or --testnet or provide the manual flags (--%s --%s --%s --%s --%s)",
					sdkflags.FlagNode, sdkflags.FlagChainID, lazycommandutils.FlagNFTContract, lazycommandutils.FlagICS721Contract, lazycommandutils.FlagICS721Channel)
			}

			if mainnet {
				// TODO: Set mainnet values (depending on isStargaze to set values)
				return fmt.Errorf("mainnet not supported yet")
			} else if testnet {
				var networkInfo lazycommandutils.StaticICS721NetworkInfo
				if isStargaze {
					networkInfo = lazycommandutils.ICS721Testnets.Stargaze
				} else {
					networkInfo = lazycommandutils.ICS721Testnets.Slotchain
				}

				chainID = networkInfo.ChainID
				if err := cmd.Flags().Set(sdkflags.FlagChainID, chainID); err != nil {
					return err
				}
				node = networkInfo.Node
				if err := cmd.Flags().Set(sdkflags.FlagNode, node); err != nil {
					return err
				}
				nftContract = networkInfo.NFTContract
				ics721Contract = networkInfo.ICS721Contract
				ics721Channel = networkInfo.ICS721Channel
			}

			msg := createTransferMsg(from, to, nftID, nftContract, ics721Contract, ics721Channel)

			if err := cmd.Flags().Set(sdkflags.FlagFrom, from); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if waitForTx {
				if err := lazycommandutils.SendAndWaitForTx(clientCtx, cmd.Flags(), &msg); err != nil {
					return err
				}

				fmt.Printf("🦥 lazy... transfer... of... sloth #%s... to... %s... done...\n", nftID, to)
				fmt.Printf("🦥 tx... finally... done... time... too... 💤!\n")

				return nil
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	cmd.Flags().Bool(lazycommandutils.FlagWaitForTx, true, "Wait for transaction to be included in a block")
	cmd.Flags().String(lazycommandutils.FlagICS721Contract, "", "ICS721 contract address")
	cmd.Flags().String(lazycommandutils.FlagICS721Channel, "", "ICS721 channel")
	cmd.Flags().String(lazycommandutils.FlagNFTContract, "", "NFT contract address")
	cmd.Flags().Bool(lazycommandutils.FlagMainnet, false, "Use mainnet (overrides transfer flags)")
	cmd.Flags().Bool(lazycommandutils.FlagTestnet, false, "Use testnet (overrides transfer flags)")

	sdkflags.AddTxFlagsToCmd(cmd)
	nodeFlag := cmd.Flags().Lookup(sdkflags.FlagNode)
	nodeFlag.Usage = "RPC endpoint of sending chain (stargaze or slothchain)"
	nodeFlag.DefValue = ""

	cmd.Flags().Lookup(sdkflags.FlagChainID).Usage = "Chain ID of sending chain (stargaze or slothchain)"

	return cmd
}

func createTransferMsg(from string, to string, nftID string, nftContract string, ics721Contract string, ics721channel string) wasmdtypes.MsgExecuteContract {
	now := time.Now()
	fiveMinutesLater := now.Add(5 * time.Minute) // TODO: Maybe more...
	sendExecMsg := fmt.Sprintf("{\"receiver\": \"%s\",\n\"channel_id\": \"%s\",\n\"timeout\": { \"timestamp\": \"%d\"}}",
		to,
		ics721channel,
		fiveMinutesLater.UnixNano(),
	)
	sendExecMsgBase64 := base64.StdEncoding.EncodeToString([]byte(sendExecMsg))

	execMsg := fmt.Sprintf(`{
  "send_nft": {
    "contract": "%s", 
    "token_id": "%s", 
    "msg": "%s"}
}`, ics721Contract, nftID, sendExecMsgBase64)
	return wasmdtypes.MsgExecuteContract{
		Sender:   from,
		Contract: nftContract,
		Msg:      []byte(execMsg),
	}
}
