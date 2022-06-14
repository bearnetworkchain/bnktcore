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

func TestPeriodicFeeValidAllow(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		Time: time.Now(),
	})

	bnkt := sdk.NewCoins(sdk.NewInt64Coin("bnkt", 555))
	smallBnkt := sdk.NewCoins(sdk.NewInt64Coin("bnkt", 43))
	leftBnkt := sdk.NewCoins(sdk.NewInt64Coin("bnkt", 512))
	oneBnkt := sdk.NewCoins(sdk.NewInt64Coin("bnkt", 1))
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 1))

	now := ctx.BlockTime()
	oneHour := now.Add(1 * time.Hour)
	twoHours := now.Add(2 * time.Hour)
	tenMinutes := time.Duration(10) * time.Minute

	cases := map[string]struct {
		allow         feegrant.PeriodicAllowance
		fee           sdk.Coins
		blockTime     time.Time
		valid         bool // all other checks are ignored if valid=false
		accept        bool
		remove        bool
		remains       sdk.Coins
		remainsPeriod sdk.Coins
		periodReset   time.Time
	}{
		"empty": {
			allow: feegrant.PeriodicAllowance{},
			valid: false,
		},
		"only basic": {
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
					SpendLimit: bnkt,
					Expiration: &oneHour,
				},
			},
			valid: false,
		},
		"empty basic": {
			allow: feegrant.PeriodicAllowance{
				Period:           tenMinutes,
				PeriodSpendLimit: smallBnkt,
				PeriodReset:      now.Add(30 * time.Minute),
			},
			blockTime:   now,
			valid:       true,
			accept:      true,
			remove:      false,
			periodReset: now.Add(30 * time.Minute),
		},
		"mismatched currencies": {
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
					SpendLimit: bnkt,
					Expiration: &oneHour,
				},
				Period:           tenMinutes,
				PeriodSpendLimit: eth,
			},
			valid: false,
		},
		"same period": {
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
					SpendLimit: bnkt,
					Expiration: &twoHours,
				},
				Period:           tenMinutes,
				PeriodReset:      now.Add(1 * time.Hour),
				PeriodSpendLimit: leftBnkt,
				PeriodCanSpend:   smallBnkt,
			},
			valid:         true,
			fee:           smallBnkt,
			blockTime:     now,
			accept:        true,
			remove:        false,
			remainsPeriod: nil,
			remains:       leftBnkt,
			periodReset:   now.Add(1 * time.Hour),
		},
		"step one period": {
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
					SpendLimit: bnkt,
					Expiration: &twoHours,
				},
				Period:           tenMinutes,
				PeriodReset:      now,
				PeriodSpendLimit: leftBnkt,
			},
			valid:         true,
			fee:           leftBnkt,
			blockTime:     now.Add(1 * time.Hour),
			accept:        true,
			remove:        false,
			remainsPeriod: nil,
			remains:       smallBnkt,
			periodReset:   oneHour.Add(tenMinutes), // one step from last reset, not now
		},
		"step limited by global allowance": {
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
					SpendLimit: smallBnkt,
					Expiration: &twoHours,
				},
				Period:           tenMinutes,
				PeriodReset:      now,
				PeriodSpendLimit: bnkt,
			},
			valid:         true,
			fee:           oneBnkt,
			blockTime:     oneHour,
			accept:        true,
			remove:        false,
			remainsPeriod: smallBnkt.Sub(oneBnkt...),
			remains:       smallBnkt.Sub(oneBnkt...),
			periodReset:   oneHour.Add(tenMinutes), // one step from last reset, not now
		},
		"period reset no spend limit": {
			allow: feegrant.PeriodicAllowance{
				Period:           tenMinutes,
				PeriodReset:      now,
				PeriodSpendLimit: bnkt,
			},
			valid:       true,
			fee:         bnkt,
			blockTime:   oneHour,
			accept:      true,
			remove:      false,
			periodReset: oneHour.Add(tenMinutes), // one step from last reset, not now
		},
		"expired": {
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
					SpendLimit: bnkt,
					Expiration: &now,
				},
				Period:           time.Hour,
				PeriodSpendLimit: smallBnkt,
			},
			valid:     true,
			fee:       smallBnkt,
			blockTime: oneHour,
			accept:    false,
			remove:    true,
		},
		"over period limit": {
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
					SpendLimit: bnkt,
					Expiration: &now,
				},
				Period:           time.Hour,
				PeriodReset:      now.Add(1 * time.Hour),
				PeriodSpendLimit: leftBnkt,
				PeriodCanSpend:   smallBnkt,
			},
			valid:     true,
			fee:       leftBnkt,
			blockTime: now,
			accept:    false,
			remove:    true,
		},
	}

	for name, stc := range cases {
		tc := stc // to make scopelint happy
		t.Run(name, func(t *testing.T) {
			err := tc.allow.ValidateBasic()
			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime(tc.blockTime)
			// now try to deduct
			remove, err := tc.allow.Accept(ctx, tc.fee, []sdk.Msg{})
			if !tc.accept {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tc.remove, remove)
			if !remove {
				assert.Equal(t, tc.remains, tc.allow.Basic.SpendLimit)
				assert.Equal(t, tc.remainsPeriod, tc.allow.PeriodCanSpend)
				assert.Equal(t, tc.periodReset.String(), tc.allow.PeriodReset.String())
			}
		})
	}
}
