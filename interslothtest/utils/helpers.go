package utils

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"time"
)

type NFTSetup struct {
	SGICS721Contract         string
	SlothChainICS721Contract string
	SlothContract            string
	SGChannel                string
	SlothChainChannel        string
}

func (s *E2ETestSuite) DeployNFTSetup(sgUser ibc.Wallet, slothUser ibc.Wallet) NFTSetup {
	sgCW721CodeID := s.DeployCW721(s.Stargaze, sgUser.KeyName())
	slothCW721CodeID := s.DeployCW721(s.Slothchain, slothUser.KeyName())

	slothContract := s.DeploySloths(slothCW721CodeID, sgUser)
	s.MintSloths(slothContract, sgUser.KeyName(), sgUser.FormattedAddress(), []string{"1", "2", "3"})

	sgICS721Contract := s.DeployICS721(s.Stargaze, sgUser.KeyName(), sgCW721CodeID)
	slothChainICS721Contract := s.DeployICS721(s.Slothchain, slothUser.KeyName(), slothCW721CodeID)

	//ics721PathName := "ics721"
	//s.NoError(s.Relayer.GeneratePath(s.Ctx, s.RelayerExecRep, s.Stargaze.Config().ChainID, s.Slothchain.Config().ChainID, ics721PathName))

	sgPortName := fmt.Sprintf("wasm.%s", sgICS721Contract)
	slothPortName := fmt.Sprintf("wasm.%s", slothChainICS721Contract)
	channelOpts := ibc.CreateChannelOptions{
		SourcePortName: sgPortName,
		DestPortName:   slothPortName,
		Order:          ibc.Unordered,
		Version:        "ics721-1",
	}
	clientOpts := ibc.CreateClientOptions{}
	s.NoError(s.Relayer.LinkPath(s.Ctx, s.RelayerExecRep, s.StargazeSlothPath, channelOpts, clientOpts))

	channels, err := s.Relayer.GetChannels(s.Ctx, s.RelayerExecRep, s.Stargaze.Config().ChainID)
	s.NoError(err)
	var sgChannel string
	var slothChainChannel string
	for _, channel := range channels {
		if channel.PortID == sgPortName {
			sgChannel = channel.ChannelID
			slothChainChannel = channel.Counterparty.ChannelID
		}
	}
	s.NotEmpty(sgChannel)
	s.NotEmpty(slothChainChannel)

	return NFTSetup{
		SGICS721Contract:         sgICS721Contract,
		SlothChainICS721Contract: slothChainICS721Contract,
		SlothContract:            slothContract,
		SGChannel:                sgChannel,
		SlothChainChannel:        slothChainChannel,
	}
}

func (s *E2ETestSuite) DeployCW721(chain *cosmos.CosmosChain, userKeyName string) string {
	codeID, err := chain.StoreContract(s.Ctx, userKeyName, "../../artifacts/cw721_base.wasm", "--gas", "auto", "--gas-adjustment", "2")
	s.NoError(err)

	return codeID
}

func (s *E2ETestSuite) DeployICS721(chain *cosmos.CosmosChain, userKeyName string, cw721CodeID string) string {
	sgICS721CodeID, err := chain.StoreContract(s.Ctx, userKeyName, "../../artifacts/ics721_base.wasm")
	s.NoError(err)

	sgICS721InstantiateMsg := fmt.Sprintf("{\"cw721_base_code_id\": %s}", cw721CodeID)
	sgICS721Contract, err := chain.InstantiateContract(s.Ctx, userKeyName, sgICS721CodeID, sgICS721InstantiateMsg, true, "--gas", "auto", "--gas-adjustment", "2")
	s.NoError(err)

	return sgICS721Contract
}

