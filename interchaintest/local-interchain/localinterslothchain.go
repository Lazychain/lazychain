package main

import (
	"cosmossdk.io/math"
	"fmt"
	"github.com/Lazychain/lazychain/interchaintest/utils"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"os"
	"os/signal"
	"syscall"
)

const mnemonic = "curve govern feature draw giggle one enemy shop wonder cross castle oxygen business obscure rule detail chaos dirt pause parrot tail lunch merit rely"

type LocalInterchain struct {
	utils.InterchainValues
}

func main() {
	fmt.Println("Running LocalInterchain... 💤")

	interchainValues := LocalInterchain{}
	interchainValues.SetupFakeT("LocalInterchain")

	defer func() {
		interchainValues.GetFakeT().ActuallyRunCleanups()
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		interchainValues.GetFakeT().ActuallyRunCleanups()
		os.Exit(1)
	}()

	interchainValues.SetupInterchainValues()
	interchainValues.TestLocalInterChain()
}

func (s *LocalInterchain) TestLocalInterChain() {

	slothUser, err := interchaintest.GetAndFundTestUserWithMnemonic(s.Ctx, "user", mnemonic, math.NewInt(10_000_000_000), s.LazyChain)
	s.NoError(err)

	sgUser, err := interchaintest.GetAndFundTestUserWithMnemonic(s.Ctx, "user", mnemonic, math.NewInt(10_000_000_000), s.Stargaze)
	s.NoError(err)

	celestiaUser, err := interchaintest.GetAndFundTestUserWithMnemonic(s.Ctx, "user", mnemonic, math.NewInt(10_000_000_000), s.Celestia)
	s.NoError(err)

	nftSetup := s.DeployNFTSetup(sgUser, slothUser, "./test-artifacts")

	s.NoError(s.Relayer.StartRelayer(s.Ctx, s.RelayerExecRep, s.StargazeSlothPath))
	s.TT().Cleanup(
		func() {
			err := s.Relayer.StopRelayer(s.Ctx, s.RelayerExecRep)
			if err != nil {
				s.TT().Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)
	s.NoError(testutil.WaitForBlocks(s.Ctx, 5, s.Stargaze, s.LazyChain))

	celestiaToLazyChainChannel, err := ibc.GetTransferChannel(s.Ctx, s.Relayer, s.RelayerExecRep, s.Celestia.Config().ChainID, s.LazyChain.Config().ChainID)
	s.NoError(err)

	slothContainer, err := s.LazyChain.GetNode().DockerClient.ContainerInspect(s.Ctx, s.LazyChain.GetNode().ContainerID())
	s.NoError(err)

	stargazeContainer, err := s.Stargaze.GetNode().DockerClient.ContainerInspect(s.Ctx, s.Stargaze.GetNode().ContainerID())
	s.NoError(err)

	celestialContainer, err := s.Celestia.GetNode().DockerClient.ContainerInspect(s.Ctx, s.Celestia.GetNode().ContainerID())
	s.NoError(err)

	fmt.Println("Local interchain is now running...")
	fmt.Println()
	fmt.Println("Users, all with the mnemonic:", mnemonic)
	fmt.Println("Sloth user address:", slothUser.FormattedAddress())
	fmt.Println("Stargaze user address:", sgUser.FormattedAddress())
	fmt.Println("Celestia user address:", celestiaUser.FormattedAddress())
	fmt.Println()
	fmt.Println("LazyChain chain-id:", s.LazyChain.Config().ChainID)
	fmt.Printf("LazyChain RPC address: tcp://localhost:%s\n", slothContainer.NetworkSettings.Ports["26657/tcp"][0].HostPort)
	fmt.Println("Stargaze chain-id:", s.Stargaze.Config().ChainID)
	fmt.Printf("Stargaze RPC address: tcp://localhost:%s\n", stargazeContainer.NetworkSettings.Ports["26657/tcp"][0].HostPort)
	fmt.Println("Celestia chain-id:", s.Celestia.Config().ChainID)
	fmt.Printf("Celestia RPC address: tcp://localhost:%s\n", celestialContainer.NetworkSettings.Ports["26657/tcp"][0].HostPort)
	fmt.Println()
	fmt.Println("ICS721 setup deployed")
	fmt.Println("ICS721 contract on Stargaze:", nftSetup.SGICS721Contract)
	fmt.Println("ICS721 contract on Sloth chain:", nftSetup.LazyChainICS721Contract)
	fmt.Println("Sloth contract:", nftSetup.CelestineSlothsContract)
	fmt.Println("Stargaze to Sloth channel:", nftSetup.SGChannel)
	fmt.Println("Sloth chain to Stargaze channel:", nftSetup.LazyChainChannel)
	fmt.Println("Celestia to Sloth channel:", celestiaToLazyChainChannel.ChannelID)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop...")

	select {}
}
