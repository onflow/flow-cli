package flow

/* WIP
func CreateTransaction(signerAccount *cli.Account, tx *flow.Transaction) (*flow.Transaction, error) {

  signerAddress := signerAccount.Address()

  fmt.Printf("Getting information for account with address 0x%s ...\n", signerAddress.Hex())

  account, err := client.GetAccount(ctx, signerAddress)
  if err != nil {
    return nil, fmt.Errorf("Failed to get account with address %s: 0x%s", signerAddress.Hex(), err)
  }

  // Default 0, i.e. first key
  accountKey := account.Keys[0]

  sealed, err := client.GetLatestBlockHeader(ctx, true)
  if err != nil {
    return nil, fmt.Errorf("Failed to get latest sealed block: %s", err)
  }

  tx.SetReferenceBlockID(sealed.ID).
    SetProposalKey(signerAddress, accountKey.Index, accountKey.SequenceNumber).
    SetPayer(signerAddress)

  err = tx.SignEnvelope(signerAddress, accountKey.Index, signerAccount.DefaultKey().Signer())
  if err != nil {
    return nil, fmt.Errorf("Failed to sign transaction: %s", err)
  }

  fmt.Printf("Submitting transaction with ID %s ...\n", tx.ID())

  err = client.SendTransaction(context.Background(), *tx)
  if err == nil {
    fmt.Printf("Successfully submitted transaction with ID %s\n", tx.ID())
  } else {
    return nil, fmt.Errorf("Failed to submit transaction: %s", err)
  }

  return tx, nil
}

func GetResult() {

}
*/
