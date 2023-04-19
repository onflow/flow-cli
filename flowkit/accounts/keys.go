/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package accounts

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	goeth "github.com/ethereum/go-ethereum/accounts"
	"github.com/lmars/go-slip10"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"
	"github.com/tyler-smith/go-bip39"

	"github.com/onflow/flow-cli/flowkit/config"
)

// AccountPublicKey contains public account key information.
type AccountPublicKey struct {
	Public   crypto.PublicKey
	Weight   int
	SigAlgo  crypto.SignatureAlgorithm
	HashAlgo crypto.HashAlgorithm
}

// Key defines functions any key representation must implement.
type Key interface {
	// Type returns the key type (hex, kms, file...)
	Type() config.KeyType
	// Index returns the key index on the account
	Index() int
	// SigAlgo returns signature algorithm used for signing
	SigAlgo() crypto.SignatureAlgorithm
	// HashAlgo returns hash algorithm used for signing
	HashAlgo() crypto.HashAlgorithm
	// Signer is used when we want to sign a transaction, using the Sign() function
	Signer(ctx context.Context) (crypto.Signer, error)
	// ToConfig converts the key to the storable key format
	ToConfig() config.AccountKey
	// Validate key
	Validate() error
	// PrivateKey returns the private key if possible,
	// depends on the key type
	PrivateKey() (*crypto.PrivateKey, error)
}

var _ Key = &HexKey{}

var _ Key = &KMSKey{}

var _ Key = &BIP44Key{}

func keyFromConfig(accountKeyConf config.AccountKey) (Key, error) {
	switch accountKeyConf.Type {
	case config.KeyTypeHex:
		return hexKeyFromConfig(accountKeyConf)
	case config.KeyTypeBip44:
		return bip44KeyFromConfig(accountKeyConf)
	case config.KeyTypeGoogleKMS:
		return kmsKeyFromConfig(accountKeyConf)
	case config.KeyTypeFile:
		return fileKeyFromConfig(accountKeyConf)
	}

	return nil, fmt.Errorf(`invalid key type: "%s"`, accountKeyConf.Type)
}

type baseKey struct {
	keyType  config.KeyType
	index    int
	sigAlgo  crypto.SignatureAlgorithm
	hashAlgo crypto.HashAlgorithm
}

func baseKeyFromConfig(accountKeyConf config.AccountKey) *baseKey {
	return &baseKey{
		keyType:  accountKeyConf.Type,
		index:    accountKeyConf.Index,
		sigAlgo:  accountKeyConf.SigAlgo,
		hashAlgo: accountKeyConf.HashAlgo,
	}
}

func (a *baseKey) Type() config.KeyType {
	return a.keyType
}

func (a *baseKey) SigAlgo() crypto.SignatureAlgorithm {
	if a.sigAlgo == crypto.UnknownSignatureAlgorithm {
		return crypto.ECDSA_P256 // default value
	}
	return a.sigAlgo
}

func (a *baseKey) HashAlgo() crypto.HashAlgorithm {
	if a.hashAlgo == crypto.UnknownHashAlgorithm {
		return crypto.SHA3_256 // default value
	}
	return a.hashAlgo
}

func (a *baseKey) Index() int {
	return a.index // default to 0
}

func (a *baseKey) Validate() error {
	return nil
}

// KMSKey implements Gcloud KMS system for signing.
type KMSKey struct {
	*baseKey
	kmsKey cloudkms.Key
}

// ToConfig convert account key to configuration.
func (a *KMSKey) ToConfig() config.AccountKey {
	return config.AccountKey{
		Type:       a.keyType,
		Index:      a.index,
		SigAlgo:    a.sigAlgo,
		HashAlgo:   a.hashAlgo,
		ResourceID: a.kmsKey.ResourceID(),
	}
}

func (a *KMSKey) Signer(ctx context.Context) (crypto.Signer, error) {
	kmsClient, err := cloudkms.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	accountKMSSigner, err := kmsClient.SignerForKey(
		ctx,
		a.kmsKey,
	)
	if err != nil {
		return nil, err
	}

	return accountKMSSigner, nil
}

func (a *KMSKey) Validate() error {
	return gcloudApplicationSignin(a.kmsKey.ResourceID())
}

