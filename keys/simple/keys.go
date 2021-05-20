// Package simple provides straightforward implementation for key management.
package simple

import (
	"context"
	"fmt"

	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/eqlabs/flow-wallet-service/keys/encryption"
	"github.com/eqlabs/flow-wallet-service/keys/google"
	"github.com/eqlabs/flow-wallet-service/keys/local"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
)

type KeyManager struct {
	db              keys.Store
	fc              *client.Client
	crypter         encryption.Crypter
	adminAccountKey keys.Private
	cfg             Config
}

// NewKeyManager initiates a new key manager.
// It uses encryption.AESCrypter to encrypt and decrypt the keys.
func NewKeyManager(db keys.Store, fc *client.Client) *KeyManager {
	cfg := ParseConfig()

	adminAccountKey := keys.Private{
		Index:    cfg.AdminAccountKeyIndex,
		Type:     cfg.AdminAccountKeyType,
		Value:    cfg.AdminAccountKeyValue,
		SignAlgo: crypto.StringToSignatureAlgorithm(cfg.DefaultSignAlgo),
		HashAlgo: crypto.StringToHashAlgorithm(cfg.DefaultHashAlgo),
	}

	crypter := encryption.NewAESCrypter([]byte(cfg.EncryptionKey))

	return &KeyManager{
		db,
		fc,
		crypter,
		adminAccountKey,
		cfg,
	}
}

func (s *KeyManager) Generate(ctx context.Context, keyIndex, weight int) (keys.Wrapped, error) {
	switch s.cfg.DefaultKeyType {
	case keys.ACCOUNT_KEY_TYPE_LOCAL:
		return local.Generate(
			keyIndex, weight,
			crypto.StringToSignatureAlgorithm(s.cfg.DefaultSignAlgo),
			crypto.StringToHashAlgorithm(s.cfg.DefaultHashAlgo))
	case keys.ACCOUNT_KEY_TYPE_GOOGLE_KMS:
		return google.Generate(ctx, keyIndex, weight)
	default:
		return keys.Wrapped{}, fmt.Errorf("keyStore.Generate() not implmented for %s", s.cfg.DefaultKeyType)
	}
}

func (s *KeyManager) GenerateDefault(ctx context.Context) (keys.Wrapped, error) {
	return s.Generate(ctx, s.cfg.DefaultKeyIndex, s.cfg.DefaultKeyWeight)
}

func (s *KeyManager) Save(key keys.Private) (keys.Storable, error) {
	encValue, err := s.crypter.Encrypt([]byte(key.Value))
	if err != nil {
		return keys.Storable{}, err
	}
	return keys.Storable{
		Index:    key.Index,
		Type:     key.Type,
		Value:    encValue,
		SignAlgo: key.SignAlgo.String(),
		HashAlgo: key.HashAlgo.String(),
	}, nil
}

func (s *KeyManager) Load(key keys.Storable) (keys.Private, error) {
	decValue, err := s.crypter.Decrypt([]byte(key.Value))
	if err != nil {
		return keys.Private{}, err
	}
	return keys.Private{
		Index:    key.Index,
		Type:     key.Type,
		Value:    string(decValue),
		SignAlgo: crypto.StringToSignatureAlgorithm(key.SignAlgo),
		HashAlgo: crypto.StringToHashAlgorithm(key.HashAlgo),
	}, nil
}

func (s *KeyManager) AdminAuthorizer(ctx context.Context) (keys.Authorizer, error) {
	return s.MakeAuthorizer(ctx, s.cfg.AdminAccountAddress)
}

func (s *KeyManager) UserAuthorizer(ctx context.Context, address string) (keys.Authorizer, error) {
	return s.MakeAuthorizer(ctx, address)
}

func (s *KeyManager) MakeAuthorizer(ctx context.Context, address string) (keys.Authorizer, error) {
	var k keys.Private

	if address == s.cfg.AdminAccountAddress {
		k = s.adminAccountKey
	} else {
		// Get the "least recently used" key for this address
		sk, err := s.db.AccountKey(address)
		if err != nil {
			return keys.Authorizer{}, err
		}
		k, err = s.Load(sk)
		if err != nil {
			return keys.Authorizer{}, err
		}
	}

	acc, err := s.fc.GetAccount(ctx, flow.HexToAddress(address))
	if err != nil {
		return keys.Authorizer{}, err
	}

	var sig crypto.Signer

	// TODO: Decide whether we want to allow this kind of flexibility
	// or should we just panic if `key.Type` != `s.defaultKeyManager`
	switch k.Type {
	case keys.ACCOUNT_KEY_TYPE_LOCAL:
		sig, err = local.Signer(k)
		if err != nil {
			return keys.Authorizer{}, err
		}
	case keys.ACCOUNT_KEY_TYPE_GOOGLE_KMS:
		sig, err = google.Signer(ctx, address, k)
		if err != nil {
			return keys.Authorizer{}, err
		}
	default:
		return keys.Authorizer{}, fmt.Errorf("key.Type not recognised: %s", k.Type)
	}

	return keys.Authorizer{
		Address: flow.HexToAddress(address),
		Key:     acc.Keys[k.Index],
		Signer:  sig,
	}, nil
}