func (s *E2ETestSuite) DeploySloths(slothContract string, sgUser ibc.Wallet) string {
	slothCW721InstantiateMsg := fmt.Sprintf("{\"name\": \"Celestine Sloth Society\", \"symbol\": \"CSS\", \"minter\": \"%s\"}", sgUser.FormattedAddress())
	slothCW721Contract, err := s.Stargaze.InstantiateContract(s.Ctx, sgUser.KeyName(), slothContract, slothCW721InstantiateMsg, true, "--gas", "auto", "--gas-adjustment", "2")
	s.NoError(err)

	return slothCW721Contract
}

func (s *E2ETestSuite) MintSloths(cw721Contract string, sgUserKeyName string, mintTo string, tokenIds []string) {
	for _, tokenId := range tokenIds {
		slothCW721MintMsg := fmt.Sprintf("{\"mint\": {\"token_id\": \"%s\", \"owner\": \"%s\"}}", tokenId, mintTo)
		_, err := s.Stargaze.ExecuteContract(s.Ctx, sgUserKeyName, cw721Contract, slothCW721MintMsg)
		s.NoError(err)
	}
}

func (s *E2ETestSuite) TransferSlothToSlothChain(nftSetup NFTSetup, from ibc.Wallet, to ibc.Wallet, tokenId string) (classID string, contractAddress string) {
	now := time.Now()
	fiveMinutesLater := now.Add(5 * time.Minute)
	sendExecMsg := fmt.Sprintf("{\"receiver\": \"%s\",\n\"channel_id\": \"%s\",\n\"timeout\": { \"timestamp\": \"%d\"}}",
		to.FormattedAddress(),
		nftSetup.SGChannel,
		fiveMinutesLater.UnixNano(),
	)
	sendExecMsgBase64 := base64.StdEncoding.EncodeToString([]byte(sendExecMsg))

	transferMsg := fmt.Sprintf("{\"send_nft\": {\"contract\": \"%s\", \"token_id\": \"%s\", \"msg\": \"%s\"}}",
		nftSetup.SGICS721Contract,
		tokenId,
		sendExecMsgBase64,
	)

	_, err := s.Stargaze.ExecuteContract(s.Ctx, from.KeyName(), nftSetup.SlothContract, transferMsg, "--gas", "auto", "--gas-adjustment", "2")
	s.NoError(err)

	s.NoError(testutil.WaitForBlocks(s.Ctx, 10, s.Stargaze, s.Slothchain, s.Celestia))

	type Response struct {
		Data [][]string `json:"data"`
	}
	var resp Response
	s.NoError(s.Slothchain.QueryContract(s.Ctx, nftSetup.SlothChainICS721Contract, "{\"nft_contracts\": {}}", &resp))

	s.Len(resp.Data, 1)

	return resp.Data[0][0], resp.Data[0][1]
}

func (s *E2ETestSuite) AllNFTs(contractAddress string) []string {
	type Response struct {
		Data struct {
			Tokens []string `json:"tokens"`
		} `json:"data"`
	}
	var resp Response
	s.NoError(s.Slothchain.QueryContract(s.Ctx, contractAddress, "{\"all_tokens\": {}}", &resp))

	return resp.Data.Tokens
}

// Different versions makes the normal helper methods fail, so the celestia transfer is done more manually:
func (s *E2ETestSuite) CelestiaIBCTransfer(channelID string, celestiaUserKeyName string, celestiaTransfer ibc.WalletAmount) {
	// Different versions makes the helper methods fail, so the celestia transfer is done more manually:
	txHash, err := s.Celestia.GetNode().SendIBCTransfer(s.Ctx, channelID, celestiaUserKeyName, celestiaTransfer, ibc.TransferOptions{})
	s.NoError(err)
	rpcNode, err := s.Celestia.GetNode().CliContext().GetNode()
	s.NoError(err)
	hash, err := hex.DecodeString(txHash)
	s.NoError(err)
	resTx, err := rpcNode.Tx(s.Ctx, hash, false)
	s.NoError(err)
	s.Equal(uint32(0), resTx.TxResult.Code)
}