func (a *KMSKey) PrivateKey() (*crypto.PrivateKey, error) {
	return nil, fmt.Errorf("private key not accessible")
}

// gcloudApplicationSignin signs in as an application user using gcloud command line tool
// currently assumes gcloud is already installed on the machine
// will by default pop a browser window to sign in
func gcloudApplicationSignin(resourceID string) error {
	googleAppCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if len(googleAppCreds) > 0 {
		return nil
	}

	kms, err := cloudkms.KeyFromResourceID(resourceID)
	if err != nil {
		return err
	}

	proj := kms.ProjectID
	if len(proj) == 0 {
		return fmt.Errorf(
			"could not get GOOGLE_APPLICATION_CREDENTIALS, no google service account JSON provided but private key type is KMS",
		)
	}

	loginCmd := exec.Command("gcloud", "auth", "application-default", "login", fmt.Sprintf("--project=%s", proj))

	output, err := loginCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to run %q: %s\n", loginCmd.String(), err)
	}

	squareBracketRegex := regexp.MustCompile(`(?s)\[(.*)\]`)
	regexResult := squareBracketRegex.FindAllStringSubmatch(string(output), -1)
	// Should only be one value. Second index since first index contains the square brackets
	googleApplicationCreds := regexResult[0][1]

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", googleApplicationCreds)

	return nil
}

func kmsKeyFromConfig(key config.AccountKey) (Key, error) {
	accountKMSKey, err := cloudkms.KeyFromResourceID(key.ResourceID)
	if err != nil {
		return nil, err
	}

	return &KMSKey{
		baseKey: &baseKey{
			keyType:  config.KeyTypeGoogleKMS,
			index:    key.Index,
			sigAlgo:  key.SigAlgo,
			hashAlgo: key.HashAlgo,
		},
		kmsKey: accountKMSKey,
	}, nil
}

// HexKey implements account key in hex representation.
type HexKey struct {
	*baseKey
	privateKey crypto.PrivateKey
}

func NewHexKeyFromPrivateKey(
	index int,
	hashAlgo crypto.HashAlgorithm,
	privateKey crypto.PrivateKey,
) *HexKey {
	return &HexKey{
		baseKey: &baseKey{
			keyType:  config.KeyTypeHex,
			index:    index,
			sigAlgo:  privateKey.Algorithm(),
			hashAlgo: hashAlgo,
		},
		privateKey: privateKey,
	}
}

func hexKeyFromConfig(accountKey config.AccountKey) (*HexKey, error) {
	return &HexKey{
		baseKey:    baseKeyFromConfig(accountKey),
		privateKey: accountKey.PrivateKey,
	}, nil
}

func (a *HexKey) Signer(ctx context.Context) (crypto.Signer, error) {
	return crypto.NewInMemorySigner(a.privateKey, a.HashAlgo())
}

func (a *HexKey) PrivateKey() (*crypto.PrivateKey, error) {
	return &a.privateKey, nil
}

func (a *HexKey) ToConfig() config.AccountKey {
	return config.AccountKey{
		Type:       a.keyType,
		Index:      a.index,
		SigAlgo:    a.sigAlgo,
		HashAlgo:   a.hashAlgo,
		PrivateKey: a.privateKey,
	}
}

func (a *HexKey) Validate() error {
	_, err := crypto.DecodePrivateKeyHex(a.sigAlgo, a.privateKeyHex())
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	return nil
}

func (a *HexKey) privateKeyHex() string {
	return hex.EncodeToString(a.privateKey.Encode())
}

// fileKeyFromConfig creates a hex account key from a file location
func fileKeyFromConfig(accountKey config.AccountKey) (*FileKey, error) {
	return &FileKey{
		baseKey:  baseKeyFromConfig(accountKey),
		location: accountKey.Location,
	}, nil
}

// NewFileKey creates a new account key that is stored to a separate file in the provided location.
//
// This type of the key is a more secure way of storing accounts. The config only includes the location and not the key.
func NewFileKey(
	location string,
	index int,
	sigAlgo crypto.SignatureAlgorithm,
	hashAlgo crypto.HashAlgorithm,
) *FileKey {
	return &FileKey{
		baseKey: &baseKey{
			keyType:  config.KeyTypeFile,
			index:    index,
			sigAlgo:  sigAlgo,
			hashAlgo: hashAlgo,
		},
		location: location,
	}
}

