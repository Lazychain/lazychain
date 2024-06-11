package sloths

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// SendAndWaitForTx is a helper function to send a series of messages and wait for a transaction to be included in a block.
// The code here is copied of the tx.go file from the Cosmos SDK and modified to wait for the transaction to be included and then return the response
func SendAndWaitForTx(clientCtx client.Context, flagSet *pflag.FlagSet, msgs ...sdk.Msg) error {
	txf, err := tx.NewFactoryCLI(clientCtx, flagSet)
	if err != nil {
		return err
	}

	// Validate all msgs before generating or broadcasting the tx.
	// We were calling ValidateBasic separately in each CLI handler before.
	// Right now, we're factorizing that call inside this function.
	// ref: https://github.com/cosmos/cosmos-sdk/pull/9236#discussion_r623803504
	for _, msg := range msgs {
		m, ok := msg.(sdk.HasValidateBasic)
		if !ok {
			continue
		}

		if err := m.ValidateBasic(); err != nil {
			return err
		}
	}

	if clientCtx.GenerateOnly {
		return txf.PrintUnsignedTx(clientCtx, msgs...)
	}

	txf, err = txf.Prepare(clientCtx)
	if err != nil {
		return err
	}

	if txf.SimulateAndExecute() || clientCtx.Simulate {
		if clientCtx.Offline {
			return errors.New("cannot estimate gas in offline mode")
		}

		_, adjusted, err := tx.CalculateGas(clientCtx, txf, msgs...)
		if err != nil {
			return err
		}

		txf = txf.WithGas(adjusted)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", tx.GasEstimateResponse{GasEstimate: txf.Gas()})
	}

	if clientCtx.Simulate {
		return nil
	}

	unsignedTx, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return err
	}

	if !clientCtx.SkipConfirm {
		encoder := clientCtx.TxConfig.TxJSONEncoder()
		if encoder == nil {
			return errors.New("failed to encode transaction: tx json encoder is nil")
		}

		txBytes, err := encoder(unsignedTx.GetTx())
		if err != nil {
			return fmt.Errorf("failed to encode transaction: %w", err)
		}

		if err := clientCtx.PrintRaw(json.RawMessage(txBytes)); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\n%s\n", err, txBytes)
		}

		buf := bufio.NewReader(os.Stdin)
		ok, err := input.GetConfirmation("confirm transaction before signing and broadcasting", buf, os.Stderr)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\ncanceled transaction\n", err)
			return err
		}
		if !ok {
			_, _ = fmt.Fprintln(os.Stderr, "canceled transaction")
			return nil
		}
	}

	if err = tx.Sign(clientCtx.CmdContext, txf, clientCtx.FromName, unsignedTx, true); err != nil {
		return err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(unsignedTx.GetTx())
	if err != nil {
		return err
	}

	// broadcast to a CometBFT node
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	if res.Code != 0 {
		return fmt.Errorf(res.RawLog)
	}
	try := 1
	maxTries := 200
	for {
		if try > maxTries {
			return fmt.Errorf("failed to wait for %s. Maximum number of tries reached", res.TxHash)
		}

		txResp, err := authtx.QueryTx(clientCtx, res.TxHash)
		if err != nil {
			fmt.Print("\033[G\033[K") // move the cursor left and clear the line
			fmt.Printf("🦥 taking... soo... long... for... %s - attempt %d/%d 💤", res.TxHash, try, maxTries)
			time.Sleep(500 * time.Millisecond)
			try++
			continue
		}

		if txResp.Code != 0 {
			panic(fmt.Errorf("transaction failed: %s", txResp.RawLog))
		}

		fmt.Print("\033[G\033[K") // move the cursor left and clear the line
		fmt.Printf("🦥 tx... %s... completed...\n", res.TxHash)

		// TODO: Add a flag for IBC transfers to wait for the packet to be received

		return nil
	}
}
