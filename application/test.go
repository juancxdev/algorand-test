package application

import (
	"context"
	"encoding/base64"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/examples"
	"github.com/algorand/go-algorand-sdk/v2/transaction"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"io/ioutil"
	"log"
)

func AppCreate(algodClient *algod.Client, creator crypto.Account) uint64 {
	// example: APP_SCHEMA
	// declare application state storage (immutable)
	var (
		localInts   uint64 = 0
		localBytes  uint64 = 0
		globalInts  uint64 = 0
		globalBytes uint64 = 0
	)

	// define schema
	globalSchema := types.StateSchema{NumUint: globalInts, NumByteSlice: globalBytes}
	localSchema := types.StateSchema{NumUint: localInts, NumByteSlice: localBytes}
	// example: APP_SCHEMA

	// example: APP_SOURCE
	//approvalTeal, err := ioutil.ReadFile("./BjaguarTransactions.approval.teal")
	approvalTeal, err := ioutil.ReadFile("./BjaguarTransactions.approval.teal")
	if err != nil {
		log.Fatalf("failed to read approval program: %s", err)
	}
	//clearTeal, err := ioutil.ReadFile("./BjaguarTransactions.clear.teal")
	clearTeal, err := ioutil.ReadFile("./BjaguarTransactions.clear.teal")
	if err != nil {
		log.Fatalf("failed to read clear program: %s", err)
	}
	// example: APP_SOURCE

	// example: APP_COMPILE
	approvalResult, err := algodClient.TealCompile(approvalTeal).Do(context.Background())
	if err != nil {
		log.Fatalf("failed to compile program: %s", err)
	}

	approvalBinary, err := base64.StdEncoding.DecodeString(approvalResult.Result)
	if err != nil {
		log.Fatalf("failed to decode compiled program: %s", err)
	}

	clearResult, err := algodClient.TealCompile(clearTeal).Do(context.Background())
	if err != nil {
		log.Fatalf("failed to compile program: %s", err)
	}

	clearBinary, err := base64.StdEncoding.DecodeString(clearResult.Result)
	if err != nil {
		log.Fatalf("failed to decode compiled program: %s", err)
	}
	// example: APP_COMPILE

	// example: APP_CREATE
	// Create application
	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		log.Fatalf("error getting suggested tx params: %s", err)
	}

	txn, err := transaction.MakeApplicationCreateTx(
		false, approvalBinary, clearBinary, globalSchema, localSchema,
		nil, nil, nil, nil, sp, creator.Address, nil,
		types.Digest{}, [32]byte{}, types.ZeroAddress,
	)
	if err != nil {
		log.Fatalf("failed to make txn: %s", err)
	}

	txid, stx, err := crypto.SignTransaction(creator.PrivateKey, txn)
	if err != nil {
		log.Fatalf("failed to sign transaction: %s", err)
	}

	_, err = algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		log.Fatalf("failed to send transaction: %s", err)
	}

	confirmedTxn, err := transaction.WaitForConfirmation(algodClient, txid, 4, context.Background())
	if err != nil {
		log.Fatalf("error waiting for confirmation:  %s", err)
	}
	appID := confirmedTxn.ApplicationIndex
	log.Printf("Created app with id: %d", appID)
	// example: APP_CREATE
	return appID
}

func AppOptIn(algodClient *algod.Client, appID uint64, caller crypto.Account) {
	// example: APP_OPTIN
	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		log.Fatalf("error getting suggested tx params: %s", err)
	}

	// Create a new clawback transaction with the target of the user address and the recipient as the creator
	// address, being sent from the address marked as `clawback` on the asset, in this case the same as creator
	txn, err := transaction.MakeApplicationOptInTx(
		appID, nil, nil, nil, nil, sp,
		caller.Address, nil, types.Digest{}, [32]byte{}, types.ZeroAddress,
	)
	if err != nil {
		log.Fatalf("failed to make txn: %s", err)
	}
	// sign the transaction
	txid, stx, err := crypto.SignTransaction(caller.PrivateKey, txn)
	if err != nil {
		log.Fatalf("failed to sign transaction: %s", err)
	}

	// Broadcast the transaction to the network
	_, err = algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		log.Fatalf("failed to send transaction: %s", err)
	}

	// Wait for confirmation
	confirmedTxn, err := transaction.WaitForConfirmation(algodClient, txid, 4, context.Background())
	if err != nil {
		log.Fatalf("error waiting for confirmation:  %s", err)
	}

	log.Printf("OptIn Transaction: %s confirmed in Round %d\n", txid, confirmedTxn.ConfirmedRound)
	// example: APP_OPTIN
}

