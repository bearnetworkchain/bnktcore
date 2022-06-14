package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestMetadataValidate(t *testing.T) {
	testCases := []struct {
		name     string
		metadata types.Metadata
		expErr   bool
	}{
		{
			"non-empty coins",
			types.Metadata{
				Name:        "Cosmos Hub Bnkt",
				Symbol:      "BNKT",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"ubnkt", uint32(0), []string{"microbnkt"}},
					{"mbnkt", uint32(3), []string{"millibnkt"}},
					{"bnkt", uint32(6), nil},
				},
				Base:    "ubnkt",
				Display: "bnkt",
			},
			false,
		},
		{
			"base coin is display coin",
			types.Metadata{
				Name:        "Cosmos Hub Bnkt",
				Symbol:      "BNKT",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"bnkt", uint32(0), []string{"BNKT"}},
				},
				Base:    "bnkt",
				Display: "bnkt",
			},
			false,
		},
		{"empty metadata", types.Metadata{}, true},
		{
			"blank name",
			types.Metadata{
				Name: "",
			},
			true,
		},
		{
			"blank symbol",
			types.Metadata{
				Name:   "Cosmos Hub Bnkt",
				Symbol: "",
			},
			true,
		},
		{
			"invalid base denom",
			types.Metadata{
				Name:   "Cosmos Hub Bnkt",
				Symbol: "BNKT",
				Base:   "",
			},
			true,
		},
		{
			"invalid display denom",
			types.Metadata{
				Name:    "Cosmos Hub Bnkt",
				Symbol:  "BNKT",
				Base:    "ubnkt",
				Display: "",
			},
			true,
		},
		{
			"duplicate denom unit",
			types.Metadata{
				Name:        "Cosmos Hub Bnkt",
				Symbol:      "BNKT",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"ubnkt", uint32(0), []string{"microbnkt"}},
					{"ubnkt", uint32(1), []string{"microbnkt"}},
				},
				Base:    "ubnkt",
				Display: "bnkt",
			},
			true,
		},
		{
			"invalid denom unit",
			types.Metadata{
				Name:        "Cosmos Hub Bnkt",
				Symbol:      "BNKT",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"", uint32(0), []string{"microbnkt"}},
				},
				Base:    "ubnkt",
				Display: "bnkt",
			},
			true,
		},
		{
			"invalid denom unit alias",
			types.Metadata{
				Name:        "Cosmos Hub Bnkt",
				Symbol:      "BNKT",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"ubnkt", uint32(0), []string{""}},
				},
				Base:    "ubnkt",
				Display: "bnkt",
			},
			true,
		},
		{
			"duplicate denom unit alias",
			types.Metadata{
				Name:        "Cosmos Hub Bnkt",
				Symbol:      "BNKT",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"ubnkt", uint32(0), []string{"microbnkt", "microbnkt"}},
				},
				Base:    "ubnkt",
				Display: "bnkt",
			},
			true,
		},
		{
			"no base denom unit",
			types.Metadata{
				Name:        "Cosmos Hub Bnkt",
				Symbol:      "BNKT",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"mbnkt", uint32(3), []string{"millibnkt"}},
					{"bnkt", uint32(6), nil},
				},
				Base:    "ubnkt",
				Display: "bnkt",
			},
			true,
		},
		{
			"base denom exponent not zero",
			types.Metadata{
				Name:        "Cosmos Hub Bnkt",
				Symbol:      "BNKT",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"ubnkt", uint32(1), []string{"microbnkt"}},
					{"mbnkt", uint32(3), []string{"millibnkt"}},
					{"bnkt", uint32(6), nil},
				},
				Base:    "ubnkt",
				Display: "bnkt",
			},
			true,
		},
		{
			"invalid denom unit",
			types.Metadata{
				Name:        "Cosmos Hub Bnkt",
				Symbol:      "BNKT",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"ubnkt", uint32(0), []string{"microbnkt"}},
					{"", uint32(3), []string{"millibnkt"}},
				},
				Base:    "ubnkt",
				Display: "ubnkt",
			},
			true,
		},
		{
			"no display denom unit",
			types.Metadata{
				Name:        "Cosmos Hub Bnkt",
				Symbol:      "BNKT",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"ubnkt", uint32(0), []string{"microbnkt"}},
				},
				Base:    "ubnkt",
				Display: "bnkt",
			},
			true,
		},
		{
			"denom units not sorted",
			types.Metadata{
				Name:        "Cosmos Hub Bnkt",
				Symbol:      "BNKT",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"ubnkt", uint32(0), []string{"microbnkt"}},
					{"bnkt", uint32(6), nil},
					{"mbnkt", uint32(3), []string{"millibnkt"}},
				},
				Base:    "ubnkt",
				Display: "bnkt",
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.metadata.Validate()

			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMarshalJSONMetaData(t *testing.T) {
	cdc := codec.NewLegacyAmino()

	testCases := []struct {
		name      string
		input     []types.Metadata
		strOutput string
	}{
		{"nil metadata", nil, `null`},
		{"empty metadata", []types.Metadata{}, `[]`},
		{
			"non-empty coins",
			[]types.Metadata{
				{
					Description: "The native staking token of the Cosmos Hub.",
					DenomUnits: []*types.DenomUnit{
						{"ubnkt", uint32(0), []string{"microbnkt"}}, // The default exponent value 0 is omitted in the json
						{"mbnkt", uint32(3), []string{"millibnkt"}},
						{"bnkt", uint32(6), nil},
					},
					Base:    "ubnkt",
					Display: "bnkt",
				},
			},
			`[{"description":"The native staking token of the Cosmos Hub.","denom_units":[{"denom":"ubnkt","aliases":["microbnkt"]},{"denom":"mbnkt","exponent":3,"aliases":["millibnkt"]},{"denom":"bnkt","exponent":6}],"base":"ubnkt","display":"bnkt"}]`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			bz, err := cdc.MarshalJSON(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.strOutput, string(bz))

			var newMetadata []types.Metadata
			require.NoError(t, cdc.UnmarshalJSON(bz, &newMetadata))

			if len(tc.input) == 0 {
				require.Nil(t, newMetadata)
			} else {
				require.Equal(t, tc.input, newMetadata)
			}
		})
	}
}
