package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"log"
	"os"
	"strings"
	"test/application"
)

// loadAccount carga la cuenta a partir de un archivo que contiene la dirección y el mnemónico.
func loadAccount(filePath string) (crypto.Account, error) {
	// Abrir el archivo para lectura
	file, err := os.Open(filePath)
	if err != nil {
		return crypto.Account{}, fmt.Errorf("error al abrir el archivo: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var address string
	var mn string

	// Leer el archivo línea por línea
	for scanner.Scan() {
		line := scanner.Text()

		// Verificar si la línea contiene la dirección o el mnemónico
		if strings.HasPrefix(line, "Dirección:") {
			address = strings.TrimSpace(strings.TrimPrefix(line, "Dirección:"))
		} else if strings.HasPrefix(line, "Mnemónico:") {
			mn = strings.TrimSpace(strings.TrimPrefix(line, "Mnemónico:"))
		}
	}

	// Verificar si leímos ambos valores
	if address == "" || mn == "" {
		return crypto.Account{}, fmt.Errorf("no se encontraron todos los datos necesarios en el archivo")
	}

	// Restaurar la clave privada a partir del mnemónico
	privateKey, err := mnemonic.ToPrivateKey(mn)
	if err != nil {
		return crypto.Account{}, fmt.Errorf("error al convertir el mnemónico a clave privada: %s", err)
	}

	// Generar la dirección a partir de la clave privada
	restoredAddress, err := crypto.GenerateAddressFromSK(privateKey)
	if err != nil {
		return crypto.Account{}, fmt.Errorf("error al generar la dirección desde la clave privada: %s", err)
	}

	// Verificar que la dirección restaurada coincida con la dirección guardada
	if restoredAddress.String() != address {
		return crypto.Account{}, fmt.Errorf("la dirección restaurada no coincide con la guardada")
	}

	// Crear la cuenta con la clave privada y la dirección restaurada
	account := crypto.Account{
		PrivateKey: privateKey,
		Address:    restoredAddress,
	}

	return account, nil
}

func main() {

	filePath := "algorand_account.txt"

	// Restaurar la cuenta desde el archivo
	account, err := loadAccount(filePath)
	if err != nil {
		log.Fatalf("Error al cargar la cuenta: %s", err)
	}

	fmt.Println(account.Address.String())

	var algodAddress = "https://testnet-api.algonode.cloud"
	var algodToken = strings.Repeat("a", 64)
	algodClient, _ := algod.MakeClient(
		algodAddress,
		algodToken,
	)

	appID := uint64(721920807) //Hello world
	//appID = application.AppCreate(algodClient, account)
	////application.AppOptIn(algodClient, appID, account)
	//
	//// example: APP_READ_STATE
	//// grab global state and config of application
	appInfo, err := algodClient.GetApplicationByID(appID).Do(context.Background())
	if err != nil {
		log.Fatalf("failed to get app info: %s", err)
	}
	log.Printf("app info: %+v", appInfo)
	//
	//// grab local state for an app id for a single account
	//acctInfo, err := algodClient.AccountApplicationInformation(
	//	account.Address.String(), appID,
	//).Do(context.Background())
	//if err != nil {
	//	log.Fatalf("failed to get app info: %s", err)
	//}
	//log.Printf("app info: %+v", acctInfo)
	//// example: APP_READ_STATE
	//
	////application.AppNoOp(algodClient, appID, account)
	////application.AppUpdate(algodClient, appID, account)
	application.AppCall(algodClient, appID, account)
	//application.AppCloseOut(algodClient, appID, account)
	//application.AppDelete(algodClient, appID, account)

}