func AppNoOp(algodClient *algod.Client, appID uint64, caller crypto.Account) {
	// example: APP_NOOP
	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		log.Fatalf("error getting suggested tx params: %s", err)
	}

	var (
		appArgs [][]byte
		accts   []string
		apps    []uint64
		assets  []uint64
	)

	// Add an arg to our app call
	appArgs = append(appArgs, []byte("arg0"))

	txn, err := transaction.MakeApplicationNoOpTx(
		appID, appArgs, accts, apps, assets, sp,
		caller.Address, nil, types.Digest{}, [32]byte{}, types.ZeroAddress,
	)
	if err != nil {
		log.Fatalf("failed to make txn: %s", err)
	}

	// sign the transaction
	txid, stx, err := crypto.SignTransaction(caller.PrivateKey, txn)
	if err != nil {
		log.Fatalf("failed to sign transaction: %s", err)
	}

	// Broadcast the transaction to the network
	_, err = algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		log.Fatalf("failed to send transaction: %s", err)
	}

	// Wait for confirmation
	confirmedTxn, err := transaction.WaitForConfirmation(algodClient, txid, 4, context.Background())
	if err != nil {
		log.Fatalf("error waiting for confirmation:  %s", err)
	}

	log.Printf("NoOp Transaction: %s confirmed in Round %d\n", txid, confirmedTxn.ConfirmedRound)
	// example: APP_NOOP
}

func AppUpdate(algodClient *algod.Client, appID uint64, caller crypto.Account) {
	approvalBinary := examples.CompileTeal(algodClient, "application/approval_refactored.teal")
	clearBinary := examples.CompileTeal(algodClient, "application/clear.teal")

	// example: APP_UPDATE
	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		log.Fatalf("error getting suggested tx params: %s", err)
	}

	var (
		appArgs [][]byte
		accts   []string
		apps    []uint64
		assets  []uint64
	)

	txn, err := transaction.MakeApplicationUpdateTx(
		appID, appArgs, accts, apps, assets, approvalBinary, clearBinary,
		sp, caller.Address, nil, types.Digest{}, [32]byte{}, types.ZeroAddress,
	)
	if err != nil {
		log.Fatalf("failed to make txn: %s", err)
	}

	// sign the transaction
	txid, stx, err := crypto.SignTransaction(caller.PrivateKey, txn)
	if err != nil {
		log.Fatalf("failed to sign transaction: %s", err)
	}

	// Broadcast the transaction to the network
	_, err = algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		log.Fatalf("failed to send transaction: %s", err)
	}

	// Wait for confirmation
	confirmedTxn, err := transaction.WaitForConfirmation(algodClient, txid, 4, context.Background())
	if err != nil {
		log.Fatalf("error waiting for confirmation:  %s", err)
	}

	log.Printf("Update Transaction: %s confirmed in Round %d\n", txid, confirmedTxn.ConfirmedRound)
	// example: APP_UPDATE
}

func AppCloseOut(algodClient *algod.Client, appID uint64, caller crypto.Account) {
	// example: APP_CLOSEOUT
	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		log.Fatalf("error getting suggested tx params: %s", err)
	}

	var (
		appArgs [][]byte
		accts   []string
		apps    []uint64
		assets  []uint64
	)

	txn, err := transaction.MakeApplicationCloseOutTx(
		appID, appArgs, accts, apps, assets, sp,
		caller.Address, nil, types.Digest{}, [32]byte{}, types.ZeroAddress,
	)
	if err != nil {
		log.Fatalf("failed to make txn: %s", err)
	}

	// sign the transaction
	txid, stx, err := crypto.SignTransaction(caller.PrivateKey, txn)
	if err != nil {
		log.Fatalf("failed to sign transaction: %s", err)
	}

	// Broadcast the transaction to the network
	_, err = algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		log.Fatalf("failed to send transaction: %s", err)
	}

	// Wait for confirmation
	confirmedTxn, err := transaction.WaitForConfirmation(algodClient, txid, 4, context.Background())
	if err != nil {
		log.Fatalf("error waiting for confirmation:  %s", err)
	}

	log.Printf("Closeout Transaction: %s confirmed in Round %d\n", txid, confirmedTxn.ConfirmedRound)
	// example: APP_CLOSEOUT
}

