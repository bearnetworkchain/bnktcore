package feegrant_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
)

func TestBasicFeeValidAllow(t *testing.T) {
	app := simapp.Setup(t, false)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	badTime := ctx.BlockTime().AddDate(0, 0, -1)
	allowace := &feegrant.BasicAllowance{
		Expiration: &badTime,
	}
	require.Error(t, allowace.ValidateBasic())

	ctx = app.BaseApp.NewContext(false, tmproto.Header{
		Time: time.Now(),
	})
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 10))
	bnkt := sdk.NewCoins(sdk.NewInt64Coin("bnkt", 555))
	smallBnkt := sdk.NewCoins(sdk.NewInt64Coin("bnkt", 43))
	bigBnkt := sdk.NewCoins(sdk.NewInt64Coin("bnkt", 1000))
	leftBnkt := sdk.NewCoins(sdk.NewInt64Coin("bnkt", 512))
	now := ctx.BlockTime()
	oneHour := now.Add(1 * time.Hour)

	cases := map[string]struct {
		allowance *feegrant.BasicAllowance
		// all other checks are ignored if valid=false
		fee       sdk.Coins
		blockTime time.Time
		valid     bool
		accept    bool
		remove    bool
		remains   sdk.Coins
	}{
		"empty": {
			allowance: &feegrant.BasicAllowance{},
			accept:    true,
		},
		"small fee without expire": {
			allowance: &feegrant.BasicAllowance{
				SpendLimit: bnkt,
			},
			fee:     smallBnkt,
			accept:  true,
			remove:  false,
			remains: leftBnkt,
		},
		"all fee without expire": {
			allowance: &feegrant.BasicAllowance{
				SpendLimit: smallBnkt,
			},
			fee:    smallBnkt,
			accept: true,
			remove: true,
		},
		"wrong fee": {
			allowance: &feegrant.BasicAllowance{
				SpendLimit: smallBnkt,
			},
			fee:    eth,
			accept: false,
		},
		"non-expired": {
			allowance: &feegrant.BasicAllowance{
				SpendLimit: bnkt,
				Expiration: &oneHour,
			},
			valid:     true,
			fee:       smallBnkt,
			blockTime: now,
			accept:    true,
			remove:    false,
			remains:   leftBnkt,
		},
		"expired": {
			allowance: &feegrant.BasicAllowance{
				SpendLimit: bnkt,
				Expiration: &now,
			},
			valid:     true,
			fee:       smallBnkt,
			blockTime: oneHour,
			accept:    false,
			remove:    true,
		},
		"fee more than allowed": {
			allowance: &feegrant.BasicAllowance{
				SpendLimit: bnkt,
				Expiration: &oneHour,
			},
			valid:     true,
			fee:       bigBnkt,
			blockTime: now,
			accept:    false,
		},
		"with out spend limit": {
			allowance: &feegrant.BasicAllowance{
				Expiration: &oneHour,
			},
			valid:     true,
			fee:       bigBnkt,
			blockTime: now,
			accept:    true,
		},
		"expired no spend limit": {
			allowance: &feegrant.BasicAllowance{
				Expiration: &now,
			},
			valid:     true,
			fee:       bigBnkt,
			blockTime: oneHour,
			accept:    false,
		},
	}

	for name, stc := range cases {
		tc := stc // to make scopelint happy
		t.Run(name, func(t *testing.T) {
			err := tc.allowance.ValidateBasic()
			require.NoError(t, err)

			ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime(tc.blockTime)

			// now try to deduct
			removed, err := tc.allowance.Accept(ctx, tc.fee, []sdk.Msg{})
			if !tc.accept {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tc.remove, removed)
			if !removed {
				assert.Equal(t, tc.allowance.SpendLimit, tc.remains)
			}
		})
	}
}
