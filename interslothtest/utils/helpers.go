package utils

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"path"
	"strconv"
	"strings"
	"time"
)

type NFTSetup struct {
	SGICS721Contract                      string
	SlothChainICS721Contract              string
	SlothChainICS721IncomingProxyContract string
	SlothContract                         string
	SGPort                                string
	SGChannel                             string
	SlothPort                             string
	SlothChainChannel                     string
}

func (s *InterchainValues) DeployNFTSetup(sgUser ibc.Wallet, slothUser ibc.Wallet, artifactsPath string) NFTSetup {
	sgCW721CodeID := s.StoreCW721(s.Stargaze, sgUser.KeyName(), artifactsPath)
	slothCW721CodeID := s.StoreCW721(s.Slothchain, slothUser.KeyName(), artifactsPath)

	slothContract := s.DeploySloths(slothCW721CodeID, sgUser)
	s.MintNFTs(slothContract, sgUser.KeyName(), sgUser.FormattedAddress(), []string{"1", "2", "3"})

	// Deploy incoming proxy contract on Slothchain
	slothChainICS721Contract := s.DeployICS721(s.Slothchain, slothUser, artifactsPath, slothCW721CodeID)

	sgICS721Contract := s.DeployICS721(s.Stargaze, sgUser, artifactsPath, sgCW721CodeID)

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

	ics721IncomingProxyContract := s.DeployICS721IncomingProxy(s.Slothchain, slothUser, artifactsPath, slothChainICS721Contract, slothContract, slothChainChannel)
	s.MigrateICS721(s.Slothchain, slothUser.KeyName(), artifactsPath, slothChainICS721Contract, ics721IncomingProxyContract, slothCW721CodeID)

	return NFTSetup{
		SGICS721Contract:                      sgICS721Contract,
		SlothChainICS721Contract:              slothChainICS721Contract,
		SlothChainICS721IncomingProxyContract: ics721IncomingProxyContract,
		SlothContract:                         slothContract,
		SGChannel:                             sgChannel,
		SGPort:                                sgPortName,
		SlothChainChannel:                     slothChainChannel,
		SlothPort:                             slothPortName,
	}
}

func (s *InterchainValues) StoreCW721(chain *cosmos.CosmosChain, userKeyName string, artifactsFolderPath string) string {
	contractPath := path.Join(artifactsFolderPath, "cw721_base.wasm")
	codeID, err := chain.StoreContract(s.Ctx, userKeyName, contractPath, "--gas", "auto", "--gas-adjustment", "2")
	s.NoError(err)

	return codeID
}

func (s *InterchainValues) DeployICS721(
	chain *cosmos.CosmosChain,
	user ibc.Wallet,
	artifactsFolderPath string,
	cw721CodeID string,
) string {
	contractPath := path.Join(artifactsFolderPath, "ics721_base.wasm")
	sgICS721CodeID, err := chain.StoreContract(s.Ctx, user.KeyName(), contractPath, "--gas", "auto", "--gas-adjustment", "2")
	s.NoError(err)

	ics721InstantiateMsg := fmt.Sprintf("{\"cw721_base_code_id\": %s}", cw721CodeID)

	ics721Contract, err := chain.InstantiateContract(s.Ctx, user.KeyName(), sgICS721CodeID, ics721InstantiateMsg, false, "--admin", user.FormattedAddress(), "--gas", "auto", "--gas-adjustment", "2")
	s.NoError(err)

	return ics721Contract
}

func (s *InterchainValues) DeployICS721IncomingProxy(
	chain *cosmos.CosmosChain,
	user ibc.Wallet,
	artifactsFolderPath string,
	ics721Contract string,
	classID string,
	ics721Channel string,
) string {
	contractPath := path.Join(artifactsFolderPath, "cw_ics721_incoming_proxy_base.wasm")
	codeID, err := chain.StoreContract(s.Ctx, user.KeyName(), contractPath, "--gas", "auto", "--gas-adjustment", "2")
	s.NoError(err)

	incomingProxyInstantiateMsg := fmt.Sprintf("{\"origin\": \"%s\", \"class_ids\": [\"%s\"], \"channels\": [\"%s\"]}", ics721Contract, classID, ics721Channel)
	incomingProxyContract, err := chain.InstantiateContract(s.Ctx, user.KeyName(), codeID, incomingProxyInstantiateMsg, false, "--admin", user.FormattedAddress(), "--gas", "auto", "--gas-adjustment", "2")
	s.NoError(err)

	return incomingProxyContract
}

