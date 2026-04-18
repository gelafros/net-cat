package main

import (
	"errors"
	"os"
)

func main() {
	args := os.Args[1:]
	getPort(args)
}

func getPort(args []string) (string, error) {
	if len(args) == 1 {
		return "8989", nil
	}

	if len(args) == 2 {
		return args[1], nil
	}

	return "", errors.New("[USAGE]: ./TCPChat $port")
}
