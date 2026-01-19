# Process-api (Middleware)

ProcessAPI is the main component in middleware server layer, this component is mainly responsible for,

1. Handling authentication and authorization of onboarding customers.
2. Handling customer onboarding process.
3. Storing customer onboarding data like personal details, KYC/KYB result, etc. and onboarding files data.
4. Storing onboarding files on S3.
5. Maintaining last onboarding step successfully completed by customer to allow him to smoothly proceed with next step if onboarding process gets terminated in between.
6. During the onboarding process if any Ledger API needs to be called, Process API takes care of signing it with its own public key.
7. Connecting with KYC-connector for KYC and KYB process required during onboarding of a customer.
8. Forwarding Ledger API calls signed with customer’s public key as it is to Ledger and returning the response received as it is to mobile application.
9. Maintaining list of accounts bookmarked by customers.

---

## Install Dependencies

1. <https://nixos.org/download/>
2. [Install direnv](https://direnv.net/docs/installation.html#from-binary-builds). Make sure to then follow the [hook installation](https://direnv.net/docs/hook.html) step.
3. (osx only) `nix-env -iA bash -f https://github.com/NixOS/nixpkgs/tarball/nixpkgs-unstable`
4. `cp .env.sample .env`. You will need the secrets from another developer.
5. `cd` to project directory and `direnv allow`. Wait for install to complete.

### How does this work?

- direnv automatically runs a script when you cd into the project directory
- this script
  - loads the `.env` from your working directory into your shells' environment variables
  - starts a nix shell using flake.nix
- nix uses the flake.nix and flake.lock to make an isolated environment for all dependencies
  - this isolates your projects environment and binaries from the rest of your system
  - this makes it easy to add and share dependencies between developers

## Run Instructions

Go to code base path and run below commands in given sequence,

1. `templ generate` - Generate Go code from .templ files in templates/
2. `go build` – To build application executable with latest code changes. This will create an executable file named as `process-api.exe`
3. `docker compose up -d` to run background services.

- This includes the process-api itself on port 5002. You can point the mobile app at your local IP:5002 to use that instance.
  - You can optionally provide `.env.docker` to override env variables.

4. `.\process-api {config_name}`: This command will run the application using configuration file name mentioned in command. Example, `.\process-api dreamfiDev`

### Admin setup

- Request invite to auth0 team: <https://accounts.auth0.com/teams/team-ww3hui6/>
- Attempt auth. Should see access denied
- Browse to dashboard: <https://manage.auth0.com/dashboard/us/dev-e4oqq2rr0qm3jdhc/>
- Sidebar Users
  - Find user in list
  - Add role: operations

## Lint Instructions

    golangci-lint is installed by nix, it can be ran locally via command `golangci-lint run`
    All default linters recommended by golandci-lint at https://golangci-lint.run/usage/linters/ ("errcheck", "gosimple", "govet", "ineffassign", "staticcheck", "unused") as well as "gofmt" and "gofumpt

## Swagger Docs Instructions

    echo-swagger will automatically generate docs for endpoints that are commented using the declarative comment format found here: https://github.com/swaggo/swag#declarative-comments-format
    Once you have commented the endpoint you want to add run "./generate_swagger.sh" from the command line to generate the endpoint documentation in swagger.yaml and swagger.json. swag CLI is made available by flake
    To view the swagger UI for documented endpoints visit localhost:5000/swagger/index.html while the application is running locally

## Useful commands

### Seed the database

```
go run cmd/seedDatabase/main.go
```

### Create a new migration

```
goose --dir ./pkg/db/scripts create <NEW_MIGRATION> sql
# or
./goose.sh create <NEW_MIGRATION> sql
```

### Redo migration

```
./goose.sh redo
```

goose.sh will pass all args to the goose command with postgresql, env variables, and dir set appropriately.

### Access PostgreSQL database shells

```
docker compose exec -it postgresql sh -c 'psql --user="$POSTGRES_USER" $POSTGRES_DB'
docker compose exec -it postgresql-test sh -c 'psql --user="$POSTGRES_USER" $POSTGRES_DB'
```

### Run integration tests

```
docker compose up -d --wait postgresql-test sendgrid twilio debtwise plaid
go test ./integration-tests
```

### Create a test user

Run this script, replacing `[you@example.com]` with your email address.
The provided email will have `+[random-hash]` appended to the name portion e.g., `you@example.com` will become `you+129c18d@example.com`.

```bash
./execute-ledger-onboarding.sh [you@example.com]
```

This script bypasses the KYC step.

After running the script, log in with the mobile app, and a public key/application key will be associated with the user.

ℹ️ After these processes complete, you'll have three files in the `test-data/` directory that correspond to the user:

1. `$USER_ID_public.pem` stores the user's public key in PEM format
2. `$USER_ID_private.txt` stores the user's private key in PEM format
3. `$USER_ID_key_id.txt` stores the user's `keyId`

## Third party services

### Plaid

#### Testing

To test Plaid webhooks locally:

1. Create an `https` tunnel (with e.g., `ngrok`)
2. Update your `.env` with the new `PLAID_WEBHOOKURL` (e.g., `https://154d500ded02.ngrok-free.app/process-api/evolvingsb/plaid/webhooks`)
3. Restart your local server
4. Run `go run cmd/migrate/main.go plaid update-webhooks` to update your existing webhooks
5. Run `go run cmd/plaidSandbox/main.go` to fire test webhook events from Plaid's sandbox

## Troubleshooting

### NetXD Ledger API is failing and I don't understand why

- Request AWS access to DreamFi's "NetXD" account. Craig and Michele are currently admins.
- Get your API keys from the portal: <https://d-9067f79b5f.awsapps.com/start/#/?tab=accounts>
- Configure the local aws cli tool. You get this for free with nix and .env

```
aws ec2 describe-instances --query 'Reservations[].Instances[].{Name:Tags[?Key==`Name`] | [0].Value,InstanceId:InstanceId,PrivateIpAddress:PrivateIpAddress}' | jq .
```

This will list the available ec2 instances name, instanceId and PrivateIpAddress. You'll want the instanceId of "Crump-Sandbox-NetXD-Ledger&Analytics". This will likely by "i-09e8c4b7777df0d89"

```
aws ssm start-session --target "i-09e8c4b7777df0d89"
$ cd /opt/PL
$ tail -f nohup.out
```