func (s *InterchainValues) MigrateICS721IncomingProxy(
	chain *cosmos.CosmosChain,
	userKeyName string,
	artifactsFolderPath string,
	ics721IncomingProxyContract string,
	classID string,
	ics721Contract string,
	ics721Channel string,
) {
	contractPath := path.Join(artifactsFolderPath, "cw_ics721_incoming_proxy_base.wasm")
	codeID, err := chain.StoreContract(s.Ctx, userKeyName, contractPath, "--gas", "auto", "--gas-adjustment", "2")
	s.NoError(err)

	migrateMsg := fmt.Sprintf("{\"with_update\": {\"origin\": \"%s\", \"class_ids\": [\"%s\"], \"channels\": [\"%s\"]}}", ics721Contract, classID, ics721Channel)

	txHash, err := chain.GetNode().ExecTx(
		s.Ctx,
		userKeyName,
		"wasm",
		"migrate", ics721IncomingProxyContract, codeID, migrateMsg,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", userKeyName,
		"--gas", "500000",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	)
	s.NoError(err)

	txResp, err := chain.GetTransaction(txHash)
	s.NoError(err)
	s.Equal(uint32(0), txResp.Code, txResp.RawLog)
}

func (s *InterchainValues) MigrateICS721(
	chain *cosmos.CosmosChain,
	userKeyName string,
	artifactsFolderPath string,
	ics721Contract string,
	ics721IncomingProxy string,
	cw721CodeID string,
) {
	contractPath := path.Join(artifactsFolderPath, "ics721_base.wasm")
	codeID, err := chain.StoreContract(s.Ctx, userKeyName, contractPath, "--gas", "auto", "--gas-adjustment", "2")
	s.NoError(err)

	migrateMsg := fmt.Sprintf("{\"with_update\": {\"incoming_proxy\": \"%s\", \"cw721_base_code_id\": %s}}", ics721IncomingProxy, cw721CodeID)

	txHash, err := chain.GetNode().ExecTx(
		s.Ctx,
		userKeyName,
		"wasm",
		"migrate", ics721Contract, codeID, migrateMsg,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", userKeyName,
		"--gas", "500000",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	)
	s.NoError(err)

	txResp, err := chain.GetTransaction(txHash)
	s.NoError(err)
	s.Equal(uint32(0), txResp.Code, txResp.RawLog)
}

func (s *InterchainValues) DeploySloths(slothCodeID string, sgUser ibc.Wallet) string {
	return s.InstantiateCW721(slothCodeID, sgUser.KeyName(), "Celestine Sloth Society", "CSS", sgUser.FormattedAddress())
}

func (s *InterchainValues) InstantiateCW721(contractAddress string, userKeyName string, name string, symbol string, minter string) string {
	instantiateMsg := fmt.Sprintf("{\"name\": \"%s\", \"symbol\": \"%s\", \"minter\": \"%s\"}", name, symbol, minter)
	contract, err := s.Stargaze.InstantiateContract(s.Ctx, userKeyName, contractAddress, instantiateMsg, true, "--gas", "auto", "--gas-adjustment", "2")
	s.NoError(err)

	return contract
}

func (s *InterchainValues) MintNFTs(cw721Contract string, sgUserKeyName string, mintTo string, tokenIds []string) {
	for _, tokenId := range tokenIds {
		slothCW721MintMsg := fmt.Sprintf("{\"mint\": {\"token_id\": \"%s\", \"owner\": \"%s\"}}", tokenId, mintTo)
		_, err := s.Stargaze.ExecuteContract(s.Ctx, sgUserKeyName, cw721Contract, slothCW721MintMsg)
		s.NoError(err)
	}
}

