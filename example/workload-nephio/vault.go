/*
Copyright 2023 The Nephio Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	vault "github.com/hashicorp/vault/api"
)

type LoginPayload struct {
	Role string `json:"role"`
	JWT  string `json:"jwt"`
}

type AuthResponse struct {
	Auth struct {
		ClientToken string `json:"client_token"`
	} `json:"auth"`
}

func AuthenticateToVault(vaultAddr, jwt, role string) (*vault.Client, string, error) {
	// Create a Vault client
	config := vault.DefaultConfig()
	config.Address = vaultAddr
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, "", fmt.Errorf("unable to create Vault client: %w", err)
	}

	// fmt.Println("VAULT PRINT JWT: ", jwt)

	payload := LoginPayload{
		Role: role,
		JWT:  jwt,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("unable to marshal payload: %w", err)
	}

	// Perform the login request
	req := client.NewRequest("POST", "/v1/auth/jwt/login")
	req.Body = bytes.NewBuffer(payloadBytes)

	// fmt.Println("VAULT PRINT REQ:", req)
	resp, err := client.RawRequest(req)
	if err != nil {
		return nil, "", fmt.Errorf("unable to perform login request: %w", err)
	}
	defer resp.Body.Close()

	// fmt.Println("VAULT PRINT: ", resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("unable to read response body: %w", err)
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, "", fmt.Errorf("unable to decode response: %w", err)
	}

	return client, authResp.Auth.ClientToken, err
}

func FetchSecret(client *vault.Client, secretPath string) (string, error) {
	// Read the secret
	secret, err := client.Logical().Read(secretPath)
	if err != nil {
		return "", fmt.Errorf("unable to read secret: %w", err)
	}

	if secret == nil {
		return "", fmt.Errorf("secret not found at path: %s", secretPath)
	}

	// Extract the Kubeconfig data
	secretData, ok := secret.Data["test"].(string)
	if !ok {
		return "", fmt.Errorf("unable to get data associated with secret")
	}

	return secretData, nil
}
