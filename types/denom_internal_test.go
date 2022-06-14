package types

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

var (
	bnkt  = "bnkt"  // 1 (base denom unit)
	mbnkt = "mbnkt" // 10^-3 (milli)
	ubnkt = "ubnkt" // 10^-6 (micro)
	nbnkt = "nbnkt" // 10^-9 (nano)
)

type internalDenomTestSuite struct {
	suite.Suite
}

func TestInternalDenomTestSuite(t *testing.T) {
	suite.Run(t, new(internalDenomTestSuite))
}

func (s *internalDenomTestSuite) TestRegisterDenom() {
	bnktUnit := OneDec() // 1 (base denom unit)

	s.Require().NoError(RegisterDenom(bnkt, bnktUnit))
	s.Require().Error(RegisterDenom(bnkt, bnktUnit))

	res, ok := GetDenomUnit(bnkt)
	s.Require().True(ok)
	s.Require().Equal(bnktUnit, res)

	res, ok = GetDenomUnit(mbnkt)
	s.Require().False(ok)
	s.Require().Equal(ZeroDec(), res)

	// reset registration
	baseDenom = ""
	denomUnits = map[string]Dec{}
}

func (s *internalDenomTestSuite) TestConvertCoins() {
	bnktUnit := OneDec() // 1 (base denom unit)
	s.Require().NoError(RegisterDenom(bnkt, bnktUnit))

	mbnktUnit := NewDecWithPrec(1, 3) // 10^-3 (milli)
	s.Require().NoError(RegisterDenom(mbnkt, mbnktUnit))

	ubnktUnit := NewDecWithPrec(1, 6) // 10^-6 (micro)
	s.Require().NoError(RegisterDenom(ubnkt, ubnktUnit))

	nbnktUnit := NewDecWithPrec(1, 9) // 10^-9 (nano)
	s.Require().NoError(RegisterDenom(nbnkt, nbnktUnit))

	res, err := GetBaseDenom()
	s.Require().NoError(err)
	s.Require().Equal(res, nbnkt)
	s.Require().Equal(NormalizeCoin(NewCoin(ubnkt, NewInt(1))), NewCoin(nbnkt, NewInt(1000)))
	s.Require().Equal(NormalizeCoin(NewCoin(mbnkt, NewInt(1))), NewCoin(nbnkt, NewInt(1000000)))
	s.Require().Equal(NormalizeCoin(NewCoin(bnkt, NewInt(1))), NewCoin(nbnkt, NewInt(1000000000)))

	coins, err := ParseCoinsNormalized("1bnkt,1mbnkt,1ubnkt")
	s.Require().NoError(err)
	s.Require().Equal(coins, Coins{
		Coin{nbnkt, NewInt(1000000000)},
		Coin{nbnkt, NewInt(1000000)},
		Coin{nbnkt, NewInt(1000)},
	})

	testCases := []struct {
		input  Coin
		denom  string
		result Coin
		expErr bool
	}{
		{NewCoin("foo", ZeroInt()), bnkt, Coin{}, true},
		{NewCoin(bnkt, ZeroInt()), "foo", Coin{}, true},
		{NewCoin(bnkt, ZeroInt()), "FOO", Coin{}, true},

		{NewCoin(bnkt, NewInt(5)), mbnkt, NewCoin(mbnkt, NewInt(5000)), false},       // bnkt => mbnkt
		{NewCoin(bnkt, NewInt(5)), ubnkt, NewCoin(ubnkt, NewInt(5000000)), false},    // bnkt => ubnkt
		{NewCoin(bnkt, NewInt(5)), nbnkt, NewCoin(nbnkt, NewInt(5000000000)), false}, // bnkt => nbnkt

		{NewCoin(ubnkt, NewInt(5000000)), mbnkt, NewCoin(mbnkt, NewInt(5000)), false},       // ubnkt => mbnkt
		{NewCoin(ubnkt, NewInt(5000000)), nbnkt, NewCoin(nbnkt, NewInt(5000000000)), false}, // ubnkt => nbnkt
		{NewCoin(ubnkt, NewInt(5000000)), bnkt, NewCoin(bnkt, NewInt(5)), false},            // ubnkt => bnkt

		{NewCoin(mbnkt, NewInt(5000)), nbnkt, NewCoin(nbnkt, NewInt(5000000000)), false}, // mbnkt => nbnkt
		{NewCoin(mbnkt, NewInt(5000)), ubnkt, NewCoin(ubnkt, NewInt(5000000)), false},    // mbnkt => ubnkt
	}

	for i, tc := range testCases {
		res, err := ConvertCoin(tc.input, tc.denom)
		s.Require().Equal(
			tc.expErr, err != nil,
			"unexpected error; tc: #%d, input: %s, denom: %s", i+1, tc.input, tc.denom,
		)
		s.Require().Equal(
			tc.result, res,
			"invalid result; tc: #%d, input: %s, denom: %s", i+1, tc.input, tc.denom,
		)
	}

	// reset registration
	baseDenom = ""
	denomUnits = map[string]Dec{}
}