func AppClearState(algodClient *algod.Client, appID uint64, caller crypto.Account) {
	// example: APP_CLEAR
	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		log.Fatalf("error getting suggested tx params: %s", err)
	}

	var (
		appArgs [][]byte
		accts   []string
		apps    []uint64
		assets  []uint64
	)

	txn, err := transaction.MakeApplicationClearStateTx(
		appID, appArgs, accts, apps, assets, sp,
		caller.Address, nil, types.Digest{}, [32]byte{}, types.ZeroAddress,
	)
	if err != nil {
		log.Fatalf("failed to make txn: %s", err)
	}

	// sign the transaction
	txid, stx, err := crypto.SignTransaction(caller.PrivateKey, txn)
	if err != nil {
		log.Fatalf("failed to sign transaction: %s", err)
	}

	// Broadcast the transaction to the network
	_, err = algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		log.Fatalf("failed to send transaction: %s", err)
	}

	// Wait for confirmation
	confirmedTxn, err := transaction.WaitForConfirmation(algodClient, txid, 4, context.Background())
	if err != nil {
		log.Fatalf("error waiting for confirmation:  %s", err)
	}

	log.Printf("ClearState Transaction: %s confirmed in Round %d\n", txid, confirmedTxn.ConfirmedRound)
	// example: APP_CLEAR
}

func AppCall(algodClient *algod.Client, appID uint64, caller crypto.Account) {
	// example: APP_CALL
	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		log.Fatalf("error getting suggested tx params: %s", err)
	}

	var (
		accts  []string
		apps   []uint64
		assets []uint64
	)

	appArgs := [][]byte{
		[]byte("hello"), // Method (por ejemplo, el string "hello")
		[]byte("world"), // Argumento de entrada (por ejemplo, el string "world")
	}

	txn, err := transaction.MakeApplicationNoOpTx(
		appID, appArgs, accts, apps, assets, sp,
		caller.Address, nil, types.Digest{}, [32]byte{}, types.ZeroAddress,
	)
	if err != nil {
		log.Fatalf("failed to make txn: %s", err)
	}

	// sign the transaction
	txid, stx, err := crypto.SignTransaction(caller.PrivateKey, txn)
	if err != nil {
		log.Fatalf("failed to sign transaction: %s", err)
	}

	// Broadcast the transaction to the network
	_, err = algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		log.Fatalf("failed to send transaction: %s", err)
	}

	// Wait for confirmation
	confirmedTxn, err := transaction.WaitForConfirmation(algodClient, txid, 4, context.Background())
	if err != nil {
		log.Fatalf("error waiting for confirmation:  %s", err)
	}

	log.Printf("NoOp Transaction: %s confirmed in Round %d\n", txid, confirmedTxn.ConfirmedRound)
	// example: APP_CALL
}

func AppDelete(algodClient *algod.Client, appID uint64, caller crypto.Account) {
	// example: APP_DELETE
	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		log.Fatalf("error getting suggested tx params: %s", err)
	}

	var (
		appArgs [][]byte
		accts   []string
		apps    []uint64
		assets  []uint64
	)

	txn, err := transaction.MakeApplicationDeleteTx(
		appID, appArgs, accts, apps, assets, sp,
		caller.Address, nil, types.Digest{}, [32]byte{}, types.ZeroAddress,
	)
	if err != nil {
		log.Fatalf("failed to make txn: %s", err)
	}

	// sign the transaction
	txid, stx, err := crypto.SignTransaction(caller.PrivateKey, txn)
	if err != nil {
		log.Fatalf("failed to sign transaction: %s", err)
	}

	// Broadcast the transaction to the network
	_, err = algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		log.Fatalf("failed to send transaction: %s", err)
	}

	// Wait for confirmation
	confirmedTxn, err := transaction.WaitForConfirmation(algodClient, txid, 4, context.Background())
	if err != nil {
		log.Fatalf("error waiting for confirmation:  %s", err)
	}

	log.Printf("Delete Transaction: %s confirmed in Round %d\n", txid, confirmedTxn.ConfirmedRound)
	// example: APP_DELETE
}
