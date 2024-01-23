package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	ibcante "github.com/cosmos/ibc-go/v6/modules/core/ante"
	ethante "github.com/evmos/ethermint/app/ante"
	txfeesante "github.com/osmosis-labs/osmosis/v15/x/txfees/ante"

	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

func newEthAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		ethante.NewEthSetUpContextDecorator(options.EvmKeeper),

		//FIXME: need to allow universal fees for Eth as well
		ethante.NewEthMempoolFeeDecorator(options.EvmKeeper),                           // Check eth effective gas price against minimal-gas-prices
		ethante.NewEthMinGasPriceDecorator(options.FeeMarketKeeper, options.EvmKeeper), // Check eth effective gas price against the global MinGasPrice

		ethante.NewEthValidateBasicDecorator(options.EvmKeeper),
		ethante.NewEthSigVerificationDecorator(options.EvmKeeper),
		ethante.NewEthAccountVerificationDecorator(options.AccountKeeper, options.EvmKeeper),
		ethante.NewCanTransferDecorator(options.EvmKeeper),
		ethante.NewEthGasConsumeDecorator(options.EvmKeeper, options.MaxTxGasWanted),
		ethante.NewEthIncrementSenderSequenceDecorator(options.AccountKeeper), // innermost AnteDecorator.
		ethante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
		ethante.NewEthEmitEventDecorator(options.EvmKeeper), // emit eth tx hash and index at the very last ante handler.
	)
}

// newLegacyCosmosAnteHandlerEip712 creates an AnteHandler to process legacy EIP-712
// transactions, as defined by the presence of an ExtensionOptionsWeb3Tx extension.
func newLegacyCosmosAnteHandlerEip712(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		ethante.RejectMessagesDecorator{}, // reject MsgEthereumTxs
		ethante.NewAuthzLimiterDecorator([]string{ // disable the Msg types that cannot be included on an authz.MsgExec msgs field
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}),
		},
		),
		authante.NewSetUpContextDecorator(),
		authante.NewValidateBasicDecorator(),
		authante.NewTxTimeoutHeightDecorator(),

		//FIXME: need to allow universal fees for EIP712
		ethante.NewMinGasPriceDecorator(options.FeeMarketKeeper, options.EvmKeeper),
		authante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, nil),

		authante.NewValidateMemoDecorator(options.AccountKeeper),
		authante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		// SetPubKeyDecorator must be called before all signature verification decorators
		authante.NewSetPubKeyDecorator(options.AccountKeeper),
		authante.NewValidateSigCountDecorator(options.AccountKeeper),
		authante.NewSigGasConsumeDecorator(options.AccountKeeper, ethante.DefaultSigVerificationGasConsumer),
		// Note: signature verification uses EIP instead of the cosmos signature validator
		NewLegacyEip712SigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		authante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		ethante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	)
}

func newCosmosAnteHandler(options HandlerOptions) sdk.AnteHandler {
	mempoolFeeDecorator := txfeesante.NewMempoolFeeDecorator(*options.TxFeesKeeper)
	deductFeeDecorator := txfeesante.NewDeductFeeDecorator(*options.TxFeesKeeper, options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper)

	return sdk.ChainAnteDecorators(
		ethante.RejectMessagesDecorator{}, // reject MsgEthereumTxs
		ethante.NewAuthzLimiterDecorator([]string{ // disable the Msg types that cannot be included on an authz.MsgExec msgs field
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}),
		},
		),
		ante.NewSetUpContextDecorator(),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		// Use Mempool Fee Decorator from our txfees module instead of default one from auth
		mempoolFeeDecorator,
		deductFeeDecorator,
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, ethante.DefaultSigVerificationGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		ethante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	)
}
