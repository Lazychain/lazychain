package utils

import (
	testifysuite "github.com/stretchr/testify/suite"
)

// Not const because we need to give them as pointers later
var (
	lazyVals           = 1
	lazyFullNodes      = 0
	celestialVals      = 1
	celestialFullNodes = 0
	stargazeVals       = 1
	stargazeFullNodes  = 0

	votingPeriod     = "15s"
	maxDepositPeriod = "10s"

	lazyChainId     = "lazytestchain-1"
	celestiaChainID = "celestiatest-1"
	sgChainID       = "stargazetest-1"
)

type E2ETestSuite struct {
	testifysuite.Suite
	InterchainValues
}

func (s *E2ETestSuite) SetupSuite() {
	s.testifySuiteRef = &s.Suite // See comment on the InterchainValues struct
	s.SetupInterchainValues()
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	if s.Interchain != nil {
		_ = s.Interchain.Close()
	}
}
