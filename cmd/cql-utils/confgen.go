/*
 * Copyright 2018 The CovenantSQL Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"

	"github.com/CovenantSQL/CovenantSQL/conf/testnet"
	"github.com/CovenantSQL/CovenantSQL/proto"
	"github.com/CovenantSQL/CovenantSQL/utils"
	"github.com/CovenantSQL/CovenantSQL/utils/log"
)

var (
	workingRoot string
)

func init() {
	flag.StringVar(&workingRoot, "root", "~/.cql", "confgen root is the working root directory containing all auto-generating keys and certifications")
}

func runConfgen() {
	workingRoot = utils.HomeDirExpand(workingRoot)
	if workingRoot == "" {
		log.Error("root directory is required for confgen")
		os.Exit(1)
	}

	privateKeyFileName := "private.key"
	publicKeystoreFileName := "public.keystore"
	privateKeyFile = path.Join(workingRoot, privateKeyFileName)

	askDeletePath(workingRoot)

	err := os.Mkdir(workingRoot, 0755)
	if err != nil {
		log.WithError(err).Error("unexpected error")
		os.Exit(1)
	}

	fmt.Println("Generating key pair...")
	publicKey := runKeygen(privateKeyFile)
	fmt.Println("Generated key pair.")

	fmt.Println("Generating nonce...")
	nonce := noncegen(publicKey)
	cliNodeID := proto.NodeID(nonce.Hash.String())
	fmt.Println("Generated nonce.")

	fmt.Println("Generating config file...")
	// Load testnet config
	testnetConfig := testnet.GetTestNetConfig()
	// Add client config
	testnetConfig.PrivateKeyFile = privateKeyFileName
	testnetConfig.PubKeyStoreFile = publicKeystoreFileName
	testnetConfig.ThisNodeID = cliNodeID
	if testnetConfig.KnownNodes == nil {
		testnetConfig.KnownNodes = make([]proto.Node, 1)
	}
	testnetConfig.KnownNodes[0] = proto.Node{
		ID:        cliNodeID,
		Role:      proto.Client,
		Addr:      "0.0.0.0:15151",
		PublicKey: publicKey,
		Nonce:     nonce.Nonce,
	}

	// Write config
	out, err := yaml.Marshal(testnetConfig)
	if err != nil {
		log.WithError(err).Error("unexpected error")
		os.Exit(1)
	}
	err = ioutil.WriteFile(path.Join(workingRoot, "config.yaml"), out, 0644)
	if err != nil {
		log.WithError(err).Error("unexpected error")
		os.Exit(1)
	}
	fmt.Println("Generated nonce.")
}
