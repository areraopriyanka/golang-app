# Salesforce Integration

## Testing

### Update `.env`:

Ensure `SALESFORCE_INTEGRATIONENABLED` env var is `true`

```bash
SALESFORCE_INTEGRATIONENABLED=true
SALESFORCE_INBOUNDAUTH0DOMAIN=dev-lzx1sufof70qu2du.us.auth0.com
SALESFORCE_INBOUNDAUTH0AUDIENCE=https://middleware.dreamfi.com/salesforce
```

(reminder: allow env vars to be picked up, and restart server)

### Get an access token:

* Replace `client_secret`'s `REDACTED` value with the value found in [Auth0 > Applications > DreamFi SalesForce Integration > Settings](https://manage.auth0.com/dashboard/us/dev-e4oqq2rr0qm3jdhc/applications/nN8QmSh7lAq8Q90LVjossg1bntu2v3pG/settings)
* note: this command begins with a space character, which, in some/most bash setups, prevents the command from being written to `~.bash_history`, which offers some protection when adding secrets to commands (in this case, `client_secret`)

```bash
 curl --request POST \
    --url https://dev-e4oqq2rr0qm3jdhc.us.auth0.com/oauth/token \
    --header 'content-type: application/json' \
    --data '{
      "client_id": "nN8QmSh7lAq8Q90LVjossg1bntu2v3pG",
      "client_secret": "REDACTED",
      "audience": "https://middleware.dreamfi.com/salesforce",
      "grant_type": "client_credentials"
    }'
```

and then extract the `access_token` field from the response payload.

### API calls

Replace `REDACTED` with `access_token` from the last step.

Replace `http://localhost:5050` as needed for your setup.

Replace `ACCOUNT_ID` with a viable `LEDGER_ACCOUNT_ID`.

Try the `/transactions` and `/balance` endpoints.

```bash
   curl -H "Authorization: Bearer REDACTED" http://localhost:5050/salesforce/accounts/LEDGER_ACCOUNT_ID/transactions
```

```bash
   curl -H "Authorization: Bearer REDACTED" http://localhost:5050/salesforce/accounts/LEDGER_ACCOUNT_ID/balance
```

#### Testing Account IDs
 What's a good `account_id` to use? It depends. I used `docker compose exec -it postgresql sh -c 'psql --user="$POSTGRES_USER" $POSTGRES_DB'` and then tried some account ids until I found some that had some/one/none transactions.

```sql
> select account_id from user_account_card;
 account_id
------------
 64421214
 71369108
 21281196
 21281208
 21281334
 21281338
 21281405
 26438023
 31605335
 44522031
 ```


## Generating Swagger Docs

After updating any of the pkg/salesforce endpoints, run:

```bash
./generate_swagger.sh --salesforce
```

And share the output with SMI.
