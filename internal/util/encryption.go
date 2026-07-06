// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"

	"github.com/heavyai/webserver/internal/config"
	"github.com/heavyai/webserver/models"
)

// DecryptString - Decrypts cipher text string into plain text string
func DecryptString(encrypted string) (decrypted string, err error) {
	keyBytes := []byte(config.Other.EncryptedCredentialsKey)
	encryptedBytes, _ := hex.DecodeString(encrypted)

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		panic(err.Error())
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	// separate nonce from cypher
	nonceSize := aesGCM.NonceSize()
	nonce, cipherBytes := encryptedBytes[:nonceSize], encryptedBytes[nonceSize:]
	//Decrypt the data
	decryptedBytes, err := aesGCM.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		panic(err.Error())
	}

	return fmt.Sprintf("%s", decryptedBytes), nil
}

// GetDecryptedCreds - Decrypts authentication credentials
func GetDecryptedCreds(creds *models.AuthCredentials) (err error) {
	creds.Username, err = DecryptString(creds.Username)
	if err != nil {
		return
	}

	creds.Password, err = DecryptString(creds.Password)
	if err != nil {
		return
	}

	creds.DBName, err = DecryptString(creds.DBName)
	if err != nil {
		return
	}

	return
}
