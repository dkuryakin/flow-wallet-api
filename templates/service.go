package templates

import (
	"fmt"
	"strings"

	"github.com/eqlabs/flow-wallet-api/flow_helpers"
	"github.com/onflow/flow-go-sdk"
)

type Service struct {
	store Store
	cfg   *config
}

func NewService(store Store) *Service {
	cfg := parseConfig()
	// Add all enabled tokens from config as fungible tokens
	for _, t := range cfg.enabledTokens {
		if _, err := store.GetByName(t.Name); err == nil {
			// Token already in database
			continue
		} else {
			if !strings.Contains(err.Error(), "record not found") {
				// We got an error that is not "record not found"
				panic(err)
			}
		}

		t.Type = FT
		t.Setup = FungibleSetupCode(&t)
		t.Transfer = FungibleTransferCode(&t)
		t.Balance = FungibleBalanceCode(&t)

		if err := store.Insert(&t); err != nil {
			panic(err)
		}
	}
	return &Service{store, cfg}
}

func (s *Service) AddToken(t *Token) error {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(t.Address, s.cfg.ChainId)
	if err != nil {
		return err
	}

	t.Address = address

	if t.Name == "" {
		return fmt.Errorf(`not a valid name: "%s"`, t.Name)
	}

	// Received code templates may have values that need replacing
	t.Setup = TokenCode(t, t.Setup)
	t.Transfer = TokenCode(t, t.Transfer)
	t.Balance = TokenCode(t, t.Balance)

	return s.store.Insert(t)
}

func (s *Service) ListTokens(tType *TokenType) (*[]BasicToken, error) {
	return s.store.List(tType)
}

func (s *Service) GetTokenById(id uint64) (*Token, error) {
	return s.store.GetById(id)
}

func (s *Service) GetTokenByName(name string) (*Token, error) {
	return s.store.GetByName(name)
}

func (s *Service) RemoveToken(id uint64) error {
	return s.store.Remove(id)
}

func (s *Service) TokenFromEvent(e flow.Event) (*Token, error) {
	// Example event:
	// A.0ae53cb6e3f42a79.FlowToken.TokensDeposited
	ss := strings.Split(e.Type, ".")

	// Token address from event
	eAddress, err := flow_helpers.ValidateAddress(ss[1], s.cfg.ChainId)
	if err != nil {
		return nil, err
	}

	token, err := s.GetTokenByName(ss[2])
	if err != nil {
		return nil, err
	}

	// Token address from database
	tAddress, err := flow_helpers.ValidateAddress(token.Address, s.cfg.ChainId)
	if err != nil {
		return nil, err
	}

	// Check if addresses match
	if eAddress != tAddress {
		return nil, fmt.Errorf("addresses do not match for %s, from event: %s, from database: %s", token.Name, eAddress, tAddress)
	}

	return token, nil
}