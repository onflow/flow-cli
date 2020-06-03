package sdk

import (
	"fmt"

	flowcrypto "github.com/dapperlabs/flow-go/crypto"
	flowhash "github.com/dapperlabs/flow-go/crypto/hash"
	flowgo "github.com/dapperlabs/flow-go/model/flow"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	sdk "github.com/onflow/flow-go-sdk"
	sdkcrypto "github.com/onflow/flow-go-sdk/crypto"
)

func SDKIdentifierToFlow(sdkIdentifier sdk.Identifier) flowgo.Identifier {
	return flowgo.Identifier(sdkIdentifier)
}

func FlowIdentifierToSDK(flowIdentifier flowgo.Identifier) sdk.Identifier {
	return sdk.Identifier(flowIdentifier)
}

func FlowIdentifiersToSDK(flowIdentifiers []flowgo.Identifier) []sdk.Identifier {
	ret := make([]sdk.Identifier, len(flowIdentifiers))
	for i, flowIdentifier := range flowIdentifiers {
		ret[i] = FlowIdentifierToSDK(flowIdentifier)
	}
	return ret
}

func SDKProposalKeyToFlow(sdkProposalKey sdk.ProposalKey) flowgo.ProposalKey {
	return flowgo.ProposalKey{
		Address:        SDKAddressToFlow(sdkProposalKey.Address),
		KeyID:          uint64(sdkProposalKey.KeyID),
		SequenceNumber: sdkProposalKey.SequenceNumber,
	}
}

func FlowProposalKeyToSDK(flowProposalKey flowgo.ProposalKey) sdk.ProposalKey {
	return sdk.ProposalKey{
		Address:        FlowAddressToSDK(flowProposalKey.Address),
		KeyID:          int(flowProposalKey.KeyID),
		SequenceNumber: flowProposalKey.SequenceNumber,
	}
}

func SDKAddressToFlow(sdkAddress sdk.Address) flowgo.Address {
	return flowgo.Address(sdkAddress)
}

func FlowAddressToSDK(flowAddress flowgo.Address) sdk.Address {
	return sdk.Address(flowAddress)
}

func SDKAddressesToFlow(sdkAddresses []sdk.Address) []flowgo.Address {
	ret := make([]flowgo.Address, len(sdkAddresses))
	for i, sdkAddress := range sdkAddresses {
		ret[i] = SDKAddressToFlow(sdkAddress)
	}
	return ret
}

func FlowAddressesToSDK(flowAddresses []flowgo.Address) []sdk.Address {
	ret := make([]sdk.Address, len(flowAddresses))
	for i, flowAddress := range flowAddresses {
		ret[i] = FlowAddressToSDK(flowAddress)
	}
	return ret
}

func SDKTransactionSignatureToFlow(sdkTransactionSignature sdk.TransactionSignature) flowgo.TransactionSignature {
	return flowgo.TransactionSignature{
		Address:     SDKAddressToFlow(sdkTransactionSignature.Address),
		SignerIndex: sdkTransactionSignature.SignerIndex,
		KeyID:       uint64(sdkTransactionSignature.KeyID),
		Signature:   sdkTransactionSignature.Signature,
	}
}

func FlowTransactionSignatureToSDK(flowTransactionSignature flowgo.TransactionSignature) sdk.TransactionSignature {
	return sdk.TransactionSignature{
		Address:     FlowAddressToSDK(flowTransactionSignature.Address),
		SignerIndex: flowTransactionSignature.SignerIndex,
		KeyID:       int(flowTransactionSignature.KeyID),
		Signature:   flowTransactionSignature.Signature,
	}
}

func SDKTransactionSignaturesToFlow(sdkTransactionSignatures []sdk.TransactionSignature) []flowgo.TransactionSignature {
	ret := make([]flowgo.TransactionSignature, len(sdkTransactionSignatures))
	for i, sdkTransactionSignature := range sdkTransactionSignatures {
		ret[i] = SDKTransactionSignatureToFlow(sdkTransactionSignature)
	}
	return ret
}

func FlowTransactionSignaturesToSDK(flowTransactionSignatures []flowgo.TransactionSignature) []sdk.TransactionSignature {
	ret := make([]sdk.TransactionSignature, len(flowTransactionSignatures))
	for i, flowTransactionSignature := range flowTransactionSignatures {
		ret[i] = FlowTransactionSignatureToSDK(flowTransactionSignature)
	}
	return ret
}

