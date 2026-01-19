## 0.1.16 - 29-Nov-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed ledger kyc URL's and credentials.
   getkycurl: "https://dreamfisb.netxd.com/ekyc/rpc/KycService/GetTransactionByReferenceId"
   postkycurl: "https://dreamfisb.netxd.com/ekyc/rpc/KycService/Check"
   addkycdocumentsurl: "https://dreamfisb.netxd.com/ekyc/rpc/KycService/AddDocuments"
   getkycdocumentsurl: "https://dreamfisb.netxd.com/ekyc/rpc/KycService/GetDocuments"
-----------------------------------------------------------------------------

## 0.1.15 - 14-Nov-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[EM-723](https://evolvingllc.atlassian.net/browse/EM-723)
- CronJob to delete 60 minutes older ledger token records from Middleware DB
- This cronjob would be executed after every 30 minutes and delete records from ledger_users table whose createdAt less than ledger token exp time(60 minutes).
-----------------------------------------------------------------------------

## 0.1.14 - 14-Nov-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[EM-724](https://evolvingllc.atlassian.net/browse/EM-724)
- Changed column size for below table fields and set to idle size
   users: email(320)
   user-otp: receiver (320), createdBy(320)
   ledger_users: email(320), jwtToken(2000), createdBy(320), updatedBy(320)
-----------------------------------------------------------------------------
## 0.1.13 - 12-Nov-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Added code to handle notification for old customers
-----------------------------------------------------------------------------

## 0.1.12 - 11-Nov-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[EM-668](https://evolvingllc.atlassian.net/issues/EM-668)
- Dependencies upgraded
-----------------------------------------------------------------------------

## 0.1.11 - 07-Nov-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[EM-711](https://evolvingllc.atlassian.net/issues/EM-711)
- Changed below ledger url for qa-ledger to fix 404 error
  endpoint: "https://plus.netxd.com/pl/jsonrpc"
  cardsEndpoint: "https://plus.netxd.com/pl/cardv2"
  paymentsEndpoint: "https://plus.netxd.com/pl/rpc/paymentv2"
  mobileRpcEndpoint: "https://plus.netxd.com/pl/jsonrpc"
  createSubscriptionApiUrl: "https://plus.netxd.com/pl/rpc/WebhookService/CreateSubscription"
-----------------------------------------------------------------------------

## 0.1.10 - 06-Nov-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed sandbox config file
-----------------------------------------------------------------------------

## 0.1.9 - 28-Oct-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[EM-695](https://evolvingllc.atlassian.net/browse/EM-695)
- Change error message in case of Invalid Otp for verifyEmailVerificationOtp Api and verifyResetPasswordOtp Api Response
- New message is Please enter valid one time password.
-----------------------------------------------------------------------------

## 0.1.8 - 18-Oct-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[EM-623](https://evolvingllc.atlassian.net/browse/EM-623)
- Updated log messages.
- Removed unessential log statements.
-----------------------------------------------------------------------------

## 0.1.7 - 22-Oct-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[EM-624](https://evolvingllc.atlassian.net/browse/EM-624)
- Adding response in kyc_error_response table in case of error from kyc ledger API
- Added create script for kyc_error_response
-----------------------------------------------------------------------------

## 0.1.6 - 16-Oct-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[EM-613](https://evolvingllc.atlassian.net/browse/EM-613)
- Added new kyc status as KYC_INITIATED.
- Setting this status till we get response from ledger kyc API
- Exposed new API version url: http://{{ProcessApiEndpoint}}/{{clientUrl}}/onboard/kyc/v2 method:POST
-----------------------------------------------------------------------------

## 0.1.5 - 15-Oct-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changes done in the template
-----------------------------------------------------------------------------