func (s *InterchainValues) TransferSlothToSlothChain(nftSetup NFTSetup, from ibc.Wallet, to ibc.Wallet, tokenId string) (classID string, contractAddress string) {
	s.NoError(s.TransferNFT(s.Stargaze, from, to, tokenId, nftSetup.SlothContract, nftSetup.SGICS721Contract, nftSetup.SGChannel))

	s.NoError(testutil.WaitForBlocks(s.Ctx, 10, s.Stargaze, s.Slothchain, s.Celestia))

	type Response struct {
		Data [][]string `json:"data"`
	}
	var resp Response
	s.NoError(s.Slothchain.QueryContract(s.Ctx, nftSetup.SlothChainICS721Contract, "{\"nft_contracts\": {}}", &resp))

	s.Len(resp.Data, 1)

	return resp.Data[0][0], resp.Data[0][1]
}

func (s *InterchainValues) TransferSlothToStargaze(nftSetup NFTSetup, from ibc.Wallet, to ibc.Wallet, tokenId string, slothContractOnSlothChain string) {
	s.NoError(s.TransferNFT(s.Slothchain, from, to, tokenId, slothContractOnSlothChain, nftSetup.SlothChainICS721Contract, nftSetup.SlothChainChannel))
}

func (s *InterchainValues) TransferNFT(chain *cosmos.CosmosChain, from ibc.Wallet, to ibc.Wallet, tokenID string, nftContract string, ics721Contract string, channel string) error {
	now := time.Now()
	fiveMinutesLater := now.Add(5 * time.Minute)
	sendExecMsg := fmt.Sprintf("{\"receiver\": \"%s\",\n\"channel_id\": \"%s\",\n\"timeout\": { \"timestamp\": \"%d\"}}",
		to.FormattedAddress(),
		channel,
		fiveMinutesLater.UnixNano(),
	)
	sendExecMsgBase64 := base64.StdEncoding.EncodeToString([]byte(sendExecMsg))

	transferMsg := fmt.Sprintf("{\"send_nft\": {\"contract\": \"%s\", \"token_id\": \"%s\", \"msg\": \"%s\"}}",
		ics721Contract,
		tokenID,
		sendExecMsgBase64,
	)

	_, err := chain.ExecuteContract(s.Ctx, from.KeyName(), nftContract, transferMsg, "--gas", "auto", "--gas-adjustment", "2")
	return err
}

func (s *E2ETestSuite) AllNFTs(chain *cosmos.CosmosChain, contractAddress string) []string {
	type Response struct {
		Data struct {
			Tokens []string `json:"tokens"`
		} `json:"data"`
	}
	var resp Response
	s.NoError(chain.QueryContract(s.Ctx, contractAddress, "{\"all_tokens\": {}}", &resp))

	return resp.Data.Tokens
}

