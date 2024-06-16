package sloths

import (
	"encoding/json"
	"fmt"
	"strings"

	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdkflags "github.com/cosmos/cosmos-sdk/client/flags"
)

type Data struct {
	Tokens []string `json:"tokens"`
}

type queryNFTSOwnedResponse struct {
	Data Data `json:"data"`
}

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "sloths",
		Short:              "Query commands for sloths",
		DisableFlagParsing: true,
		RunE:               client.ValidateCmd,
	}

	cmd.AddCommand(QuerySlothsCmd())

	return cmd
}

func QuerySlothsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "owned-by [owner]",
		Short: "Query sloths owned by an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			owner := args[0]

			node, _ := cmd.Flags().GetString(sdkflags.FlagNode)
			nftContract, _ := cmd.Flags().GetString(flagNFTContract)
			mainnet, _ := cmd.Flags().GetBool(flagMainnet)
			testnet, _ := cmd.Flags().GetBool(flagTestnet)

			if !mainnet && !testnet &&
				(node == "" || nftContract == "") {
				return fmt.Errorf("missing required flags. Either set --mainnet or --testnet or provide the manual flags (--%s --%s)",
					sdkflags.FlagNode, flagNFTContract)
			}

			isStargaze := strings.HasPrefix(owner, "stars")
			if mainnet {
				// TODO: Set mainnet values (depending on isStargaze to set rpc and contract to query)
				return fmt.Errorf("mainnet not supported yet")
			} else if testnet {
				var networkInfo StaticNetworkInfo
				if isStargaze {
					networkInfo = Testnet.Stargaze
				} else {
					networkInfo = Testnet.Slotchain
				}

				nftContract = networkInfo.NFTContract
				node = networkInfo.Node

				// Needed because this flag is picked up later by the clientCtx
				if err := cmd.Flags().Set(sdkflags.FlagNode, node); err != nil {
					return err
				}
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := wasmdtypes.NewQueryClient(clientCtx)
			query := fmt.Sprintf(`{"tokens":{"owner":"%s"}}`, owner)
			res, err := queryClient.SmartContractState(
				cmd.Context(),
				&wasmdtypes.QuerySmartContractStateRequest{
					Address:   nftContract,
					QueryData: []byte(query),
				},
			)
			if err != nil {
				return err
			}

			queryNFTSStringOutput, err := clientCtx.Codec.MarshalJSON(res)
			if err != nil {
				return err
			}

			var nftsOwnedResponse queryNFTSOwnedResponse
			if err := json.Unmarshal(queryNFTSStringOutput, &nftsOwnedResponse); err != nil {
				return err
			}

			cmd.Printf("%d... sloths... found... for... %s...\n", len(nftsOwnedResponse.Data.Tokens), owner)
			for _, nft := range nftsOwnedResponse.Data.Tokens {
				cmd.Printf("🦥 #%s\n", nft)
			}
			if len(nftsOwnedResponse.Data.Tokens) != 0 {
				cmd.Println("too... much... work... time... to... 💤")
			}

			return nil
		},
	}

	cmd.Flags().Bool(flagMainnet, false, "Use mainnet values")
	cmd.Flags().Bool(flagTestnet, false, "Use testnet values")
	cmd.Flags().String(flagNFTContract, "", "NFT contract address")

	sdkflags.AddQueryFlagsToCmd(cmd)
	nodeFlag := cmd.Flags().Lookup(sdkflags.FlagNode)
	nodeFlag.DefValue = ""
	nodeFlag.Usage = "RPC endpoint of chain to query (stargaze or slothchain)"

	return cmd
}
