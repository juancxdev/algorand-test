package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/algorand/go-algorand-sdk/v2/abi"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/transaction"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"io/ioutil"
	"log"
	"os"
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

func AppCall(algodClient *algod.Client, appID uint64, caller crypto.Account) {
	b, err := os.ReadFile("./Bjaguar.arc32.json")
	if err != nil {
		log.Fatalf("Failed to open contract file: %+v", err)
	}

	// Crear un mapa vacío para almacenar el JSON parseado
	var contractJSON map[string]interface{}

	// Parsear el JSON al mapa
	err = json.Unmarshal(b, &contractJSON)
	if err != nil {
		fmt.Println("Error al parsear el JSON:", err)
		return
	}

	contractData, ok := contractJSON["contract"]
	if !ok {
		log.Fatal("El campo 'contract' no fue encontrado en el JSON")
	}

	// Convertir el valor de 'contract' nuevamente a JSON (para poder deserializarlo)
	contractBytes, err := json.Marshal(contractData)
	if err != nil {
		log.Fatalf("Error al serializar 'contract' nuevamente: %+v", err)
	}

	contract := &abi.Contract{}
	if err := json.Unmarshal(contractBytes, contract); err != nil {
		log.Fatalf("Failed to marshal contract: %+v", err)
	}

	//app_addr := crypto.GetApplicationAddress(appID)

	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		log.Fatalf("Failed to get suggeted params: %+v", err)
	}

	// Create a signer and some common parameters
	signer := transaction.BasicAccountTransactionSigner{Account: caller}
	mcp := transaction.AddMethodCallParams{
		AppID:           appID,
		Sender:          caller.Address,
		SuggestedParams: sp,
		OnComplete:      types.NoOpOC,
		Signer:          signer,
	}

	var atc = transaction.AtomicTransactionComposer{}

	err = atc.AddMethodCall(combine(mcp, getMethod(contract, "hello"), []interface{}{"success"}))

	if err != nil {
		log.Fatalf("Failed to add method call: %+v", err)
	}

	ret, err := atc.Execute(algodClient, context.Background(), 2)
	if err != nil {
		log.Fatalf("Failed to execute call: %+v", err)
	}

	for _, r := range ret.MethodResults {
		log.Printf("%+v", r.TransactionInfo.Logs)
		log.Printf("%s returned %+v", r.Method.Name, r.ReturnValue)
	}
}

func getMethod(c *abi.Contract, name string) abi.Method {
	m, err := c.GetMethodByName(name)
	if err != nil {
		log.Fatalf("No method named: %s", name)
	}
	return m
}

func combine(mcp transaction.AddMethodCallParams, m abi.Method, a []interface{}, boxes ...types.AppBoxReference) transaction.AddMethodCallParams {
	mcp.Method = m
	mcp.MethodArgs = a
	mcp.BoxReferences = boxes
	return mcp
}
