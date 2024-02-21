package utils

import (
	"log"
	"os"
)

var (
	key []byte
)

func init() {
	key_str, ok := os.LookupEnv("AES_KEY")
	if !ok {
		log.Fatal("No AES key provided. Expected AES_KEY")
	}
	key = []byte(key_str)
}

func Encrypt(data []byte) ([]byte, error) {
	return []byte("TODO"), nil
}

func Decrypt(data []byte) ([]byte, error) {
	return []byte("TODO"), nil
}