func SDKTransactionToFlow(sdkTx sdk.Transaction) (flowgo.TransactionBody, error) {
	var err error

	arguments := make([][]byte, len(sdkTx.Arguments))
	for i, arg := range sdkTx.Arguments {
		arguments[i], err = jsoncdc.Encode(arg)
		if err != nil {
			return flowgo.TransactionBody{}, fmt.Errorf("failed to encode argument at index %d: %w", i, err)
		}
	}

	return flowgo.TransactionBody{
		ReferenceBlockID:   SDKIdentifierToFlow(sdkTx.ReferenceBlockID),
		Script:             sdkTx.Script,
		Arguments:          arguments,
		GasLimit:           sdkTx.GasLimit,
		ProposalKey:        SDKProposalKeyToFlow(sdkTx.ProposalKey),
		Payer:              SDKAddressToFlow(sdkTx.Payer),
		Authorizers:        SDKAddressesToFlow(sdkTx.Authorizers),
		PayloadSignatures:  SDKTransactionSignaturesToFlow(sdkTx.PayloadSignatures),
		EnvelopeSignatures: SDKTransactionSignaturesToFlow(sdkTx.EnvelopeSignatures),
	}, nil
}

func FlowTransactionToSDK(flowTx flowgo.TransactionBody) (sdk.Transaction, error) {
	var err error

	arguments := make([]cadence.Value, len(flowTx.Arguments))
	for i, arg := range flowTx.Arguments {
		arguments[i], err = jsoncdc.Decode(arg)
		if err != nil {
			return sdk.Transaction{}, fmt.Errorf("failed to decode argument at index %d: %w", i, err)
		}
	}

	return sdk.Transaction{
		ReferenceBlockID:   FlowIdentifierToSDK(flowTx.ReferenceBlockID),
		Script:             flowTx.Script,
		Arguments:          arguments,
		GasLimit:           flowTx.GasLimit,
		ProposalKey:        FlowProposalKeyToSDK(flowTx.ProposalKey),
		Payer:              FlowAddressToSDK(flowTx.Payer),
		Authorizers:        FlowAddressesToSDK(flowTx.Authorizers),
		PayloadSignatures:  FlowTransactionSignaturesToSDK(flowTx.PayloadSignatures),
		EnvelopeSignatures: FlowTransactionSignaturesToSDK(flowTx.EnvelopeSignatures),
	}, nil
}

func FlowEventToSDK(flowEvent flowgo.Event) (sdk.Event, error) {
	cadenceValue, err := jsoncdc.Decode(flowEvent.Payload)
	if err != nil {
		return sdk.Event{}, err
	}

	cadenceEvent, ok := cadenceValue.(cadence.Event)
	if !ok {
		return sdk.Event{}, fmt.Errorf("cadence value not of type event: %s", cadenceValue)
	}

	return sdk.Event{
		Type:             string(flowEvent.Type),
		TransactionID:    FlowIdentifierToSDK(flowEvent.TransactionID),
		TransactionIndex: int(flowEvent.TransactionIndex),
		EventIndex:       int(flowEvent.EventIndex),
		Value:            cadenceEvent,
	}, nil
}