func (s *internalDenomTestSuite) TestConvertDecCoins() {
	bnktUnit := OneDec() // 1 (base denom unit)
	s.Require().NoError(RegisterDenom(bnkt, bnktUnit))

	mbnktUnit := NewDecWithPrec(1, 3) // 10^-3 (milli)
	s.Require().NoError(RegisterDenom(mbnkt, mbnktUnit))

	ubnktUnit := NewDecWithPrec(1, 6) // 10^-6 (micro)
	s.Require().NoError(RegisterDenom(ubnkt, ubnktUnit))

	nbnktUnit := NewDecWithPrec(1, 9) // 10^-9 (nano)
	s.Require().NoError(RegisterDenom(nbnkt, nbnktUnit))

	res, err := GetBaseDenom()
	s.Require().NoError(err)
	s.Require().Equal(res, nbnkt)
	s.Require().Equal(NormalizeDecCoin(NewDecCoin(ubnkt, NewInt(1))), NewDecCoin(nbnkt, NewInt(1000)))
	s.Require().Equal(NormalizeDecCoin(NewDecCoin(mbnkt, NewInt(1))), NewDecCoin(nbnkt, NewInt(1000000)))
	s.Require().Equal(NormalizeDecCoin(NewDecCoin(bnkt, NewInt(1))), NewDecCoin(nbnkt, NewInt(1000000000)))

	coins, err := ParseCoinsNormalized("0.1bnkt,0.1mbnkt,0.1ubnkt")
	s.Require().NoError(err)
	s.Require().Equal(coins, Coins{
		Coin{nbnkt, NewInt(100000000)},
		Coin{nbnkt, NewInt(100000)},
		Coin{nbnkt, NewInt(100)},
	})

	testCases := []struct {
		input  DecCoin
		denom  string
		result DecCoin
		expErr bool
	}{
		{NewDecCoin("foo", ZeroInt()), bnkt, DecCoin{}, true},
		{NewDecCoin(bnkt, ZeroInt()), "foo", DecCoin{}, true},
		{NewDecCoin(bnkt, ZeroInt()), "FOO", DecCoin{}, true},

		// 0.5bnkt
		{NewDecCoinFromDec(bnkt, NewDecWithPrec(5, 1)), mbnkt, NewDecCoin(mbnkt, NewInt(500)), false},       // bnkt => mbnkt
		{NewDecCoinFromDec(bnkt, NewDecWithPrec(5, 1)), ubnkt, NewDecCoin(ubnkt, NewInt(500000)), false},    // bnkt => ubnkt
		{NewDecCoinFromDec(bnkt, NewDecWithPrec(5, 1)), nbnkt, NewDecCoin(nbnkt, NewInt(500000000)), false}, // bnkt => nbnkt

		{NewDecCoin(ubnkt, NewInt(5000000)), mbnkt, NewDecCoin(mbnkt, NewInt(5000)), false},       // ubnkt => mbnkt
		{NewDecCoin(ubnkt, NewInt(5000000)), nbnkt, NewDecCoin(nbnkt, NewInt(5000000000)), false}, // ubnkt => nbnkt
		{NewDecCoin(ubnkt, NewInt(5000000)), bnkt, NewDecCoin(bnkt, NewInt(5)), false},            // ubnkt => bnkt

		{NewDecCoin(mbnkt, NewInt(5000)), nbnkt, NewDecCoin(nbnkt, NewInt(5000000000)), false}, // mbnkt => nbnkt
		{NewDecCoin(mbnkt, NewInt(5000)), ubnkt, NewDecCoin(ubnkt, NewInt(5000000)), false},    // mbnkt => ubnkt
	}

	for i, tc := range testCases {
		res, err := ConvertDecCoin(tc.input, tc.denom)
		s.Require().Equal(
			tc.expErr, err != nil,
			"unexpected error; tc: #%d, input: %s, denom: %s", i+1, tc.input, tc.denom,
		)
		s.Require().Equal(
			tc.result, res,
			"invalid result; tc: #%d, input: %s, denom: %s", i+1, tc.input, tc.denom,
		)
	}

	// reset registration
	baseDenom = ""
	denomUnits = map[string]Dec{}
}

func (s *internalDenomTestSuite) TestDecOperationOrder() {
	dec, err := NewDecFromStr("11")
	s.Require().NoError(err)
	s.Require().NoError(RegisterDenom("unit1", dec))
	dec, err = NewDecFromStr("100000011")
	s.Require().NoError(RegisterDenom("unit2", dec))

	coin, err := ConvertCoin(NewCoin("unit1", NewInt(100000011)), "unit2")
	s.Require().NoError(err)
	s.Require().Equal(coin, NewCoin("unit2", NewInt(11)))

	// reset registration
	baseDenom = ""
	denomUnits = map[string]Dec{}
}