## 0.1.4 - 14-Oct-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[EM-549](https://evolvingllc.atlassian.net/browse/EM-549)
[EM-562](https://evolvingllc.atlassian.net/browse/EM-562)
1. Expose new version of GenerateEmailVerificationOtp API
    - Request Method changed to post
    - Request data contains firstName and lastName
    - URL: POST http://{{ProcessApiEndpoint}}/{{clientUrl}}/onboard/email/otp/v2?email=priyanka.arerao@springct.com&isRetry=false

2. Expose new API version for ResetPasswordOtp API
    - URL: POST  http://{{ProcessApiEndpoint}}/{{clientUrl}}/login/reset/emailotp/v2
-----------------------------------------------------------------------------

## 0.1.3 - 14-Oct-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Direct-deposite template changed.
-----------------------------------------------------------------------------

## 0.1.2 - 09-Oct-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed fcm config for sandbox as changes are not integrated in mobile.
-----------------------------------------------------------------------------

## 0.1.1 - 07-Oct-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Refactored RenderStaticContent handler function. 
- Changed dash color. Added padding to number.
-----------------------------------------------------------------------------

## 0.1.0 - 04-Oct-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- Removed NetXd code.
-----------------------------------------------------------------------------

## 0.17.20_hotfix2 - 11-oct-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Added aria-hidden="true" property
- Added cursor: default; property
- Added line-height
- Removed js
-----------------------------------------------------------------------------------------

## 0.17.20_hotfix1 - 04-oct-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
[EM-497](https://evolvingllc.atlassian.net/browse/EM-497)
### Changed
- Changed html template used inline css instead of internal
- Changed url to http://{{ProcessApiEndpoint}}/{{clientUrl}}/static/DIRECT_DEPOSITS
- Moved .html file to static dir
- Added new group with MiddlewareTokenAuthentication
-----------------------------------------------------------------------------------------

## 0.17.20 - 03-oct-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
[EM-497](https://evolvingllc.atlassian.net/browse/EM-497)
### Changed
- Exposed an API to send direct deposit template. Added to middlewareledger group.
- URL: http://{{ProcessApiEndpoint}}/{{clientUrl}}/jsonrpc/template/direct-deposits/DIRECT_DEPOSITS.
-----------------------------------------------------------------------------------------

## 0.17.19 - 30-sep-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
[EM-482](https://evolvingllc.atlassian.net/browse/EM-482)
### Changed
- Changed otp expiry time to 3 minutes.
- Sending otpExpiryDuration in response of sendEmailVerificationOtp and resetPasswordOtp response.
-----------------------------------------------------------------------------------------

## 0.17.18 - 27-sep-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
[EM-479](https://evolvingllc.atlassian.net/browse/EM-479)
### Changed
- Loading env specific configuration file for FCM
- Reading filename form env specific config file
- Removed conditional code
-----------------------------------------------------------------------------------------

## 0.17.17_hotfix3 - 25-sep-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Removed request body from the logs which contains password
- Added Logs when API name field is changed 
-----------------------------------------------------------------------------------------

## 0.17.17_hotfix2 - 23-sep-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed 'Beneficiary' to 'Account owner' in notification msg
-----------------------------------------------------------------------------------------

## 0.17.17_hotfix1 - 20-sep-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Removed approval related transactions and biller related notifications
- Added Transaction.UPDATE approve event
-----------------------------------------------------------------------------------------

## 0.17.17 - 13-sep-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Added DreamFi webhook credentials and FCM config file
- Changed date format in push notification for DreamFi
-----------------------------------------------------------------------------------------

## 0.17.16 - 09-sep-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-1569](https://netxd.atlassian.net/browse/MAPP-1569)
- Changed category type for DELETE_PENDING to BENEFICIARY_DELETED_PENDING for Business user
-----------------------------------------------------------------------------------------

## 0.17.15 - 06-sep-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Added condition in Transaction decline case where send notification only for credit = false flag
-----------------------------------------------------------------------------------------

## 0.17.14 - 02-sep-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed sendPushNotification code to handle edge case for sending notification to only last logged in user device.
- Added status column in customer_device_subscription_mapping.
- Removed RemoveDuplicate function
- Added ReadMe file
-----------------------------------------------------------------------------------------
## 0.17.13 - 02-sep-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed Transaction declined message to
Transaction is declined
Transaction of $50.00 is declined for your CHECKING account ending in XXXXXXXXXXXXXX2427 on 2024-09-02
-----------------------------------------------------------------------------------------

## 0.17.12 - 02-sep-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Added constants in templates.
- Changed error messages(Removed fullstop)
-----------------------------------------------------------------------------------------

## 0.17.11 - 30-aug-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed in ledger_user table. Removed primary key for User_Email column and added id as a PK
- Creating new record each time for ledger login in ledger_users
- Sending  UNAUTHORIZED_USER_TO_ACCESS_NOTIFICATION error for invalid customer number in GetNotificationPayload
-----------------------------------------------------------------------------------------

## 0.17.10 - 22-aug-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added 
- Added constants in webhook handler
-------------------------------------------------------------------------------------------

## 0.17.09 - 13-aug-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1494](https://netxd.atlassian.net/browse/MAPP-1494)
### Added 
- Formatted biller,beneficiery,account statement and update related push notification
- Added api to get notification payload
-------------------------------------------------------------------------------------------

## 0.17.08 - 09-aug-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1494](https://netxd.atlassian.net/browse/MAPP-1494)
### Added 
- Formatted transaction related push notification
-------------------------------------------------------------------------------------------

## 0.17.7_hotfix7 - 17-sep-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Added config file for dreamfi-qa env
-----------------------------------------------------------------------------------------

## 0.17.7_hotfix6 - 13-sept-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
[EM-396](https://evolvingllc.atlassian.net/browse/EM-396)
### Changed
- Added addCardResponse column in onboardingData table and storing addCard response in that column.
- Sending addCardResponse field in getOnboardingData response.  
-------------------------------------------------------------------------------------------

## 0.17.7_hotfix5 - 11-sept-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
[EM-396](https://evolvingllc.atlassian.net/browse/EM-396)
### Changed
- Changed addCardVirtual to addCard to handle add physical card flow during onboarding.
-------------------------------------------------------------------------------------------

## 0.17.7_hotfix4 - 10-sept-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
[EM-388](https://evolvingllc.atlassian.net/browse/EM-388)
### Changed
- Changed dreamfiSandbox host url to https://dreamfisb.netxd.com/pl.
- Changes in below url’s at middleware:
   ledgerendpoint: "https://dreamfisb.netxd.com/pl/jsonrpc"
   cardsEndpoint: "https://dreamfisb.netxd.com/pl/cardv2"
   paymentsEndpoint: "https://dreamfisb.netxd.com/pl/rpc/paymentv2"
   mobileRpcEndpoint: "https://dreamfisb.netxd.com/gw/mobilerpc"
-------------------------------------------------------------------------------------------

## 0.17.7_hotfix3 - 05-sept-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed error message for invalid password to Incorrect password entered.
-------------------------------------------------------------------------------------------

## 0.17.7_hotfix2 - 03-sept-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- Added New API for Add dreamfi saving account.
-------------------------------------------------------------------------------------------

## 0.17.7_hotfix - 30-august-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed OTP related error msgs
-------------------------------------------------------------------------------------------

## 0.17.07 - 05-aug-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[EM-292](https://evolvingllc.atlassian.net/browse/EM-292)
### Changed 
- Changes in fundFlowToday api
  1. Added condition to call get transaction intrafi api only for netxd savings acc
- Ledger token validation is removed for below apis 
  1. ManageSignedLedgerCall
  2. ManageUnsignedLedgerApiCall
  3. ManageCards
  4. ManagePayments
  5. ManageMobileRPC
  6. RefreshToken
-------------------------------------------------------------------------------------------

## 0.17.06 - 01-aug-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1434](https://netxd.atlassian.net/browse/MAPP-1434)
### Added 
- Signature verification added for webhook
- Api url is changed for get count notification - getNotificationCount
-------------------------------------------------------------------------------------------

## 0.17.05 - 31-july-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[EM-285](https://evolvingllc.atlassian.net/browse/EM-285)
### Added 
- INDIVIDUAL_ADDSAVINGSACCOUNTCATEGORY_CONFIG is added for dreamfi
- Reading webhook private key from secrete manager
-------------------------------------------------------------------------------------------

## 0.17.04 - 29-july-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1436](https://netxd.atlassian.net/browse/MAPP-1436)
### Added 
- Webhook
1. Added api to mark notification as read based on notificatio id and customer number.
2. Added logic for beneficiary business and individual customer category
-------------------------------------------------------------------------------------------

## 0.17.03 - 26-july-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1435](https://netxd.atlassian.net/browse/MAPP-1435)
### Added 
- Webhook
1. Added api to get device details from UI and create subscription
-------------------------------------------------------------------------------------------

## 0.17.02 - 25-july-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1432](https://netxd.atlassian.net/browse/MAPP-1432)
### Added 
- Webhook
1. Added api to clear notification
2. Added logic to form category and notification type for biller and beneficiary events
-------------------------------------------------------------------------------------------

## 0.17.01 - 23-july-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1424](https://netxd.atlassian.net/browse/MAPP-1424)
### Added 
- Webhook
1. Added api to filter notification and get count based on category
2. Added logic to save webhook events in middleware db
3. Logic to decide category for transaction events
-------------------------------------------------------------------------------------------

## 0.16.17 - 24-july-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
- Changed buffer refresh time to 15 mnts on qa-netxd
-------------------------------------------------------------------------------------------

## 0.16.16 - 19-july-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1391](https://netxd.atlassian.net/browse/MAPP-1391)
### Added 
- Added config file for dreamfi dev env
-------------------------------------------------------------------------------------------

## 0.16.15 - 15-july-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1391](https://netxd.atlassian.net/browse/MAPP-1391)
### Changed 
- Changed kyc url and credentials for netxd sandbox and dreamfi
-------------------------------------------------------------------------------------------

## 0.16.14 - 11-july-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added 
[MAPP-1293](https://netxd.atlassian.net/browse/MAPP-1293)
1. Added handler to fetch events from webhook
2. Changed channel from API to INTERNAL in INTERNAL_OUT_PAYMENTS_DEFAULT_CONFIG for qa ledger and dreamfi
-------------------------------------------------------------------------------------------

## 0.16.13 - 08-july-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added 
[MAPP-1365](https://netxd.atlassian.net/browse/MAPP-1365)
- Added api to kill user's session
1. Middleware url - http://{{ProcessApiEndpoint}}/{{clientUrl}}/killSession
2. Sample request body
   {
   "userName": "pooname9@netxd.com"
   }
3. Success response - Status code ok(200)
-------------------------------------------------------------------------------------------

## 0.16.12 - 05-july-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added 
[MAPP-1076](https://netxd.atlassian.net/browse/MAPP-1076)
- API sequencing 
1. In order to track last success api in case of onboarding is completed partially , apiName is sent as part of login response
2. API name in sequence is renamed
- Changed session time out of dreamfi to 2 mnts
-------------------------------------------------------------------------------------------

## 0.16.11 - 05-july-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
[EM-148](https://evolvingllc.atlassian.net/browse/EM-148)
### Removed
- Removed saveNonUSUser endpoint and handler function
- Removed saveNonUSUser DB script
- commented DreamFi group
------------------------------------------------------------------------------------------

## 0.16.10 - 20-june-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1297](https://netxd.atlassian.net/browse/MAPP-1297)
### Added 
- Added scheduler to delete unused and expired otp from middleware db
- Added max file size config script
-------------------------------------------------------------------------------------------

## 0.16.9 - 19-june-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1295](https://netxd.atlassian.net/browse/MAPP-1295)
### Changed 
- Changes in bookmark
1. Bookmark api should be accessible only by onboarded users
2. UI will send below params for Add and delete bookmark
	1.legalRepId 
	2.accNo
3. 'legalRepId' will be sent as a param for GetAllBookmark api.
4. In GetAllBookmark api, bookmark will be filtered at middleware based on legal rep id provided by UI.
-------------------------------------------------------------------------------------------

## 0.16.8 - 13-june-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1265](https://netxd.atlassian.net/browse/MAPP-1265)
### Changed 
- Added ledger token validation
-------------------------------------------------------------------------------------------

## 0.16.7 - 11-june-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[EM-83](https://evolvingllc.atlassian.net/browse/EM-83)
### Changed 
- Rename Evolve to DreamFi
-------------------------------------------------------------------------------------------

## 0.16.6 - 05-june-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
- Added generalized error msg in api response
- Changed session timeout on qa to 30 minutes
=======
## 0.16.5_hotfix1 - 15-july-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1391](https://netxd.atlassian.net/browse/MAPP-1391)
### Changed 
- Changed kyc url and credentials for netxd sandbox
-------------------------------------------------------------------------------------------
=======
  2. For dreamfi, GetTransaction api will be called for saving account fundFlowToday
- Text OTP is changed to 'One Time Password' for dreamfi
-----------------------------------------------------------------------------------------

## 0.16.5_hotfix1 - 17-july-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1391](https://netxd.atlassian.net/browse/MAPP-1391)
### Changed 
- Changed kyc url and credentials for dreamfi
- Changed channel from API to INTERNAL in INTERNAL_OUT_PAYMENTS_DEFAULT_CONFIG for dreamfi
-----------------------------------------------------------------------------------------

## 0.16.5 - 30-may-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
- Changed session timeout on test to 2 minutes
-------------------------------------------------------------------------------------------

## 0.16.4 - 30-may-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
[MAPP-1214](https://netxd.atlassian.net/browse/MAPP-1214)
### Changed 
- Changed error msg in api response. Showing below msg instead of technical details
{
    "code": "INTERNAL_SERVER_ERROR",
    "msg": "Internal server error."
}
-------------------------------------------------------------------------------------------

## 0.16.3 - 24-may-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
[MAPP-1134](https://netxd.atlassian.net/browse/MAPP-1134)
- Removed prod config file and default values from config.go
- Added method to load config from aws secret manager
- Handled error while starting the server 
-------------------------------------------------------------------------------------------

## 0.16.2 - 24-may-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
- Log msg refactoring
-------------------------------------------------------------------------------------------

## 0.16.1 - 17-may-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added 
[MAPP-1166](https://netxd.atlassian.net/browse/MAPP-1166)
- Added wrapper for api UpdateAccountSettings
1. Middleware url: http://{{ProcessApiEndpoint}}/{{clientUrl}}/onboard/account/settings
2. Success response will contain customerNumber and status
3. User status will be updated to UPDATE_ACCOUNT_SETTINGS_SUCCESSFUL
- Added variable 'clientUrl' in api url(in main.go) 
-------------------------------------------------------------------------------------------

## 0.16.0 - 13-may-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added 
[MAPP-1132](https://netxd.atlassian.net/browse/MAPP-1132)
- Added feature of db versioning
-------------------------------------------------------------------------------------------

## 0.15.6 - 13-may-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added 
- Added different JWT secrete key for each envs
-------------------------------------------------------------------------------------------

## 0.15.5 - 10-may-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
[MAPP-1125](https://netxd.atlassian.net/browse/MAPP-1125)
- Added new group evolveGroup
- Added middleware for this group for evolve API validation
------------------------------------------------------------------------------------------

## 0.15.4 - 09-may-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
[MAPP-1074](https://netxd.atlassian.net/browse/MAPP-1074)
- Changed mobileRpc url from "https://demobox.netxd.com/gw/mobilerpc" to "https://demobox.netxd.com/pl/jsonrpc" in demo config
-------------------------------------------------------------------------------------------

## 0.15.3 - 26-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added 
[MAPP-1076](https://netxd.atlassian.net/browse/MAPP-1076)
- Changes in GetConfig api
   1. Added column 'userType' in config table which will accept only 'ONBOARDING', 'ONBOARDED', 'BOTH' 
   2. For onboarding users, only config with 'ONBOARDING' and 'BOTH' will be shown
        "configName": "ADDRESS_MODE_CONFIG",
        "configName": "INDIVIDUAL_EXCLUSION_BANK_COUNT_CONFIG",          
        "configName": "BUSINESS_EXCLUSION_BANK_COUNT_CONFIG",          
        "configName": "INDIVIDUAL_GETQUESTIONAIESBYPRODUCT_CONFIG",
        "configName": "BUSINESS_GETQUESTIONAIESBYPRODUCT_CONFIG",     
        "configName": "INDIVIDUAL_GETTERMSANDCONDTIONBYPRODUCT_CONFIG",
        "configName": "BUSINESS_GETTERMSANDCONDTIONBYPRODUCT_CONFIG",
        "configName": "INDIVIDUAL_ADDACCOUNTCATEGORY_CONFIG",
        "configName": "BUSINESS_ADDACCOUNTCATEGORY_CONFIG",
   3. For onboarded users, only config with 'ONBOARDED' and 'BOTH' will be shown
- Added EndOnboarding api 
   1. Set user status ACTIVE and delete from user table
-------------------------------------------------------------------------------------------

## 0.15.2 - 24-april-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added 
[MAPP-1080](https://netxd.atlassian.net/browse/MAPP-1080)
- Added code to restrict log level to Error for prod ENV.
- Printing Application version, env name and config information in prod env.
-------------------------------------------------------------------------------------------

## 0.15.1 - 22-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
[MAPP-1074](https://netxd.atlassian.net/browse/MAPP-1074)
- Changed url in sandbox from https://sandbox.netxd.com/gw/mobilerpc to https://sandbox.netxd.com/PLMASTER/jsonrpc
--------------------------------------------------------------------------------------------

## 0.15.0 - 22-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added 
[MAPP-1064](https://netxd.atlassian.net/browse/MAPP-1066)
- Added open API to save non US user details for evolve
1. Url: http://{{ProcessApiEndpoint}}/{{clientUrl}}/saveNonUsUser
2. Sample Input payload:
   {
      "firstName":"firstName",
      "lastName":"lastName",
      "email": "sample@domain.com",
      "mobileNo": "1034567896",
      "countryCode":"UK"    
   } 
3. Success response: StatusCode OK
--------------------------------------------------------------------------------------------

## 0.14.51 - 22-april-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Removed
[MAPP-1064](https://netxd.atlassian.net/browse/MAPP-1064)
- Removed below code related to AddCountry API
 1. Middleware AddCountry Endpoint
 2. Handler function for AddCountry
 3. Country field from UserDao struct
- Added Drop column script for country column in 'users' table
 ----------------------------------------------------------------------------------------

## 0.14.50 - 18-april-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Removed
[MAPP-1069](https://netxd.atlassian.net/browse/MAPP-1069)
- Removed below code related to middleware bank exclusion api
1. GetBankExclusionList API endpoint
2. GetBankExclusionList handler function
3. CronJob function
4. BanksDao and ServiceLogsDao structures
5. Banks and Servicelogs DB script
6. SyncBanksScheduler function
- Removed deprecated api endpoint and respective handler functions.
- Removed FundFlowToday_Old and InitiateResetPasswordAndSendEmail_Old API.
-------------------------------------------------------------------------------------------

## 0.14.49 - 17-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
[MAPP-1074](https://netxd.atlassian.net/browse/MAPP-1074)
- Change in ledger url from https://connectors.cbwpayments.com/gw/mobilerpc to https://connectors.cbwpayments.com/PLMASTER/jsonrpc
- Updated url in dev, test, QA-Netxd config
-------------------------------------------------------------------------------------------

## 0.14.48_QAHotfix2 - 24-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
- Changed onboardingUserAutoLogoffTime from 3 minutes to 20 minutes in qa-netxd
--------------------------------------------------------------------------------------------

## 0.14.48_QAHotfix1 - 11-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
[MAPP-1074](https://netxd.atlassian.net/browse/MAPP-1074)
- Change in ledger url from https://connectors.cbwpayments.com/gw/mobilerpc to https://connectors.cbwpayments.com/PLMASTER/jsonrpc
- Updated url in dev,test,QA-NetXD config
-------------------------------------------------------------------------------------------

## 0.14.48 - 12-april-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
[MAPP-1064](https://netxd.atlassian.net/browse/MAPP-1064)
- Added AddCountry API to save country for onboarding user
 1. Middleware Api url: http://{{ProcessApiEndpoint}}/{{clientUrl}}/onboard/addCountry
 2. Query param name is 'countryCode'
 3. Input: 2 char country code
 4. Output: status code 200
--------------------------------------------------------------------------------------------

## 0.14.47_evolvingHotfix1 - 12-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
[MAPP-1066](https://netxd.atlassian.net/browse/MAPP-1066)
- Sending user status as ACTIVE instead of ADD_CHECKING_ACCOUNT_SUCCESSFUL for addCheckingAccount Api.
- Deleting user from db.
--------------------------------------------------------------------------------------------

## 0.14.47_sandboxHotfix1 - 22-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
[MAPP-1074](https://netxd.atlassian.net/browse/MAPP-1074)
Changed url in sandbox from https://sandbox.netxd.com/gw/mobilerpc to https://sandbox.netxd.com/PLMASTER/jsonrpc
--------------------------------------------------------------------------------------------

## 0.14.47 - 11-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
[MAPP-1003](https://netxd.atlassian.net/browse/MAPP-1003)
- Changed client url for demoBox
--------------------------------------------------------------------------------------------

## 0.14.46 - 10-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
[MAPP-1058](https://netxd.atlassian.net/browse/MAPP-1058)
- Added db script for config table 
--------------------------------------------------------------------------------------------

## 0.14.45 - 10-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
[MAPP-1003](https://netxd.atlassian.net/browse/MAPP-1003)
- Changed client url for demoBox
--------------------------------------------------------------------------------------------

## 0.14.44 - 10-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
[MAPP-1003](https://netxd.atlassian.net/browse/MAPP-1003)
- Changed client url for demoBox
--------------------------------------------------------------------------------------------

## 0.14.43 - 10-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
[MAPP-1003](https://netxd.atlassian.net/browse/MAPP-1003)
- Updated aws keys for demoBox
- Added config script for GetQuestionaiesByProduct
--------------------------------------------------------------------------------------------

## 0.14.42 - 10-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
[MAPP-1003](https://netxd.atlassian.net/browse/MAPP-1003)
- Updated aws keys for demoBox
--------------------------------------------------------------------------------------------

## 0.14.41 - 05-april-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
[MAPP-1003](https://netxd.atlassian.net/browse/MAPP-1003)
- Updated configuration for demoBox
- Moved all keys to key folder under 'config'
- Read days to delete older logs from config
--------------------------------------------------------------------------------------------

## 0.14.40 - 27-march-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added 
[MAPP-1003](https://netxd.atlassian.net/browse/MAPP-1003)
- Removed smarty creds from evolving config
- Added config for demoBox - config-demo.yml
- Added links for username in changeLog
--------------------------------------------------------------------------------------------

## 0.14.39 - 20-march-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed cronJob expression to execute it only once at 02:30AM daily
- Removed GetBankExclusionList_LedgerSearch response from log in CallLedgerAPIWithUrlAndGetRawResponse function
--------------------------------------------------------------------------------------------

## 0.14.38 - 19-march-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added 
[MAPP-977](https://netxd.atlassian.net/browse/MAPP-977)
- Changes in evolving config
1. Rename evolve config file name and private key
2. Added evolving kyc details in config file
3. Added db script for GETQUESTIONAIESBYPRODUCT api in middleware config table
4. Removed ID from GetConfig api response
--------------------------------------------------------------------------------------------

## 0.14.37.1 - 22-march-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed cron package version to v3
-------------------------------------------------------------------------------------------

## 0.14.37 - 19-march-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed cronJob expression to execute it only once at 12:30AM daily
--------------------------------------------------------------------------------------------

## 0.14.36 - 18-march-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- Added code to delete 30 day's older log files 
- Goroutine will execute at 12:30AM daily using cronJob
--------------------------------------------------------------------------------------------

## 0.14.35 - 15-march-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Deprecated
[MAPP-967](https://netxd.atlassian.net/browse/MAPP-967)
-  Commented below code
1. GetBankExclusionList API endpoint
2. GetBankExclusionList handler function
3. CronJob function
4. BanksDao and ServiceLogsDao structures
5. Banks and Servicelogs DB script
6. SyncBanksScheduler function
----------------------------------------------------------------------------------------------

## 0.14.34 - 15-march-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added 
[MAPP-968](https://netxd.atlassian.net/browse/MAPP-968)
- Restriction on adding bank exclusion list
1. Middleware config table - Added db script to insert record for count of bank exclusion
2. Added below params in GetConfig API response:
   INDIVIDUAL_EXCLUSION_BANK_COUNT_CONFIG = 1000
   BUSINESS_EXCLUSION_BANK_COUNT_CONFIG = 1000
-----------------------------------------------------------------------------------------------

## 0.14.33 - 06-march-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-886](https://netxd.atlassian.net/browse/MAPP-886)
[MAPP-888](https://netxd.atlassian.net/browse/MAPP-888)
- Changed GetBankExclusionList_LedgerSearch method to POST.
----------------------------------------------------------------------------------------------

## 0.14.32 - 04-march-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added 
[MAPP-895](https://netxd.atlassian.net/browse/MAPP-895)
-Restriction on adding external bank 
1. Middleware config table - Added db script to insert record for count of external bank
2. Individual customer to add at max of 5 External bank
3. Business customer to add at max of 10 external bank
-----------------------------------------------------------------------------------------------

## 0.14.31 - 29-feb-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added 
[MAPP-886](https://netxd.atlassian.net/browse/MAPP-886)
[MAPP-888](https://netxd.atlassian.net/browse/MAPP-888)
- Exposed new route to get bank exclusion list from ledger DB for onboarding user.
- Signed request with middleware key.
-----------------------------------------------------------------------------------------------

## 0.14.30 - 28-feb-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added 
[MAPP-882](https://netxd.atlassian.net/browse/MAPP-882)
- Added BufferTimeRefreshToken param in ledger login response.
- Set BufferTimeRefreshToken value in milliseconds in all configs(5 minutes).
-----------------------------------------------------------------------------------------------

## 0.14.29 - 26-feb-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-861](https://netxd.atlassian.net/browse/MAPP-861)
- Handled error response of GetToken API
- Added null check for 'result' in response
------------------------------------------------------------------------------------------------

## 0.14.28 - 23-feb-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-861](https://netxd.atlassian.net/browse/MAPP-861)
- Refactored AutologoffTime related code. Changed AutologoffTime type to int.
- Removed hardcoded json response structure and used map[string]interface{}.
-------------------------------------------------------------------------------------------------

## 0.14.27 - 22-feb-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
[MAPP-861](https://netxd.atlassian.net/browse/MAPP-861)
- Added AutologoffTime for onboarding and ledger user in configs and set to 3 minutes.
-------------------------------------------------------------------------------------------------

## 0.14.26 - 19-feb-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-824](https://netxd.atlassian.net/browse/MAPP-824)
1. Middleware api url - Read client url from config file 
2. Added separate config file for crump-sandbox
--------------------------------------------------------------------------------------------------

## 0.14.25 - 13-feb-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-716](https://netxd.atlassian.net/browse/MAPP-716)
- Address Mode Config
 1. Changed 'type' to MANUAL in config db script
--------------------------------------------------------------------------------------------------

## 0.14.24 - 12-feb-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
- Added db script to add address mode type in middleware config table
- Type is 'GEOLOCATION'
- Possible 'type' values for 'ADDRESS_MODE_CONFIG' are 1) AUTO 2) MANUAL 3) GEOLOCATION
- Added comment for possible type values in configHandler 
--------------------------------------------------------------------------------------------------

## 0.14.23 - 06-feb-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
- Added QA ledger configs in dev and test config yml
- Added qa ledger configs in config.go
--------------------------------------------------------------------------------------------------

## 0.14.22 - 02-feb-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Fixed
[MAPP-687](https://netxd.atlassian.net/browse/MAPP-687)
- Changed ledger api endpoint for billpay.transfer api to "https://sandbox.netxd.com/PLMASTER/rpc/paymentv2"

-------------------------------------------------------------------------------------------------
## 0.14.21 - 30-jan-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
- Added encryption key for qa-netxd env
- Changed global url var to local
- Changed name of var 'credentials' to 'credential' in config.go as it is 'credential' in all env specific configs
--------------------------------------------------------------------------------------------------

## 0.14.20 - 30-jan-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
- Moved version.go to pkg/version folder
- Corrected log text in authenticationHandler
--------------------------------------------------------------------------------------------------

## 0.14.19 - 29-jan-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- Added GetApplicationVersion api to get application version 
--------------------------------------------------------------------------------------------------

## 0.14.18 - 25-jan-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
- Added Get apis for InitiateResetPasswordAndSendEmail and fundFlow for backward compatibility
- Code refactored in Post apis
--------------------------------------------------------------------------------------------------

## 0.14.17 - 24-jan-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-630](https://netxd.atlassian.net/browse/MAPP-630)
InitiateResetPasswordAndSendEmail Api
- Converted Get to post api
- Changed url of api to "/{{clientUrl}}/login/reset/emailotp" as conflicted with post VerifyResetPasswordOtp api url
--------------------------------------------------------------------------------------------------

## 0.14.16 - 24-jan-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-630](https://netxd.atlassian.net/browse/MAPP-630)
- Changed FundFlowToday API request httpmethod to POST
---------------------------------------------------------------------------------------------------

## 0.14.15 - 24-jan-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-630](https://netxd.atlassian.net/browse/MAPP-630)
- Changed FundFlowToday API to get accNo, accType and mfp through request body instead of queryParam.
- Removed Mfp changes from Register and VerifyUserAlreadyExists API 
----------------------------------------------------------------------------------------------------

## 0.14.14 - 17-jan-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
[MAPP-630](https://netxd.atlassian.net/browse/MAPP-630)
- Added Mfp parameter in ledger Request struct inside Api parameter
- Added Mfp parameter in Login request 
- Changed log statement in GetKycDetailsFromDB to make kycResponse readable
-----------------------------------------------------------------------------------------------------

## 0.14.13 - 16-jan-2024 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
[MAPP-610](https://netxd.atlassian.net/browse/MAPP-610)
- Added accType QueryParam in FundFlowToday API request
- Added "IntraFiConnectorService.GetTransactionsFromIntraFI" method for "SAVINGS" transaction
- Getting result in shadowTransactions parameter. Added this parameter in Fund flow response structure
-------------------------------------------------------------------------------------------------------

## 0.14.12 - 08-jan-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
-Added configs for qa-netxd env
-Renamed stage to sandbox
------------------------------------------------------------------------------------------------------

## 0.14.11 - 03-jan-2024 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-589](https://netxd.atlassian.net/browse/MAPP-589)
-Converted 'LastLoginDateTime' in RFC3339 time format for login api
-Converted 'CreatedAt' in RFC3339 time format for GetAllAccountBookmark api
------------------------------------------------------------------------------------------------------

## 0.14.10 - 21-dec-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- Added new kyc status PASS_AND_REVIEW or FAIL_AND_REVIEW for kyc refer in GetKycStatus API.
- Sending kyc status as KYC_REFER for above status
-----------------------------------------------------------------------------------------------------

## 0.14.9 - 20-dec-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Removed
-Removed config db script for ICS_WITHDRAW and ICS_DEPOSIT
------------------------------------------------------------------------------------------------------

## 0.14.8 - 19-dec-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
-Added config db script for ICS_WITHDRAW and ICS_DEPOSIT
------------------------------------------------------------------------------------------------------

## 0.14.7 - 14-dec-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed InitiateKycResponse when status is KYC_ERROR
----------------------------------------------------------------------------------------------------

## 0.14.6 - 12-dec-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- Added KYC_REFER flow in GetKycStatus API.
- Checking reviewStatus parameter when Kyc_Status is REFER
-----------------------------------------------------------------------------------------------------

## 0.14.5 - 12-dec-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-498](https://netxd.atlassian.net/browse/MAPP-498)
-Added unsigned ledger api call under onboarding group
-Added log msg in reset password handler
------------------------------------------------------------------------------------------------------

## 0.14.4 - 12-dec-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Removed
[MAPP-498](https://netxd.atlassian.net/browse/MAPP-498)
- Removed loginRetryCount field when user logins for the first time part
-----------------------------------------------------------------------------------------------------

## 0.14.3 - 11-dec-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Removed
[MAPP-498](https://netxd.atlassian.net/browse/MAPP-498)
-Removed logic to block onboarding customer for multiple attempts with invalid password
1. Deleted 'isLocked' and 'loginRetryCount' column from user table
2. Deleted 'maxLoginRetryCount' field from all configs
3. Removed code for checking maxLoginRetryCount and isLocked user from authenticationHandler and jwtAuthorization
4. Commented 'UnblockCustomer' admin api from main
------------------------------------------------------------------------------------------------------

## 0.14.2 - 05-dec-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
[MAPP-299](https://netxd.atlassian.net/browse/MAPP-299)
[MAPP-463](https://netxd.atlassian.net/browse/MAPP-463)
-Backend - add config in DB for ACH_PULL
------------------------------------------------------------------------------------------------------

## 0.14.1 - 01-dec-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-6](https://netxd.atlassian.net/browse/MAPP-6)
-Added Makefile to project
------------------------------------------------------------------------------------------------------

## 0.14.0 - 30-nov-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed log message and added customerId and KycRefId
- Added getKycDocuments anf initiateKycResp in GetOnboarding response
-------------------------------------------------------------------------------------------------------------

## 0.13.21 - 29-nov-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added 
[MAPP-290](https://netxd.atlassian.net/browse/MAPP-290)
- Added uploadDocumentsToKycConnector method to upload kyc documents to kyc connector 
- Added uploadSingleDocument method to upload single file to kyc connector
- Added addKycDocumentsToDB method to add uploaded documents to GetKycDocuments field of Onboarding_data table
-------------------------------------------------------------------------------------------------------------

## 0.13.20 - 29-nov-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-382](https://netxd.atlassian.net/browse/MAPP-382)
- Changed BanksDao table's PK to "FDICCertNumber" from "ID"
- Changed Query for updating banks table
- Changed CronJob execution time to 2:30AM UTC
------------------------------------------------------------------------------------------------------------

## 0.13.19 - 24-nov-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- Changed Register User and Reset Password API to store encrypted password to DB
- Changed verify password to compare encrypted password and plain text
--------------------------------------------------------------------------------------------------------

## 0.13.18 - 24-nov-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
[MAPP-428](https://netxd.atlassian.net/browse/MAPP-428)
- Smarty API authentication to be removed from middleware
[MAPP-429](https://netxd.atlassian.net/browse/MAPP-429)
- state and city list API remove authentication from middleware
1. Added public api for get country, state and city list and address autocomplete
------------------------------------------------------------------------------------------------------

## 0.13.17 - 22-nov-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Removed
[MAPP-6](https://netxd.atlassian.net/browse/MAPP-6)
-Removed otp from api response of GenerateEmailVerificationOtp and InitiateResetPasswordAndSendEmail
------------------------------------------------------------------------------------------------------

## 0.13.16 - 17-nov-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
[MAPP-382](https://netxd.atlassian.net/browse/MAPP-382)
- Implement a cron job to pull bank details from ledger
- Added service_logs table to maintain cronjob logs
------------------------------------------------------------------------------------------------------

## 0.13.15 - 13-nov-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
[MAPP-392](https://netxd.atlassian.net/browse/MAPP-392)
- As a bank customer I want to change the password using change password
1. Expose generic api for mobilerpc server
-----------------------------------------------------------------------------------------------------

## 0.13.14 - 10-nov-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-399](https://netxd.atlassian.net/browse/MAPP-399)
- Added update customer settings api
1. Set user status UPDATE_CUSTOMER_SETTINGS_SUCCESSFUL in middleware db for ledger success response
2. Send customer number and updated user status in response
3. Send response as it is from ledger api in case of error
-----------------------------------------------------------------------------------------------------

## 0.13.13 - 07-nov-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-386](https://netxd.atlassian.net/browse/MAPP-386)
- Changes in user status for add legal rep to account api
1. Get account type CHECKING/SAVINGS in query parameter
2. If account type is CHECKING then set user status to ADD_LEGAL_REP_TO_CHECKING_ACCOUNT_SUCCESS in db
3. If account type is SAVINGS then set user status to ADD_LEGAL_REP_TO_SAVINGS_ACCOUNT_SUCCESS in db
-----------------------------------------------------------------------------------------------------

## 0.13.12 - 08-nov-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-380](https://netxd.atlassian.net/browse/MAPP-380)
-GetBankExclusionList API
1.If no records found, return empty result and status ok
--------------------------------------------------------------------------------------------------

## 0.13.11 - 06-nov-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-380](https://netxd.atlassian.net/browse/MAPP-380)
-Added API to get bank exclusion list
1. Added script to create bank table and insert static bank exclusion list data 
2. Api to get bank exclusion list by search query parameter
--------------------------------------------------------------------------------------------------

## 0.13.10 - 03-nov-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Fixed
[MAPP-374](https://netxd.atlassian.net/browse/MAPP-374)
-Fixed bug in unblocking user
1. User was getting deleted after unblocking as status set to 'ACTIVE'
2. To fixed this, removed 'BLOCKED' status of user and status is not set in admiin api
3. Added 'isLocked' column in user table
4. Set 'isLocked'=1 if user is blocked
5. Set 'isLocked'=0 if admin unblocked the user using admin API 
--------------------------------------------------------------------------------------------------

## 0.13.9 - 02-nov-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-368](https://netxd.atlassian.net/browse/MAPP-368)
-Added API to create legal rep login
1. Provided wrapper for ledger 'CreateLegalRepLogin' API
2. Updated customer status 'CREATE_LEGAL_REP_LOGIN_SUCCESS' in middleware db
3. Sent customer number and status in response
--------------------------------------------------------------------------------------------------

## 0.13.8 - 30-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
-Removed temporary changes from addCardHolder API
-Added .jpg extension in config.go
--------------------------------------------------------------------------------------------------

## 0.13.7 - 27-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-334](https://netxd.atlassian.net/browse/MAPP-334)
-Added API to get list of countries from db
1. Inserted static country list in middleware db
--------------------------------------------------------------------------------------------------

## 0.13.6 - 26-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
-Added restrictions for file format before uploading files to s3 bucket
1. Only files with extension .pdf, .jpeg, .png are allowed to upload
--------------------------------------------------------------------------------------------------

## 0.13.5 - 26-oct-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
[MAPP-262](https://netxd.atlassian.net/browse/MAPP-262)
-Make register api open, that can be called without login token.
--------------------------------------------------------------------------------------------------

## 0.13.4 - 26-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-262](https://netxd.atlassian.net/browse/MAPP-262)
-Changed format of login response for middleware
1. Added login response in result struct
2. Changed to init small all fields in response
--------------------------------------------------------------------------------------------------

## 0.13.3 - 25-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-319](https://netxd.atlassian.net/browse/MAPP-319)
- Saved response of below APIs in onboard data table
1. addCustomer
2. addCheckingAcc
3. addSavingAcc
4. addLegalRep
5. addCardHolder
- Fetch updated responses in getOnboardingDetails API
- Changed field name to init small letter in API response
--------------------------------------------------------------------------------------------------

## 0.13.2 - 25-oct-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-262](https://netxd.atlassian.net/browse/MAPP-262)
- Changed AddCardHolder API- Removed ResetToken part once user is deleted from middleware DB. 
-------------------------------------------------------------------------------------------------

## 0.13.1 - 23-oct-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-262](https://netxd.atlassian.net/browse/MAPP-262)
- Added user type in Middleware login response. 
-------------------------------------------------------------------------------------------------

## 0.13.0 - 19-oct-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
[MAPP-262](https://netxd.atlassian.net/browse/MAPP-262)
 - Login flow- If user is an onboarding user then using middleware login flow else if user is not present at middleware DB then calling ledger GetToken API
 - Logout flow- If logout API is called using ledger token then directly logging out user with Success response
 - Ledger token authentiction- Added middleware for ledger token authentiction and checking if token is present or not
 - For all postlogin ledger API's jwt token generated at ledger will be used so during authentication only checking if token present or not, since ledger will take care of validating their token.
 - At end of onboarding process i.e. after AddCardHolderAPI is success and user status is changed to ACTIVE - Deleteing user from middleware DB. After onboarding is done user login will be handled by Ledger.
------------------------------------------------------------------------------------------------

## 0.12.30 - 20-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Commited temporary changes in result of add card holder api
1. Skipped ledger call as ledger api is giving error
2. Set user status ACTIVE in db and sent in response
--------------------------------------------------------------------------------------------------

## 0.12.29 - 19-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-6](https://netxd.atlassian.net/browse/MAPP-6)
- Set PRODUCT in request header for reset and update password api 
--------------------------------------------------------------------------------------------------

## 0.12.28 - 19-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-6](https://netxd.atlassian.net/browse/MAPP-6)
-Ledger forgot password response
1. Added condition to check if 'message' is present in result
--------------------------------------------------------------------------------------------------

## 0.12.27 - 18-oct-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
[MAPP-301](https://netxd.atlassian.net/browse/MAPP-301)
Verify if the user exists in the system while onboarding
--------------------------------------------------------------------------------------------------

## 0.12.26 - 18-oct-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
[MAPP-295](https://netxd.atlassian.net/browse/MAPP-295)
Backend - add new URL for calling reset password and update password APIS
--------------------------------------------------------------------------------------------------

## 0.12.25 - 17-oct-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
[MAPP-288](https://netxd.atlassian.net/browse/MAPP-288)
Modified Ledger caller to pass headers to ledger for all api calls
--------------------------------------------------------------------------------------------------

## 0.12.24 - 17-oct-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-288](https://netxd.atlassian.net/browse/MAPP-288)
Changed LedgerCaller function to pass the request header
--------------------------------------------------------------------------------------------------

## 0.12.23 - 17-oct-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
Changed GetOnboardingDetails and UpdateOnboardingDetails to accept tempCustomer number as well as permanent  
--------------------------------------------------------------------------------------------------

## 0.12.22 - 16-oct-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
[MAPP-288](https://netxd.atlassian.net/browse/MAPP-288)
-Backend - add separate URL for calling dive registration APIS
--------------------------------------------------------------------------------------------------

## 0.12.21 - 16-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-6](https://netxd.atlassian.net/browse/MAPP-6)
-Added blank checks for username and password in update password flow
--------------------------------------------------------------------------------------------------

## 0.12.20 - 16-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-6](https://netxd.atlassian.net/browse/MAPP-6)
-Modified update password flow
1. Called ledger update password api if customer is not present in middleware db or has permanent customer number
2. Updated new password in middleware db if ledger send success response
3. If custmer is temp, check for password validation and update new password in middleware db
--------------------------------------------------------------------------------------------------

## 0.12.19 - 12-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-6](https://netxd.atlassian.net/browse/MAPP-6)
-Renamed GenerateVerificationOtpAndSendEmail handler name to InitiateResetPasswordAndSendEmail
-Added constant for result
--------------------------------------------------------------------------------------------------

## 0.12.18 - 12-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-6](https://netxd.atlassian.net/browse/MAPP-6)
- Modified flow for ledger forgot password and generate email otp
1. Called ledger reset password api if user not found in middleware or has permanent customer number
2. If user has temp customer number sent otp from middleware
--------------------------------------------------------------------------------------------------

## 0.12.17 - 11-oct-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
Refactored Encryption and Decryption part and changed condition in Kyc API to check user type
Inserted UserType in register API
---------------------------------------------------------------------------------------------------

## 0.12.16 - 11-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Refactor the code for legal rep api
1. Added constant
2. Checked if result contains ID
--------------------------------------------------------------------------------------------------

## 0.12.15 - 11-oct-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
Changed Encryption part in UpdateOnboarding details
----------------------------------------------------------------------

## 0.12.14 - 11-oct-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
Changed Decryption part in GetOnboarding details
----------------------------------------------------------------------

## 0.12.13 - 11-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Added legal rep to account api
1. Signed request with key and called AddLegalRepToAccount api
2. Set user status 'ADD_LEGAL_REP_TO_ACCOUNT_SUCCESS' and sent customerNumber and status in response
-----------------------------------------------------------------------------------------------------

## 0.12.12 - 11-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Changed user status to 'ACTIVE' in add card holder api
-----------------------------------------------------------------------------------------------------

## 0.12.11 - 10-oct-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Changed InitiateKyc API and GetKyc API
1. Finding user type from DB instead of taking it from url
2. Stored entire kyc response in kycDetails column as a Raw Response for InitiateKyc
3. Stored entire kyc response in kycDetails column as a Raw Response for GetKyc if no error in response
----------------------------------------------------------------------------------------------------

 ## 0.12.10 - 10-oct-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- modified onboarding details table to store extra information like bank exclusion list. Added new column
- modified getOnboardingDetails and updateOnboardingDetails API response
----------------------------------------------------------------------------------------------------

## 0.12.9 - 10-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Added intrafi add saving account api
1. Signed request with key and called intrafi add saving account api
2. Set user status 'ADD_SAVING_ACCOUNT_SUCCESSFUL' and sent customer,messege and status in response
-----------------------------------------------------------------------------------------------------

## 0.12.8 - 10-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Added card virtual api
1. Signed request with key and called ledger add card virtual api
2. Set user status 'ACTIVE' and sent card details and status in response
-----------------------------------------------------------------------------------------------------

## 0.12.7 - 10-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Added card holder api
1. Signed request with key and called ledger card holder api
2. Set user status 'ADD_CARD_HOLDER_SUCCESSFUL' and sent cardHolderId and status in response
-----------------------------------------------------------------------------------------------------

## 0.12.6 - 10-oct-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Changed email regx and reafctored Register API code
-----------------------------------------------------------------------------------------------------

## 0.12.5 - 10-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Changed user status to 'ADD_CHECKING_ACCOUNT_SUCCESSFUL' in add account api
-----------------------------------------------------------------------------------------------------

## 0.12.4 - 10-oct-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Update encryption flow to support large data encryption
-----------------------------------------------------------------------------------------------------

## 0.12.3 - 09-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Changed response body of add account api
1. Added account number and ID from ledger in response body and send to UI
-----------------------------------------------------------------------------------------------------

## 0.12.2 - 05-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Changed field names in all configs after code review
- Added constant for query params
-----------------------------------------------------------------------------------------------------

## 0.12.1 - 05-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Added smarty US address autocomplete API
-----------------------------------------------------------------------------------------------------

## 0.12.0 - 05-oct-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
[MAPP-111](https://netxd.atlassian.net/browse/MAPP-111)
-  [MAPP-214](https://netxd.atlassian.net/browse/MAPP-214)
- Provided Wrapper API to call add legal rep and bypass its response to update details at server DB
--------------------------------------------------------------------------------------------------------

## 0.11.3 - 05-oct-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
[MAPP-6](https://netxd.atlassian.net/browse/MAPP-6)
- Forgot Password flow
- Generate OTP and send email
- Verify OTP
-----------------------------------------------------------------------------------------------------

## 0.11.2 - 04-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-6](https://netxd.atlassian.net/browse/MAPP-6)
- Lower camel case the json name for 'newPassword' in reset password API
-----------------------------------------------------------------------------------------------------

## 0.11.1 - 04-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-6](https://netxd.atlassian.net/browse/MAPP-6)
- Added reset password API
-----------------------------------------------------------------------------------------------------

## 0.11.0 - 04-oct-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
[MAPP-6](https://netxd.atlassian.net/browse/MAPP-6)
- Forgot Password flow
--------------------------------------------------------------------------------------------------------

## 0.10.19_hotfix1 - 06-oct-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Fixed
[MAPP-245](https://netxd.atlassian.net/browse/MAPP-245)
- Fixed bug for logout API
1. Sent status code 200 for session time out instead of SESSION_TIMEOUT error and 500 status code
--------------------------------------------------------------------------------------------------------

## 0.10.19 - 29-sep-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
[MAPP-212] (https://netxd.atlassian.net/browse/MAPP-212)
- Added API to get Kyc details of logged in user from Database and return in response
--------------------------------------------------------------------------------------------------------

## 0.10.18 - 29-sep-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Changed Kyc API
 [MAPP-152]  Initiate Kyc (https://netxd.atlassian.net/browse/MAPP-152)
 [MAPP-153]  Get Kyc status (https://netxd.atlassian.net/browse/MAPP-153)
 - Changed response for Both KYC Api
 - Setting RefID, ApplicationID and KycDetails to NULL if status = KYC_ERROR
 -------------------------------------------------------------------------------------------------------

## 0.10.17 - 29-sep-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Added add account Api
1. Get account number from ledger and update user status actice in user db
3. Generate new token with updated details and save new token generation dateTime in user table
4. Send error response as it is from ledger if add account fails
5. Refactor the add customer code
-----------------------------------------------------------------------------------------------------

## 0.10.16 - 29-sep-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
[MAPP-175] (https://netxd.atlassian.net/browse/MAPP-175)
- Changed Register API to check user's email or mobile already registered
-----------------------------------------------------------------------------------------------------

## 0.10.15 - 29-sep-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Onboarding User
 [MAPP-172]  Backend - Add data encryption flow while saving onboarding details in MW DB (https://netxd.atlassian.net/browse/MAPP-172)
 Added AWS KMS for encryption decription of onboarding data
-----------------------------------------------------------------------------------------------------

## 0.10.14 - 28-sep-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Changed Kyc API
 [MAPP-152]  Initiate Kyc (https://netxd.atlassian.net/browse/MAPP-152)
 [MAPP-153]  Get Kyc status (https://netxd.atlassian.net/browse/MAPP-153)
 Refactored Both API code and handled KYC_ERROR response
---------------------------------------------------------------------------------------------------

## 0.10.13 - 28-sep-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Added encryption while registering user 
-----------------------------------------------------------------------------------------------------

## 0.10.12 - 27-sep-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Added add customer Api
1. [MAPP-154] Add customer API (https://netxd.atlassian.net/browse/MAPP-154)
2. Get customer number from ledger and update it to onboarding and user db
3. Generate new token with updated details and save new token generation dateTime in user table
4. Send error response as it is from ledger if add customer fails
-----------------------------------------------------------------------------------------------------

## 0.10.11 - 26-sep-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Added Kyc API
1. [MAPP-152]  Initiate Kyc (https://netxd.atlassian.net/browse/MAPP-152)
2. [MAPP-153]  Get Kyc status (https://netxd.atlassian.net/browse/MAPP-153)
3. Updated token after both API to get appropriate response
----------------------------------------------------------------------------------------

## 0.10.10 - 22-sep-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Added utils for encryption and decryption of data
-----------------------------------------------------------------------------------------------------

## 0.10.9 - 22-sep-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Added apis to call ledger APIs with middleware signature.
-----------------------------------------------------------------------------------------------------

## 0.10.8 - 21-sep-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Added apis to get list of US states and cities
- Added db script for create and insert data for state and city
-----------------------------------------------------------------------------------------------------

## 0.10.7 - 21-sep-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- [MAPP-168] - Added GenerateOtp API
- [MAPP-169]- Added VerifyOtp API
   Added SendEmail function

## 0.10.6 - 18-sep-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Opened AWS connection in main
-----------------------------------------------------------------------------------------------------

## 0.10.5 - 18-sep-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
[MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- Added apis for uploading file to S3 and get pre signed url 
-----------------------------------------------------------------------------------------------------

## 0.10.4 - 15-sep-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- Added UpdateOnboarding and GetOnboardingDetails API 
-----------------------------------------------------------------------------------------------------

## 0.10.3 - 15-sep-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
- [MAPP-155] (https://netxd.atlassian.net/browse/MAPP-155)
Individual User Onboarding
 - Provide Wrapper API to call get onboarding questions from Ledger 

## 0.10.2 - 14-sep-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- [MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
Changed RegisterUser API
 - Added code to store data to Onboarding Data table
 - Segragated Login success actions from Login handler to use it in Register User API 

## 0.10.1 - 13-sep-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- [MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
Changed RegisterUser API response. Added User status

## 0.10.0 - 13-sep-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- [MAPP-69](https://netxd.atlassian.net/browse/MAPP-69)
Added RegisetUser API to store user's data at middleware
---------------------------------------------------------------------------------------------------------

## 0.9.6 - 06-sep-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Fixed
 - [MAPP-96](https://netxd.atlassian.net/browse/MAPP-96)
 Fixed bug
 - The transferred amount is reflected only under DR (Debit Amount) not CR for From Account under 'Funds Flow Today'
 - Modified condition to check whether the amount is debited or credited 

## 0.9.5 - 04-sep-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
- [MAPP-11](https://netxd.atlassian.net/browse/MAPP-11)
Added application version

## 0.9.4 - 01-sep-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- [MAPP-11](https://netxd.atlassian.net/browse/MAPP-11)
Changed log file path from root directory to log directory

## 0.9.3 - 31-aug-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Fixed 
- [MAPP-61](https://netxd.atlassian.net/browse/MAPP-61)
Fixed bug
- Getting error message for time parsing as expected time format was minimum 3 digit after decimal.
- Changed time format that will accept time upto 3 digit after decimal

## 0.9.2 - 31-aug-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed 
- [MAPP-11](https://netxd.atlassian.net/browse/MAPP-11)
Optimized code for Formatting logs- Removed CallerPrettyfier function and added that code to Format function 

## 0.9.1 - 30-aug-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed 
- [MAPP-11](https://netxd.atlassian.net/browse/MAPP-11)
Modified CallerPrettyfier function in logger.go to change log format

## 0.9.0 - 30-aug-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-71](https://netxd.atlassian.net/browse/MAPP-71)
1. Add admin API to get all audit trails for process-api
2. Integrate Audit reading code shared by NetXD Team (Jiten)
---------------------------------------------------------------------------------------------------------

## 0.8.4 - 30-aug-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
- [MAPP-11](https://netxd.atlassian.net/browse/MAPP-11)
Modification related to rolling out logs
1. Changed max size of log file from 1 to 20 mb

## 0.8.3 - 29-aug-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
- [MAPP-11](https://netxd.atlassian.net/browse/MAPP-11)
Added logic for rolling out logs
1. Set params to roll out logs based on size and days in logger.go using lumberjack
2. Generated log file name based on time and date
3. Formatted the log msg
4. Added logger reated configuration in all env specific config files

## 0.8.2 - 25-aug-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-45](https://netxd.atlassian.net/browse/MAPP-45)
1. Updated Card API to remove signed by middleware part
2. Remove deprecated code

## 0.8.1 - 23-aug-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-40](https://netxd.atlassian.net/browse/MAPP-40)
Updation in Payment Api
1. Added logger to print request sent to ledger
2. Updated db script of payments configs - Replaced value 'INTERNAL' to 'API' in insert script

## 0.8.0 - 22-aug-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-40](https://netxd.atlassian.net/browse/MAPP-40)
Added payment Api
1. Updated configs script for payments config
2. Added payment endpoint in all env specific config
3. Added ManagePayments handler in paymentHandler.go. Get request body and pass it to ledger API to get response.
-----------------------------------------------------------------------------

## 0.7.14 - 18-aug-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-15](https://netxd.atlassian.net/browse/MAPP-15)
Removed logic for signing manage card request with middleware key from handleCardsServiceCalls() and added in ManageCards()

## 0.7.13 - 14-aug-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Updated token expiry time from 2 minute to 15 minute in config-test.yml

## 0.7.12 - 11-aug-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Added/Removed text.txt for jenkins testing

## 0.7.11 - 10-aug-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Added/Removed text.txt for jenkins testing

## 0.7.10 - 09-aug-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Fixed
- [MAPP-52](https://netxd.atlassian.net/browse/MAPP-52)
Fixed issue in FundFlowToday API
1. Fixed bug- Updated toDate from one day after to current_date
2. Handled error response - Getting different error response from ledger
3. Modified Fund-flow ErrorResponse struct to handle this responses

## 0.7.9 - 09-aug-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-45](https://netxd.atlassian.net/browse/MAPP-45)
Updated Card API to be signed by middleware

## 0.7.8 - 08-aug-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- [MAPP-43](https://netxd.atlassian.net/browse/MAPP-43)
Removed insert statement from DB script

## 0.7.7 - 08-aug-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-43](https://netxd.atlassian.net/browse/MAPP-43)
Updated API key value for ledger request in all env specific config

## 0.7.6 - 08-aug-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- [MAPP-43](https://netxd.atlassian.net/browse/MAPP-43)
Card config management API
1. Added DB script to create config table
2. Created handler to get all configuration from DB
3. If there is no data in DB send empty array in response
4. If configuration data exists in DB and type is 'JSON' then unmarshal config value and return id, configname, type, value in response. If it is string or int then return value as it is in response 
5. Created config response struct in code

## 0.7.5 - 08-aug-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Updated auth flow to allow multiple token for async calling

## 0.7.4 - 07-aug-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
1. Updated timeout in config-test.yml
2. Logged ledger URL in logger

## 0.7.3 - 03-aug-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- [MAPP-52](https://netxd.atlassian.net/browse/MAPP-52)
Changed FundFlowToday API
1. Updated fromDate to one day before current_date and toDate to one day after current_date
2. Added condition if timezone value is set then only change the current_time as per the specified timezone

## 0.7.2 - 02-aug-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- [MAPP-52](https://netxd.atlassian.net/browse/MAPP-52)
Updated data type of InstructedAmount from int to float64

## 0.7.1 - 02-aug-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Fixed
- [MAPP-07](https://netxd.atlassian.net/browse/MAPP-7)
Fixed issue related to duplicate bookmark entry

## 0.7.0 - 02-aug-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Fixed
- [MAPP-15](https://netxd.atlassian.net/browse/MAPP-15)
Created Card Management API
1. Get request body and call appropriate ledger API to get response
2. Added card endpoint in all env specific config
-----------------------------------------------------------------------------

## 0.6.1 - 01-aug-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- [MAPP-52](https://netxd.atlassian.net/browse/MAPP-52)
Updated FundFlowToday API - updated Param name from account_number to accNo

## 0.6.0  - 01-aug-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- [MAPP-52](https://netxd.atlassian.net/browse/MAPP-52)
Added FundFlowToday API 
1. Created request instance by setting Method, Id and Params fields and call ledger API to get response
2. Display amount credited and amount debited for provided account number for current_timestamp
-----------------------------------------------------------------------------

## 0.5.1 - 27-jul-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Changed
- [MAPP-11](https://netxd.atlassian.net/browse/MAPP-11)
Replaced fmt statements by log

## 0.5.0 - 26-jul-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
- [MAPP-11](https://netxd.atlassian.net/browse/MAPP-11)
Added Connection pooling 
1. Add connection pooling parameters in all config files(config-dev, config-prod, config-stage, config-test)
2. Set this parameters during DB connection to achieve connection pooling
3. Open DB connection in main.go instead of opening it in each handler function
-----------------------------------------------------------------------------

## 0.4.7- 26-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Update configs
1. Updated server port in config-prod.yml and config-stage.yml
2. Updated ledger endpoint in config-test.yml

## 0.4.6 - 25-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
1. Update logs - Added log message in each handler function 
2. Update configs - Updated token expiry time from 15 minute to 2 minute on DEV

## 0.4.5 - 24-jul-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
 1. Refactored Middleware chain in main.go
 2. Refactored Admin validation middleware - Used GetClaimsFromToken() function to extract claims

## 0.4.4 - 24-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Update logs- Added log message in each handler function 

## 0.4.3 - 21-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Updated main.go
1. Changed http method from GET to POST for AddAccountBookmark
2. Changed http method from POST to PUT for UnblockCustomer
3. Changed http method from PUT to PATCH for UnblockCustomer
4. Sent CurrentLogInDT in login response

## 0.4.2 - 21-jul-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Created admin handler
1. Segragated groups for admin and post login api
2. Added constant for user type
3. Updated db script to add 'UserType' column in User table
4. Removed admin related methods from authentication handler to admin handler
5. Added method for admin validation

## 0.4.1 - 21-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Fixed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Updated login response structure to handle issue if lastLoginDT is zero

## 0.4.0 - 21-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-07](https://netxd.atlassian.net/browse/MAPP-7)
CRUD for acocunt bookmarks
1. Added db script to create 'customer_account_bookmark' table
2. Added method to get claim and customer number from token
3. AddAccountBookmark API - Get customer number from token and add account related details of customer in db
4. DeleteAccountBookmark API - Find user in db and delete bookmark for that user
5. GetAllAccountBookmark API - Find user in db based on provided customer number and fetch all bookmarks
-----------------------------------------------------------------------------

## 0.3.12 - 21-jul-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Added Api to unblock user
1. Find user in db based on provided customer no/mobile/email
2. Set status as ACTIVE, loginRetryCount as 0, UpdatedBy as ADMIN if user exists and blocked in db

## 0.3.11 - 21-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Changed endpoint for sandbox in test config - pointing to "http://43.227.21.34:4043/pl/jsonrpc"

## 0.3.10 - 20-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Fixed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Fixed blocked user error msg while login

## 0.3.9 - 19-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Changed endpoint for sandbox in dev config

## 0.3.8 - 19-jul-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Changed package from 'dgrijalva' to 'golang-jwt'

## 0.3.7 - 19-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Changes in config files
1. Changed host and password for test server

## 0.3.6 - 18-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Token validation
1. Removed token from logout response header
2. Code refactoring

## 0.3.5 - 18-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Fixed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Issue with token string - Added code to get token string without bearer

## 0.3.4 - 18-jul-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
JWT Validation
1. Read token expiry time from config
2. Updated token expiration time and send updated token in response in TokenValidation method
3. Handled error if token is not generated
4. Added code to update db if user login first time and current_login time is not set

## 0.3.3 - 18-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Updated config for public port of dev server

## 0.3.2 - 18-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Fixed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Handled issues related to reading private key and public port of dev
1. Fixed issue related to private key - Issue occurred while reading private key from private.pem. Updated the file path of private.pem in code
2. Logged CustomerNo, Log in and log out time in log file

## 0.3.1 - 17-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Changed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Single Sign On Implementation
1. Added constants for user status
2. AuthorizationHandler - Replaced error messages with constant added in constant/error.go
3. Modified config-dev.yml file with valid dev db host and password

## 0.3.0 - 17-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Single Sign On Implementation
1. Added logic to block user after max login retry count
2. Added maxLoginRetryCount and token timeout parameters in all env specific configs
3. Added db script to add 'loginRetryCount' column in User table
-----------------------------------------------------------------------------

## 0.2.1 - 17-jul-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Modified code to manage JWT validation
1. Added error constant in error.go
2. Set expiry time for token in log out API
3. Restrict user for login if status is BLOCKED
4. Login API - Return error response 'ACTIVE_SESSION_ONGOING_ERROR' if user is already logged in
5. Handle token creation and expiry time while generating and validating the token

## 0.2.0 - 14-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Exposed end point for device registration and Manage onboarding
1. Added methods to call ledger API, sign payload with middleware key
2. Manage customer - Get request data and send it to ledger API to get response
3. Device registration -  Get data from request and sign it with middleware key, call ledger API to get response
4. Manage onboarding - Get data from request and sign it with middleware key, call ledger API to get response
-----------------------------------------------------------------------------

## 0.1.9 - 13-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Updated error code- UNAUTHORIZED_ERROR
Added End point for prelogin API's

## 0.1.8 - 13-jul-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Fixed
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Fixed Logout API issue- Unable to update isLoggedIn flag to false in DB. Replaced struct by map to fix this issue.

## 0.1.7 - 13-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Removed env file and added env specific config files

## 0.1.6 - 13-jul-2023 - from [@PoonamHajare](https://github.com/PoonamHajare)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Added Logout API
1. Extract token from request header and fetch customer number
2. If customer number exists in DB set isLoggedIn to false and lastLogOutDT to current_timestamp

## 0.1.5 - 13-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Added golangci.yaml file and added gitignore file

## 0.1.4 - 13-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
 Updated logger - Added error messages in log file

## 0.1.3 - 12-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
 Added ledger caller with signing logic

## 0.1.2 - 12-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Modified code to standardize config readings
1. Read server port from config file in main.go
2. Added secret key in constant file

## 0.1.1 - 12-jul-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Added basic project structure
1. Added structure for login request and response, error response
2. Added constant to project
3. Added script to create user table
4. Added structure for JWTClaims

## 0.1.0 - 12-jul-2023 - from [@PriyankaArerao](https://github.com/areraopriyanka)
### Added
- [MAPP-04](https://netxd.atlassian.net/browse/MAPP-4)
Added Login feature
1. Added code for Database connectivity
2. Find if user exists in Database based on provided credentials
3. If user exists then generate token 
4. Created token_generation.go file and added method to generate token
5. Added middleware function for token validation
6. Updated user details in database post successful login
7. Created log file under pkg/logging
-----------------------------------------------------------------------------

## 0.0.2 - 10-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- Added basic project structure for middleware
- Updated env file path in config.go

## 0.0.1 - 07-jul-2023 - from [@NishaChavan](https://github.com/NishaChavan)
### Added
- Added README.md file
-----------------------------------------------------------------------------

### Added
- Use Added for new features.

### Changed
- Use Changed for changes in existing functionality.

### Deprecated
- Use Deprecated for soon-to-be removed features.

### Removed
- Use Removed for now removed features.

### Security
- Use Security in case of vulnerabilities.

### Fixed
- Use Fixed for any bug fixes.









### Added
- Use Added for new features.

### Changed
- Use Changed for changes in existing functionality.

### Deprecated
- Use Deprecated for soon-to-be removed features.

### Removed
- Use Removed for now removed features.

### Security
- Use Security in case of vulnerabilities.

### Fixed
- Use Fixed for any bug fixes.