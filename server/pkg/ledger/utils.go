package ledger

func GetTransactionAccountMerchantName(account ListTransactionsByAccountResultTransactionAccount) string {
	// TODO: Use the merchant name provided by the ledger once it is provided
	if len(account.Party.Name) != 0 {
		return account.Party.Name
	}
	if len(account.Nickname) != 0 {
		return account.Nickname
	}
	if len(account.CustomerName) != 0 {
		return account.CustomerName
	}
	if len(account.InstitutionName) != 0 {
		return account.InstitutionName
	}
	return "Transaction"
}