// Different versions makes the normal helper methods fail, so the celestia transfer is done more manually:
func (s *InterchainValues) CelestiaIBCTransfer(channelID string, celestiaUserKeyName string, celestiaTransfer ibc.WalletAmount) {
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

// AssertPacketRelayed asserts that the packet commitment does not exist on the sending chain.
// The packet commitment will be deleted upon a packet acknowledgement or timeout.
func (s *InterchainValues) AssertPacketRelayed(chain *cosmos.CosmosChain, portID, channelID string, sequence uint64) {
	_, err := GRPCQuery[channeltypes.QueryPacketCommitmentResponse](s.Ctx, chain, &channeltypes.QueryPacketCommitmentRequest{
		PortId:    portID,
		ChannelId: channelID,
		Sequence:  sequence,
	})
	s.ErrorContains(err, "packet commitment hash not found")
}

// GRPCQuery queries the chain with a query request and deserializes the response to T
func GRPCQuery[T any](ctx context.Context, chain ibc.Chain, req proto.Message, opts ...grpc.CallOption) (*T, error) {
	path, err := getProtoPath(req)
	if err != nil {
		return nil, err
	}

	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	defer grpcConn.Close()

	resp := new(T)
	err = grpcConn.Invoke(ctx, path, req, resp, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getProtoPath(req proto.Message) (string, error) {
	typeURL := "/" + proto.MessageName(req)

	switch {
	case strings.Contains(typeURL, "Query"):
		return getQueryProtoPath(typeURL)
	case strings.Contains(typeURL, "cosmos.base.tendermint"):
		return getCmtProtoPath(typeURL)
	default:
		return "", fmt.Errorf("unsupported typeURL: %s", typeURL)
	}
}

func getQueryProtoPath(queryTypeURL string) (string, error) {
	queryIndex := strings.Index(queryTypeURL, "Query")
	if queryIndex == -1 {
		return "", fmt.Errorf("invalid typeURL: %s", queryTypeURL)
	}

	// Add to the index to account for the length of "Query"
	queryIndex += len("Query")

	// Add a slash before the query
	urlWithSlash := queryTypeURL[:queryIndex] + "/" + queryTypeURL[queryIndex:]
	if !strings.HasSuffix(urlWithSlash, "Request") {
		return "", fmt.Errorf("invalid typeURL: %s", queryTypeURL)
	}

	return strings.TrimSuffix(urlWithSlash, "Request"), nil
}

func getCmtProtoPath(cmtTypeURL string) (string, error) {
	cmtIndex := strings.Index(cmtTypeURL, "Get")
	if cmtIndex == -1 {
		return "", fmt.Errorf("invalid typeURL: %s", cmtTypeURL)
	}

	// Add a slash before the commitment
	urlWithSlash := cmtTypeURL[:cmtIndex] + "Service/" + cmtTypeURL[cmtIndex:]
	if !strings.HasSuffix(urlWithSlash, "Request") {
		return "", fmt.Errorf("invalid typeURL: %s", cmtTypeURL)
	}

	return strings.TrimSuffix(urlWithSlash, "Request"), nil
}

// QueryTxsByEvents runs the QueryTxsByEvents command on the given chain.
// https://github.com/cosmos/cosmos-sdk/blob/65ab2530cc654fd9e252b124ed24cbaa18023b2b/x/auth/client/cli/query.go#L33
func (s *InterchainValues) QueryTxsByEvents(
	chain ibc.Chain,
	page, limit int, queryReq, orderBy string,
) (*sdk.SearchTxsResult, error) {
	cosmosChain, ok := chain.(*cosmos.CosmosChain)
	if !ok {
		return nil, fmt.Errorf("QueryTxsByEvents must be passed a cosmos.CosmosChain")
	}

	cmd := []string{"txs"}

	cmd = append(cmd, "--query", queryReq)
	// cmd = append(cmd, "--events", queryReq) ??

	if orderBy != "" {
		cmd = append(cmd, "--order_by", orderBy)
	}
	if page != 0 {
		cmd = append(cmd, "--"+flags.FlagPage, strconv.Itoa(page))
	}
	if limit != 0 {
		cmd = append(cmd, "--"+flags.FlagLimit, strconv.Itoa(limit))
	}

	stdout, _, err := cosmosChain.GetNode().ExecQuery(s.Ctx, cmd...)
	if err != nil {
		return nil, err
	}

	result := &sdk.SearchTxsResult{}
	err = getEncodingConfig().Codec.UnmarshalJSON(stdout, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ExtractValueFromEvents extracts the value of an attribute from a list of events.
// If the attribute is not found, the function returns an empty string and false.
// If the attribute is found, the function returns the value and true.
func (*E2ETestSuite) ExtractValueFromEvents(events []abci.Event, eventType, attrKey string) (string, bool) {
	for _, event := range events {
		if event.Type != eventType {
			continue
		}

		for _, attr := range event.Attributes {
			if attr.Key != attrKey {
				continue
			}

			return attr.Value, true
		}
	}

	return "", false
}
