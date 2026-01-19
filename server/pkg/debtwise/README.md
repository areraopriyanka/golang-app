Generated debtwise code from root with:

```
oapi-codegen --generate types,client --response-type-suffix "ResponseResult" --include-tags "Users,Accounts,Credit Scores,Equifax,Debt Payoff Plan,Debt Payoff Plan Payments,Debt Calculator" --package debtwise spec/debtwise-mockoon.json > pkg/debtwise/client.gen.go
```

`--include-tags` is a comma delimited list that allows us refine which code we generate.
