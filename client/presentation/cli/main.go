package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/maksroxx/DeliveryService/client/application"
	"github.com/maksroxx/DeliveryService/client/domain"
	"github.com/maksroxx/DeliveryService/client/infrastructure"
)

func main() {
	var (
		client   = infrastructure.NewHTTPDeliveryClient("http://localhost:8228")
		createUC = application.NewCreatePackageUseCase(client)
		statusUC = application.NewGetStatusUseCase(client)
	)

	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "create":
		handleCreate(createUC)
	case "status":
		handleStatus(statusUC)
	default:
		printHelp()
	}
}

func handleCreate(uc *application.CreatePackageUseCase) {
	if len(os.Args) != 6 {
		fmt.Println("Use: create <weight> <from> <to> <when> <address>")
		return
	}

	weight, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		fmt.Println("Weight error:", err)
		return
	}

	req := domain.PackageRequest{
		Weight:  weight,
		From:    os.Args[3],
		To:      os.Args[4],
		Address: os.Args[5],
	}

	if err := uc.Execute(req); err != nil {
		fmt.Println("Error:", err)
	}
}

func handleStatus(uc *application.GetStatusUseCase) {
	if len(os.Args) != 3 {
		fmt.Println("Use: status <id>")
		return
	}

	if err := uc.Execute(os.Args[2]); err != nil {
		fmt.Println("Error:", err)
	}
}

func printHelp() {
	fmt.Println(`Commands:
	create <weight> <from> <to> <address> - Create package
	status <id> - Check status`)
}