func FlowEventsToSDK(flowEvents []flowgo.Event) ([]sdk.Event, error) {
	ret := make([]sdk.Event, len(flowEvents))
	var err error
	for i, flowEvent := range flowEvents {
		ret[i], err = FlowEventToSDK(flowEvent)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func FlowSignAlgoToSDK(signAlgo flowcrypto.SigningAlgorithm) sdkcrypto.SignatureAlgorithm {
	return sdkcrypto.StringToSignatureAlgorithm(signAlgo.String())
}

func SDKSignAlgoToFlow(signAlgo sdkcrypto.SignatureAlgorithm) flowcrypto.SigningAlgorithm {
	return flowcrypto.StringToSignatureAlgorithm(signAlgo.String())
}

func flowhashAlgoToSDK(hashAlgo flowhash.HashingAlgorithm) sdkcrypto.HashAlgorithm {
	return sdkcrypto.StringToHashAlgorithm(hashAlgo.String())
}

func SDKHashAlgoToFlow(hashAlgo sdkcrypto.HashAlgorithm) flowhash.HashingAlgorithm {
	return flowhash.StringToHashAlgorithm(hashAlgo.String())
}

func FlowAccountPublicKeyToSDK(flowPublicKey flowgo.AccountPublicKey, id int) (sdk.AccountKey, error) {
	// TODO - Looks like SDK contains copy-paste of code from flow-go
	// Once crypto become its own separate library, this can possibly be simplified or not needed
	encodedPublicKey := flowPublicKey.PublicKey.Encode()

	sdkSignAlgo := FlowSignAlgoToSDK(flowPublicKey.SignAlgo)

	sdkPublicKey, err := sdkcrypto.DecodePublicKey(sdkSignAlgo, encodedPublicKey)
	if err != nil {
		return sdk.AccountKey{}, err
	}

	sdkHashAlgo := flowhashAlgoToSDK(flowPublicKey.HashAlgo)

	return sdk.AccountKey{
		ID:             id,
		PublicKey:      sdkPublicKey,
		SigAlgo:        sdkSignAlgo,
		HashAlgo:       sdkHashAlgo,
		Weight:         flowPublicKey.Weight,
		SequenceNumber: flowPublicKey.SeqNumber,
	}, nil
}

func SDKAccountPublicKeyToFlow(sdkPublicKey sdk.AccountKey) (flowgo.AccountPublicKey, error) {
	encodedPublicKey := sdkPublicKey.PublicKey.Encode()

	flowSignAlgo := SDKSignAlgoToFlow(sdkPublicKey.SigAlgo)

	flowPublicKey, err := flowcrypto.DecodePublicKey(flowSignAlgo, encodedPublicKey)
	if err != nil {
		return flowgo.AccountPublicKey{}, err
	}

	flowhashAlgo := SDKHashAlgoToFlow(sdkPublicKey.HashAlgo)

	return flowgo.AccountPublicKey{
		PublicKey: flowPublicKey,
		SignAlgo:  flowSignAlgo,
		HashAlgo:  flowhashAlgo,
		Weight:    sdkPublicKey.Weight,
		SeqNumber: sdkPublicKey.SequenceNumber,
	}, nil
}

func FlowAccountPublicKeysToSDK(flowPublicKeys []flowgo.AccountPublicKey) ([]*sdk.AccountKey, error) {
	ret := make([]*sdk.AccountKey, len(flowPublicKeys))
	for i, flowPublicKey := range flowPublicKeys {
		v, err := FlowAccountPublicKeyToSDK(flowPublicKey, i)
		if err != nil {
			return nil, err
		}
		ret[i] = &v
	}
	return ret, nil
}

func FlowAccountToSDK(flowAccount flowgo.Account) (sdk.Account, error) {
	sdkPublicKeys, err := FlowAccountPublicKeysToSDK(flowAccount.Keys)
	if err != nil {
		return sdk.Account{}, err
	}
	return sdk.Account{
		Address: FlowAddressToSDK(flowAccount.Address),
		Balance: flowAccount.Balance,
		Code:    flowAccount.Code,
		Keys:    sdkPublicKeys,
	}, nil
}

func FlowHeaderToSDK(flowHeader *flowgo.Header) sdk.BlockHeader {
	return sdk.BlockHeader{
		ID:        FlowIdentifierToSDK(flowHeader.ID()),
		ParentID:  FlowIdentifierToSDK(flowHeader.ParentID),
		Height:    flowHeader.Height,
		Timestamp: flowHeader.Timestamp,
	}
}

func FlowCollectionGuaranteeToSDK(flowGuarantee flowgo.CollectionGuarantee) sdk.CollectionGuarantee {
	return sdk.CollectionGuarantee{
		CollectionID: FlowIdentifierToSDK(flowGuarantee.CollectionID),
	}
}

func FlowCollectionGuaranteesToSDK(flowGuarantees []*flowgo.CollectionGuarantee) []*sdk.CollectionGuarantee {
	ret := make([]*sdk.CollectionGuarantee, len(flowGuarantees))
	for i, flowGuarantee := range flowGuarantees {
		sdkGuarantee := FlowCollectionGuaranteeToSDK(*flowGuarantee)
		ret[i] = &sdkGuarantee
	}
	return ret
}

func FlowSealToSDK(flowSeal flowgo.Seal) sdk.BlockSeal {
	return sdk.BlockSeal{
		//TODO
	}
}

func FlowSealsToSDK(flowSeals []*flowgo.Seal) []*sdk.BlockSeal {
	ret := make([]*sdk.BlockSeal, len(flowSeals))
	for i, flowSeal := range flowSeals {
		sdkSeal := FlowSealToSDK(*flowSeal)
		ret[i] = &sdkSeal
	}
	return ret
}

func FlowPayloadToSDK(flowPayload *flowgo.Payload) sdk.BlockPayload {
	return sdk.BlockPayload{
		CollectionGuarantees: FlowCollectionGuaranteesToSDK(flowPayload.Guarantees),
		Seals:                FlowSealsToSDK(flowPayload.Seals),
	}
}

func FlowBlockToSDK(flowBlock flowgo.Block) sdk.Block {
	h := FlowHeaderToSDK(flowBlock.Header)
	p := FlowPayloadToSDK(flowBlock.Payload)

	return sdk.Block{
		BlockHeader:  h,
		BlockPayload: p,
	}
}

func FlowLightCollectionToSDK(flowCollection flowgo.LightCollection) sdk.Collection {
	return sdk.Collection{
		TransactionIDs: FlowIdentifiersToSDK(flowCollection.Transactions),
	}
}