// FileKey represents a key that is saved in a seperate file and will be lazy-loaded.
//
// The FileKey stores location of the file where private key is stored in hex-encoded format.
type FileKey struct {
	*baseKey
	privateKey crypto.PrivateKey
	location   string
}

func (f *FileKey) Signer(ctx context.Context) (crypto.Signer, error) {
	key, err := f.PrivateKey()
	if err != nil {
		return nil, err
	}

	return crypto.NewInMemorySigner(*key, f.HashAlgo())
}

func (f *FileKey) PrivateKey() (*crypto.PrivateKey, error) {
	if f.privateKey == nil { // lazy load the key
		key, err := os.ReadFile(f.location) // TODO(sideninja) change to use the state ReaderWriter
		if err != nil {
			return nil, fmt.Errorf("could not load the key for the account from provided location %s: %w", f.location, err)
		}
		pkey, err := crypto.DecodePrivateKeyHex(f.sigAlgo, strings.TrimPrefix(string(key), "0x"))
		if err != nil {
			return nil, fmt.Errorf("could not decode the key from provided location %s: %w", f.location, err)
		}
		f.privateKey = pkey
	}
	return &f.privateKey, nil
}

func (f *FileKey) ToConfig() config.AccountKey {
	return config.AccountKey{
		Type:     config.KeyTypeFile,
		SigAlgo:  f.sigAlgo,
		HashAlgo: f.hashAlgo,
		Location: f.location,
	}
}

// BIP44Key implements https://github.com/onflow/flow/blob/master/flips/20201125-bip-44-multi-account.md
type BIP44Key struct {
	*baseKey
	privateKey     crypto.PrivateKey
	mnemonic       string
	derivationPath string
}

func bip44KeyFromConfig(key config.AccountKey) (Key, error) {
	return &BIP44Key{
		baseKey: &baseKey{
			keyType:  config.KeyTypeBip44,
			index:    key.Index,
			sigAlgo:  key.SigAlgo,
			hashAlgo: key.HashAlgo,
		},
		derivationPath: key.DerivationPath,
		mnemonic:       key.Mnemonic,
	}, nil
}

func (a *BIP44Key) Signer(ctx context.Context) (crypto.Signer, error) {
	pkey, err := a.PrivateKey()
	if err != nil {
		return nil, err
	}

	return crypto.NewInMemorySigner(*pkey, a.HashAlgo())
}

func (a *BIP44Key) PrivateKey() (*crypto.PrivateKey, error) {
	if a.privateKey == nil { // lazy load
		err := a.Validate()
		if err != nil {
			return nil, err
		}
	}
	return &a.privateKey, nil
}

func (a *BIP44Key) ToConfig() config.AccountKey {
	return config.AccountKey{
		Type:           a.keyType,
		Index:          a.index,
		SigAlgo:        a.sigAlgo,
		HashAlgo:       a.hashAlgo,
		PrivateKey:     a.privateKey,
		Mnemonic:       a.mnemonic,
		DerivationPath: a.derivationPath,
	}
}

func (a *BIP44Key) Validate() error {

	if !bip39.IsMnemonicValid(a.mnemonic) {
		return fmt.Errorf("invalid mnemonic defined for account in flow.json")
	}

	derivationPath, err := goeth.ParseDerivationPath(a.derivationPath)
	if err != nil {
		return fmt.Errorf("invalid derivation path defined for account in flow.json")
	}

	seed := bip39.NewSeed(a.mnemonic, "")
	curve := slip10.CurveBitcoin
	if a.sigAlgo == crypto.ECDSA_P256 {
		curve = slip10.CurveP256
	}
	accountKey, err := slip10.NewMasterKeyWithCurve(seed, curve)
	if err != nil {
		return err
	}

	for _, n := range derivationPath {
		accountKey, err = accountKey.NewChildKey(n)

		if err != nil {
			return err
		}
	}
	a.privateKey, err = crypto.DecodePrivateKey(a.SigAlgo(), accountKey.Key)
	if err != nil {
		return err
	}
	return nil
}
