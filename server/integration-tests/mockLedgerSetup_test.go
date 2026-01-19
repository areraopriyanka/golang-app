package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"process-api/pkg/config"
	"process-api/pkg/ledger"
)

func SetupMockForLedger(suite *IntegrationTestSuite) *httptest.Server {
	config.Config.Ledger.PrivateKey = "-----BEGIN PRIVATE KEY-----\nMIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgcDHoj2lIAyDg4++F78rdf01Rc9nyluRXNBYEwr0tu9ehRANCAATTc9wQB5fPi/mgnsKHal6V9n5OrtQ1HhmnpGgTwHg46UD68c2fUG71cP28luTEWv1QhZLX7NCxiEsaTd94hdp3\n-----END PRIVATE KEY-----"
	config.Config.Ledger.CardsPublicKey = "-----BEGIN PUBLIC KEY-----\nMIGeMA0GCSqGSIb3DQEBAQUAA4GMADCBiAKBgEgiYAXZyoZzdUhqkCrwNJxBtbPBoEVaOGSCko+IkDCR93UJzuzBIv3286IVM7xXUEpmIj9MKnebY5CgKb9hAv6kt1clhuNpPYWYRHU/uq/PH31fYL6yf/e7bG4YoAHu1Ov212oqjgejerbTZVyeel3AKPdVP9mGu4sqmXLa+QQXAgMBAAE=\n-----END PUBLIC KEY-----"
	// Sets up mock ledger to specify response to CallLedgerAPIWithUrlAndGetRawResponse
	mockLedgerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jsonrpc" && r.URL.Path != "/cardv2" && r.URL.Path != "/paymentv2" {
			suite.T().Errorf("Expected to request '/jsonrpc', '/cardv2', or '/paymentv2', got: %s", r.URL.Path)
		}

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			suite.Require().NoError(err, "Failed to read request body")
		}

		var request ledger.Request
		err = json.Unmarshal(payload, &request)
		if err != nil {
			suite.Require().NoError(err, "Failed to unmarshal request body")
		}

		var requestPayload map[string]interface{}
		err = json.Unmarshal(request.Params.Payload, &requestPayload)
		if err != nil {
			suite.Require().NoError(err, "Failed to unmarshal request payload")
		}

		var responseBody string
		switch request.Method {
		case ("CustomerService.AddCustomer"):
			suite.Assert().NotEmpty(requestPayload["DOB"], "DOB")
			suite.Assert().NotEmpty(requestPayload["UserName"], "UserName")
			suite.Assert().NotEmpty(requestPayload["address"], "address")
			suite.Assert().NotEmpty(requestPayload["contact"], "contact")
			suite.Assert().NotEmpty(requestPayload["firstName"], "firstName")
			suite.Assert().NotEmpty(requestPayload["identification"], "identification")
			suite.Assert().NotEmpty(requestPayload["lastName"], "lastName")
			suite.Assert().NotEmpty(requestPayload["password"], "password")
			suite.Assert().NotEmpty(requestPayload["title"], "title")
			suite.Assert().NotEmpty(requestPayload["type"], "type")
			responseBody = `{
					"id": "1",
					"result": {
						"Id": "13346",
						"CustomerNumber": "100000000052004",
						"Status": "ACTIVE"
					},
					"error": null
				}`
		case ("CustomerService.AddUserKey"):
			suite.Assert().NotEmpty(requestPayload["publicKey"], "publicKey")
			suite.Assert().NotEmpty(requestPayload["status"], "status")
			suite.Assert().NotEmpty(requestPayload["userName"], "userName")
			responseBody = `{
					"result": {
						"KeyID": "ledger_provided_key_id",
						"status": "ACTIVE",
						"apiKey": "e6b45dead7e946f9b25fe319106b3312"
					},
					"error": null
				}`

		case ("CustomerService.AddAccount"):
			suite.Assert().NotEmpty(requestPayload["AccountCategory"], "AccountCategory")
			suite.Assert().NotEmpty(requestPayload["accountType"], "accountType")
			suite.Assert().NotEmpty(requestPayload["currency"], "currency")
			suite.Assert().NotEmpty(requestPayload["customerID"], "customerID")
			suite.Assert().NotEmpty(requestPayload["name"], "name")
			responseBody = `{
					"id": "1",
					"result": {
						"ID": "3173041",
						"status": "ACTIVE",
						"accountNumber": "200982345048071",
						"accountType": "WALLET",
						"institutionID": "101115399",
						"customerID": "100000000052004"
					},
					"error": null
				}`

		case ("CustomerService.GetCustomer"):
			customerDoesNotExist := `{
					"id": "1",
					"error": {
						"code": "NOT_FOUND_CUSTOMER",
						"message": "Customer does not exist"
					},
					"jsonrpc": "2.0"
				}`
			customerExists := `{
					"id": "1",
					"jsonrpc": "2.0",
					"result": {
						"ID": "10000000001",
						"type": "INDIVIDUAL",
						"businessNameLegal": "Test User",
						"DOB": "20000805",
						"title": "Ms",
						"firstName": "Test",
						"lastName": "User",
						"status": "ACTIVE",
						"createdDate": "2025-02-21T12:47:01.61Z",
						"updatedDate": "2025-02-21T12:47:01.61Z",
						"identification": [
							{
								"type": "SSN",
								"value": "*****6668"
							}
						],
						"contact": {
							"email": "**@gmail.com",
							"phoneNumber": "******5222"
						},
						"address": {
							"addressLine1": "123 A St.",
							"city": "Providence",
							"state": "RI",
							"country": "US",
							"zip": "00000"
						},
						"accounts": [
							{
								"id": "217140",
								"name": "Default DreamFi Account",
								"nickName": "Default DreamFi Account",
								"number": "987546218371925",
								"customerID": "10000000001",
								"institutionName": "XD Legder",
								"accountCategory": "DreamFi_MVP",
								"accountType": "CHECKING",
								"currency": "USD",
								"status": "ACTIVE",
								"institutionID": "124303298",
								"glAccount": "20006445021",
								"balance": 10000,
								"holdBalance": 0,
								"ledgerBalance": 0,
								"preAuthBalance": 0,
								"accrualBalance": 0,
								"createdDate": "2025-02-21T12:47:02.118Z",
								"updatedDate": "2025-02-21T12:47:02.118Z"
							}
						]
					}
				}`
			customerWithNoBalances := `{
					"id": "1",
					"jsonrpc": "2.0",
					"result": {
						"ID": "10000000001",
						"type": "INDIVIDUAL",
						"businessNameLegal": "Test User",
						"DOB": "20000805",
						"title": "Ms",
						"firstName": "Test",
						"lastName": "User",
						"status": "ACTIVE",
						"createdDate": "2025-02-21T12:47:01.61Z",
						"updatedDate": "2025-02-21T12:47:01.61Z",
						"identification": [
							{
								"type": "SSN",
								"value": "*****6668"
							}
						],
						"contact": {
							"email": "**@gmail.com",
							"phoneNumber": "******5222"
						},
						"address": {
							"addressLine1": "123 A St.",
							"city": "Providence",
							"state": "RI",
							"country": "US",
							"zip": "00000"
						},
						"accounts": [
							{
								"id": "217140",
								"name": "Default DreamFi Account",
								"nickName": "Default DreamFi Account",
								"number": "6E6F62616C616E636573",
								"customerID": "10000000001",
								"institutionName": "XD Legder",
								"accountCategory": "DreamFi_MVP",
								"accountType": "CHECKING",
								"currency": "USD",
								"status": "ACTIVE",
								"institutionID": "124303298",
								"glAccount": "20006445021",
								"balance": 0,
								"holdBalance": 0,
								"ledgerBalance": 0,
								"preAuthBalance": 0,
								"accrualBalance": 0,
								"createdDate": "2025-02-21T12:47:02.118Z",
								"updatedDate": "2025-02-21T12:47:02.118Z"
							}
						]
					}
				}`
			// Mock response to check duplicate ssn or email
			if _, exists := requestPayload["Identification"]; exists {
				suite.Assert().NotEmpty(requestPayload["Identification"], "identification")
				responseBody = customerDoesNotExist
			} else if contact, exists := requestPayload["contact"].(map[string]any); exists {
				suite.Assert().NotEmpty(contact, "contact record must exist")
				if contact["email"] == "existing-ledger@gmail.com" {
					responseBody = customerExists
				} else {
					responseBody = customerDoesNotExist
				}
			} else if _, exists := requestPayload["customerNumber"]; exists {
				// Mock response to get customer data using customerNumber
				suite.Assert().NotEmpty(requestPayload["customerNumber"], "CustomerNumber")
				if requestPayload["customerNumber"] == "6E6F62616C616E636573" {
					responseBody = customerWithNoBalances
				} else {
					responseBody = customerExists
				}
			}
		case ("CustomerService.UpdateCustomer"):
			suite.Assert().NotEmpty(requestPayload["ID"], "CustomerId")
			suite.Assert().NotEmpty(requestPayload["firstName"], "firstName")
			suite.Assert().NotEmpty(requestPayload["lastName"], "lastName")
			responseBody = `{
					"id": "1",
					"result": {
						"ID": "100000000052004",
						"type": "INDIVIDUAL",
						"businessNameLegal": "Test Bar",
						"identification": [
						{
							"type": "SSN",
							"value": "*****6789"
						}
						],
						"contact": {
							"email": "te******r@gmail.com",
							"phoneNumber": "********1234"
						},
						"address": {
							"addressLine1": "123 Main St",
							"city": "Adak",
							"state": "AK",
							"country": "US",
							"zip": "11111"
						},
						"DOB": "20001109",
						"title": "",
						"firstName": "Test",
						"lastName": "Bar",
						"createdDate": "2025-08-29T00:07:57.377Z",
						"updatedDate": "2025-08-29T00:42:46.695Z",
						"status": "ACTIVE"
					},
					"jsonrpc": "2.0"
				}`
		case ("CustomerService.UpdateCustomerSettings"):
			suite.Assert().NotEmpty(requestPayload["customerId"], "CustomerId")
			suite.Assert().NotEmpty(requestPayload["pciCheck"], "pciCheck")
			responseBody = `{
					"id": "1",
					"result": {
						"message": "Update Customer Settings  Successfully"
					}
				}`

		case ("AccountService.GetAccount"):
			responseBody = `{
				"id": "1",
				"result": {
					"account": {
						"id": "217140",
						"name": "Default DreamFi Account",
						"number": "987546218371925",
						"nickName": "Default DreamFi Account",
						"createdDate": "2025-07-03T21:08:46.512Z",
						"updatedDate": "2025-07-03T21:08:46.512Z",
						"balance": 10000,
						"debit": false,
						"minimumBalance": 0,
						"holdBalance": 0,
						"subLedgerCode": "SL_200",
						"tags": [
							"AccountLevel: DreamFi_MVP"
						],
						"final": true,
						"parent": {
							"ID": "1001",
							"code": "",
							"name": "DreamFi Customer Activity Pool",
							"number": "200064450217486"
						},
						"customerID": "10000000001",
						"customerName": "Test User",
						"institutionName": "XD Legder",
						"accountCategory": "LIABILITY",
						"accountType": "CHECKING",
						"currency": "USD",
						"currencyCode": "840",
						"legalReps": [
							{
								"ID": "26438022",
								"name": "Test User",
								"createdDate": "0001-01-01T00:00:00Z",
								"updatedDate": "0001-01-01T00:00:00Z"
							}
						],
						"status": "ACTIVE",
						"institutionID": "124303298",
						"glAccount": "20006445021",
						"DDAAccount": true,
						"isVerify": true,
						"minimumRouteApprovers": 0,
						"newRouteAlert": false,
						"ceTransactionNumber": "PL26438024",
						"accountLevel": "DreamFi_MVP",
						"ledgerBalance": 10000,
						"preAuthBalance": 0,
						"accountFinderSync": false,
						"isGLVerify": false,
						"isShadowAccount": false,
						"sweep": false,
						"isClosed": false,
						"pendingCredits": 0,
						"pendingDebits": 0,
						"lienBalance": 0,
						"program": "",
						"productID": "",
						"externalLedger": false
					}
				},
				"jsonrpc": "2.0"
			}`

		case ("AccountService.UpdateAccountStatus"):
			suite.Assert().NotEmpty(requestPayload["accountNumber"], "accountNumber")
			suite.Assert().NotEmpty(requestPayload["status"], "status")

			switch requestPayload["status"] {
			case "CLOSED":
				responseBody = `{
						"id": "1",
						"result": {
							"CustomerId": "100000000006001",
							"AccountNumber": "400320588344662",
							"InstitutionId": "101115315",
							"Name": "General Account",
							"Status": "CLOSED"
						}
					}`
			case "SUSPENDED":
				responseBody = `{
						"id": "1",
						"result": {
							"CustomerId": "100000000006001",
							"AccountNumber": "400320588344662",
							"InstitutionId": "101115315",
							"Name": "General Account",
							"Status": "SUSPENDED"
						}
					}`
			}

		case ("ledger.CARD.request"):
			suite.Assert().NotEmpty(requestPayload["channel"], "channel")
			suite.Assert().NotEmpty(requestPayload["product"], "product")
			suite.Assert().NotEmpty(requestPayload["program"], "program")
			suite.Assert().NotEmpty(requestPayload["transactionType"], "transactionType")

			switch requestPayload["transactionType"] {
			case "ADD_CARD":
				suite.Assert().NotEmpty(requestPayload["customerId"], "customerId")
				suite.Assert().NotEmpty(requestPayload["accountNumber"], "accountNumber")
				suite.Assert().NotEmpty(requestPayload["reference"], "reference")
				responseBody = `{
						"id": "1",
						"result": {
							"card": {
								"cardId": "d3360652aec34493976fa0d24b9d098d",
								"cardType": "PHYSICAL",
								"postedDate": "2025-02-25T13:49:00.403178435Z",
								"updatedDate": "2025-02-25T13:49:00.403178548Z",
								"cardMaskNumber": "************2083",
								"cardStatus": "CARD_IS_NOT_ACTIVATED",
								"cardExpiryDate": "202802",
								"allowAtm": false,
								"allowEcommerce": false,
								"allowMoto": false,
								"allowPos": false,
								"allowTips": false,
								"allowPurchase": false,
								"allowRefund": false,
								"allowCashback": false,
								"allowWithdraw": false,
								"allowAuthAndCompletion": false,
								"smart": false,
								"checkAvsZip": false,
								"checkAvsAddr": false,
								"cvv": "",
								"transactionMade": false,
								"orderStatus": "ORDER_PLACED",
								"orderId": "4UPHU2HB4YU3000",
								"isReIssue": false,
								"isReplace": false,
								"orderSubStatus": "ORDER_PENDING"
							},
							"api": {
								"type": "ADD_CARD_ACK",
								"reference": "REFvisadpsa77dc311-1fd8-4169-a56c-cdd0046c0c2b",
								"dateCreated": 1740491340,
								"originalReference": "visadpsa77dc311-1fd8-4169-a56c-cdd0046c0c2b"
							}
						}
					}`
			case "GET_CARD_DETAILS":
				suite.Assert().NotEmpty(requestPayload["customerId"], "customerId")
				suite.Assert().NotEmpty(requestPayload["accountNumber"], "accountNumber")
				suite.Assert().NotEmpty(requestPayload["reference"], "reference")
				suite.Assert().NotEmpty(requestPayload["cardId"], "cardId")
				switch requestPayload["cardId"] {
				case "frozen":
					responseBody = `{
						"id": "2",
						"result": {
							"api": {
								"type": "GET_CARD_DETAILS_ACK",
								"reference": "REFgetcarddetails_1000000000266_1747650389939272600",
								"dateCreated": 1747650733,
								"originalReference": "getcarddetails_1000000000266_1747650389939272600"
							},
							"cardDetails": {
								"cardId": "701f5958c14d46549a11390340763ace",
								"cardProduct": "DreamFi_Employee_Consumer_Debit",
								"createdDate": "2025-05-19T10:10:04.290Z",
								"updatedDate": "2025-05-19T10:10:04.297Z",
								"cardMaskNumber": "************7815",
								"cardStatus": "TEMPRORY_BLOCKED_BY_CLIENT",
								"orderStatus": "ORDER_PLACED",
								"orderId": "4UQ77HMSM18D000",
								"isReIssue": false,
								"isReplace": false,
								"externalCardId": "v-402-1f782908-428f-4c36-9ee2-a752894c41ac",
								"orderSubStatus": "ORDER_PENDING",
								"cardExpiryDate": "202805"
							}
						}
					}`
				case "report_lost_stolen":
					responseBody = `{
						"id": "2",
						"result": {
							"api": {
								"type": "GET_CARD_DETAILS_ACK",
								"reference": "REFgetcarddetails_1000000000266_1747650389939272600",
								"dateCreated": 1747650733,
								"originalReference": "getcarddetails_1000000000266_1747650389939272600"
							},
							"cardDetails": {
								"cardId": "701f5958c14d46549a11390340763ace",
								"cardProduct": "DreamFi_Employee_Consumer_Debit",
								"createdDate": "2025-05-19T10:10:04.290Z",
								"updatedDate": "2025-05-19T10:10:04.297Z",
								"cardMaskNumber": "************7815",
								"cardStatus": "ACTIVATED",
								"orderStatus": "ORDER_PLACED",
								"orderId": "4UQ77HMSM18D000",
								"isReIssue": false,
								"isReplace": false,
								"externalCardId": "v-402-1f782908-428f-4c36-9ee2-a752894c41ac",
								"orderSubStatus": "ORDER_PENDING",
								"cardExpiryDate": "202805"
							}
						}
					}`
				default:
					responseBody = `{
							"id": "1",
							"result": {
								"api": {
									"type": "GET_CARD_DETAILS_ACK",
									"reference": "REFgetcarddetails_1000000000266_1747650389939272600",
									"dateCreated": 1747650733,
									"originalReference": "getcarddetails_1000000000266_1747650389939272600"
								},
								"cardDetails": {
									"cardId": "701f5958c14d46549a11390340763ace",
									"cardProduct": "DreamFi_Employee_Consumer_Debit",
									"createdDate": "2025-05-19T10:10:04.290Z",
									"updatedDate": "2025-05-19T10:10:04.297Z",
									"cardMaskNumber": "************7815",
									"cardStatus": "CARD_IS_NOT_ACTIVATED",
									"orderStatus": "ORDER_PLACED",
									"orderId": "4UQ77HMSM18D000",
									"isReIssue": false,
									"isReplace": false,
									"externalCardId": "v-402-1f782908-428f-4c36-9ee2-a752894c41ac",
									"orderSubStatus": "ORDER_PENDING",
									"cardExpiryDate": "202805"
								}
							}
						}`
				}
			case "ADD_CARD_HOLDER_WITH_PRIMARY_ADDRESS":
				suite.Assert().NotEmpty(requestPayload["customerId"], "customerId")
				suite.Assert().NotEmpty(requestPayload["accountNumber"], "accountNumber")
				suite.Assert().NotEmpty(requestPayload["reference"], "reference")
				responseBody = `{
						"id": "1",
						"result": {
							"cardHolderId": "CH00000000009004",
							"api": {
							"type": "ADD_CARD_HOLDER_WITH_PRIMARY_ADDRESS_ACK",
							"reference": "REFe626fde4-5045-4f61-b8d9-d763682e5c62",
							"dateCreated": 1740491339,
							"originalReference": "e626fde4-5045-4f61-b8d9-d763682e5c62"
							}
						}
					}`

			case "GET_CARD":
				suite.Assert().NotEmpty(requestPayload["customerId"], "customerId")
				suite.Assert().NotEmpty(requestPayload["accountNumber"], "accountNumber")
				suite.Assert().NotEmpty(requestPayload["reference"], "reference")
				suite.Assert().NotEmpty(requestPayload["cardId"], "cardId")
				responseBody = `{
						"id": "1",
						"result": {
							"cardHolder": {
								"cardHolderId": "CH00000000020003",
								"firstName": "Matthew",
								"lastName": "Sandra",
								"phoneNumber": "2562322420",
								"addressLine1": "3456W",
								"addressLine2": "1st street",
								"city": "Kansus",
								"state": "KS",
								"zipCode": "56213",
								"country": "US",
								"emailId": "mathewsandra+1@gmail.com",
								"createdDate": "2024-11-20T06:51:22.476Z",
								"updatedDate": "2024-11-20T06:51:22.476Z"
							},
							"card": {
								"cardId": "6f586be7bf1c44b8b4ea11b2e2510e25",
								"cardProduct": "2c6b841a-dfc1-4a6a-a12a-13c21035b5be",
								"postedDate": "2024-11-20T06:54:26.822Z",
								"updatedDate": "2024-11-20T06:59:13.805Z",
								"cardMaskNumber": "************5461",
								"cardStatus": "ACTIVE",
								"cardExpiryDate": "202611",
								"allowAtm": false,
								"allowEcommerce": false,
								"allowMoto": false,
								"allowPos": false,
								"allowTips": false,
								"allowPurchase": false,
								"allowRefund": false,
								"allowCashback": false,
								"allowWithdraw": false,
								"allowAuthAndCompletion": false,
								"smart": false,
								"checkAvsZip": false,
								"checkAvsAddr": false,
								"cvv": "",
								"transactionMade": false,
								"orderStatus": "ORDER_PLACED",
								"isReIssue": false,
								"isReplace": false,
								"orderSubStatus": "ORDER_PENDING"
							},
							"api": {
								"type": "GET_CARD_ACK",
								"reference": "REFvisadps100010",
								"dateCreated": 1732086223,
								"originalReference": "visadps100010"
							}
						}
					}`
			case "VALIDATE_CVV":
				suite.Assert().NotEmpty(requestPayload["cardId"], "cardId")
				suite.Assert().NotEmpty(requestPayload["cardCvv"], "cardCvv")
				responseBody = `{
							"id": "1",
							"result": {
									"cardId": "6f586be7bf1c44b8b4ea11b2e2510e25",
									"api": {
									"type": "VALIDATE_CVV_ACK",
									"reference": "REF",
									"dateCreated": 1732087371
									},
									"message": "CVV has been validated successfully"
								}
							}`
			case "PIN_CHANGE":
				suite.Assert().NotEmpty(requestPayload["customerId"], "customerId")
				suite.Assert().NotEmpty(requestPayload["accountNumber"], "accountNumber")
				suite.Assert().NotEmpty(requestPayload["reference"], "reference")
				suite.Assert().NotEmpty(requestPayload["cardId"], "cardId")
				suite.Assert().NotEmpty(requestPayload["newPIN"], "newPIN")
				responseBody = `{
						"id": "1",
						"result": {
							"api": {
								"type": "PIN_CHANGE_ACK",
								"reference": "REFvisadps100008",
								"dateCreated": 1732086021,
								"originalReference": "visadps100008"
							}
						}
					}`
			case "UPDATE_STATUS":
				suite.Assert().NotEmpty(requestPayload["customerId"], "customerId")
				suite.Assert().NotEmpty(requestPayload["accountNumber"], "accountNumber")
				suite.Assert().NotEmpty(requestPayload["reference"], "reference")
				suite.Assert().NotEmpty(requestPayload["cardId"], "cardId")
				suite.Assert().NotEmpty(requestPayload["statusAction"], "statusAction")
				switch requestPayload["statusAction"] {
				case "ACTIVATE":
					suite.Assert().NotEmpty(requestPayload["cardCvv"], "cardCvv")
					responseBody = `{
						"id": "1",
						"result": {
							"card": {
								"cardId": "6f586be7bf1c44b8b4ea11b2e2510e25",
								"postedDate": "2024-11-20T07:28:31.801Z",
								"updatedDate": "2024-11-20T07:30:48.584Z",
								"cardStatus": "ACTIVATED",
								"allowAtm": false,
								"allowEcommerce": false,
								"allowMoto": false,
								"allowPos": false,
								"allowTips": false,
								"allowPurchase": false,
								"allowRefund": false,
								"allowCashback": false,
								"allowWithdraw": false,
								"allowAuthAndCompletion": false,
								"smart": false,
								"checkAvsZip": false,
								"checkAvsAddr": false,
								"cvv": "",
								"transactionMade": false,
								"isReIssue": false,
								"isReplace": false
							},
							"api": {
								"type": "UPDATE_STATUS_ACK",
								"reference": "REFvisadps100019",
								"dateCreated": 1732087851,
								"originalReference": "visadps100019"
							}
						}
					}`
				case "UNLOCK":
					responseBody = `{
						"id": "1",
						"result": {
							"card": {
								"cardId": "6f586be7bf1c44b8b4ea11b2e2510e25",
								"postedDate": "2024-11-20T07:28:31.801Z",
								"updatedDate": "2024-11-20T07:30:48.584Z",
								"cardStatus": "ACTIVATED",
								"allowAtm": false,
								"allowEcommerce": false,
								"allowMoto": false,
								"allowPos": false,
								"allowTips": false,
								"allowPurchase": false,
								"allowRefund": false,
								"allowCashback": false,
								"allowWithdraw": false,
								"allowAuthAndCompletion": false,
								"smart": false,
								"checkAvsZip": false,
								"checkAvsAddr": false,
								"cvv": "",
								"transactionMade": false,
								"isReIssue": false,
								"isReplace": false
							},
							"api": {
								"type": "UPDATE_STATUS_ACK",
								"reference": "REFvisadps100019",
								"dateCreated": 1732087851,
								"originalReference": "visadps100019"
							}
						}
					}`
				case "LOCK":
					responseBody = `{
						"id": "1",
						"result": {
							"card": {
							"cardId": "8e7992ec6d544142a8dd19e096e40a1d",
							"postedDate": "2025-03-24T14:59:04.263Z",
							"updatedDate": "2025-03-24T14:59:52.913Z",
							"cardStatus": "TEMPRORY_BLOCKED_BY_CLIENT",
							"allowAtm": false,
							"allowEcommerce": false,
							"allowMoto": false,
							"allowPos": false,
							"allowTips": false,
							"allowPurchase": false,
							"allowRefund": false,
							"allowCashback": false,
							"allowWithdraw": false,
							"allowAuthAndCompletion": false,
							"smart": false,
							"checkAvsZip": false,
							"checkAvsAddr": false,
							"cvv": "",
							"transactionMade": false,
							"isReIssue": false,
							"isReplace": false
							},
							"api": {
							"type": "UPDATE_STATUS_ACK",
							"reference": "REFupdatestatus_100000000021172_1742828390320036500",
							"dateCreated": 1742828393,
							"originalReference": "updatestatus_100000000021172_1742828390320036500"
							}
						}
					}`

				case "REPORT_LOST_STOLEN":
					responseBody = `{
						"id": "1",
						"result": {
							"card": {
								"cardId": "cbc8843c5b2c493699da93913e0c3fc8",
								"postedDate": "2025-03-19T14:36:13.375Z",
								"updatedDate": "2025-03-19T14:40:13.417Z",
								"cardStatus": "LOST_STOLEN",
								"allowAtm": false,
								"allowEcommerce": false,
								"allowMoto": false,
								"allowPos": false,
								"allowTips": false,
								"allowPurchase": false,
								"allowRefund": false,
								"allowCashback": false,
								"allowWithdraw": false,
								"allowAuthAndCompletion": false,
								"smart": false,
								"checkAvsZip": false,
								"checkAvsAddr": false,
								"cvv": "",
								"transactionMade": false,
								"isReIssue": false,
								"isReplace": false
							},
							"api": {
								"type": "UPDATE_STATUS_ACK",
								"reference": "REFupdatestatus_100000000021133_1742395193886724900",
								"dateCreated": 1742395213,
								"originalReference": "updatestatus_100000000021133_1742395193886724900"
							}
						}
					}`
				case "CLOSE":
					responseBody = `{
						"id": "1",
						"result": {
							"card": {
								"cardId": "fd8c4acda46f4bb2874c60d520a10e83",
								"postedDate": "2025-08-01T22:00:21.997Z",
								"updatedDate": "2025-08-01T22:02:02.241Z",
								"cardStatus": "CLOSED",
								"allowAtm": false,
								"allowEcommerce": false,
								"allowMoto": false,
								"allowPos": false,
								"allowTips": false,
								"allowPurchase": false,
								"allowRefund": false,
								"allowCashback": false,
								"allowWithdraw": false,
								"allowAuthAndCompletion": false,
								"smart": false,
								"checkAvsZip": false,
								"checkAvsAddr": false,
								"cvv": "",
								"transactionMade": false,
								"isReIssue": false,
								"isReplace": false
							},
							"api": {
								"type": "UPDATE_STATUS_ACK",
								"reference": "REFupdatestatus_100000000034019_1754085722203092227",
								"dateCreated": 1754085722,
								"originalReference": "updatestatus_100000000034019_1754085722203092227"
							}
						}
					}`
				default:
					suite.Require().Failf("Unknown statusAction in request: %s", request.Method)
				}
			case "REPLACE_CARD":
				suite.Assert().NotEmpty(requestPayload["customerId"], "customerId")
				suite.Assert().NotEmpty(requestPayload["accountNumber"], "accountNumber")
				suite.Assert().NotEmpty(requestPayload["reference"], "reference")
				suite.Assert().NotEmpty(requestPayload["cardId"], "cardId")
				suite.Assert().NotEmpty(requestPayload["statusAction"], "statusAction")
				isReplace := requestPayload["statusAction"] == ledger.REPLACE_CARD
				isReIssue := !isReplace
				responseBody = fmt.Sprintf(`{
						"id": "1",
						"result": {
							"card": {
								"cardId": "6f586be7bf1c44b8b4ea11b2e2510e25",
								"cardHolderId": "CH00000000020003",
								"cardHolderName": "Matthew Sandra",
								"cardProduct": "2c6b841a-dfc1-4a6a-a12a-13c21035b5be",
								"customerId": "100000000006001",
								"accountId": "332004",
								"product": "DEFAULT",
								"program": "DEFAULT",
								"cardType": "PHYSICAL",
								"postedDate": "2024-11-20T06:54:26.822Z",
								"updatedDate": "2024-11-20T07:25:56.513Z",
								"cardMaskNumber": "************5461",
								"cardNumber": "649f748bd576e66fdb4f2100bb79a91c",
								"cardStatus": "LOST_STOLEN",
								"cardExpiryDate": "202611",
								"allowAtm": true,
								"allowEcommerce": true,
								"allowMoto": true,
								"allowPos": true,
								"allowTips": true,
								"allowPurchase": true,
								"allowRefund": true,
								"allowCashback": true,
								"allowWithdraw": false,
								"allowAuthAndCompletion": false,
								"smart": true,
								"checkAvsZip": true,
								"checkAvsAddr": true,
								"cvv": "",
								"accountNumber": "400320588344662",
								"cardName": "Matthew Sandra",
								"patterns": [
									"CARDNUMBER:5461",
									"DATE:20112024",
									"CARDHOLDERNAME:MATTHEW SANDRA",
									"CARDNAME:MATTHEW SANDRA",
									"ACCOUNTNUMBER:400320588344662",
									"CUSTOMERID:100000000006001",
									"ACCOUNTNAME:DAVID SAVINGS ACCOUNT",
									"CUSTOMERNAME:DAVID TEST",
									"CARDSTATUS:CARD_IS_NOT_ACTIVATED"
								],
								"transactionMade": false,
								"orderStatus": "ORDER_PLACED",
								"orderId": "4UOO1ER8E6SV000",
								"network": "VISA_DPS",
								"isReIssue": %t,
								"isReplace": %t,
								"externalCardId": "v-401-27c8bce1-1178-4282-8541-ae8401e65d0e",
								"cardCreatedYear": "2024",
								"orderSubStatus": "ORDER_PENDING",
								"accountName": "David Savings Account",
								"customerName": "David Test"
							},
							"newCard": {
								"cardId": "b4ea11b2e2510e256f586be7bf1c44b8",
								"cardHolderId": "CH00000000020003",
								"cardHolderName": "Matthew Sandra",
								"cardProduct": "2c6b841a-dfc1-4a6a-a12a-13c21035b5be",
								"customerId": "100000000006001",
								"accountId": "332004",
								"product": "DEFAULT",
								"program": "DEFAULT",
								"cardType": "PHYSICAL",
								"postedDate": "2024-11-20T06:54:26.822Z",
								"updatedDate": "2024-11-20T07:25:56.513Z",
								"cardMaskNumber": "************5461",
								"cardNumber": "649f748bd576e66fdb4f2100bb79a91c",
								"cardStatus": "LOST_STOLEN",
								"cardExpiryDate": "202701",
								"allowAtm": true,
								"allowEcommerce": true,
								"allowMoto": true,
								"allowPos": true,
								"allowTips": true,
								"allowPurchase": true,
								"allowRefund": true,
								"allowCashback": true,
								"allowWithdraw": false,
								"allowAuthAndCompletion": false,
								"smart": true,
								"checkAvsZip": true,
								"checkAvsAddr": true,
								"cvv": "",
								"accountNumber": "400320588344662",
								"cardName": "Matthew Sandra",
								"patterns": [
									"CARDNUMBER:5461",
									"DATE:20112024",
									"CARDHOLDERNAME:MATTHEW SANDRA",
									"CARDNAME:MATTHEW SANDRA",
									"ACCOUNTNUMBER:400320588344662",
									"CUSTOMERID:100000000006001",
									"ACCOUNTNAME:DAVID SAVINGS ACCOUNT",
									"CUSTOMERNAME:DAVID TEST",
									"CARDSTATUS:CARD_IS_NOT_ACTIVATED"
								],
								"transactionMade": false,
								"orderStatus": "ORDER_PLACED",
								"orderId": "4UOO1ER8E6SV000",
								"network": "VISA_DPS",
								"isReIssue": false,
								"isReplace": false,
								"externalCardId": "v-401-27c8bce1-1178-4282-8541-ae8401e65d0e",
								"cardCreatedYear": "2024",
								"orderSubStatus": "ORDER_PENDING",
								"accountName": "David Savings Account",
								"customerName": "David Test"
							},
							"api": {
								"type": "REPLACE_CARD_ACK",
								"reference": "REFvisadps100017",
								"dateCreated": 1732087711,
								"originalReference": "visadps100017"
							}
						}
					}`, isReIssue, isReplace)
			case "GET_CARD_LIMIT":
				suite.Assert().NotEmpty(requestPayload["customerId"], "customerId")
				suite.Assert().NotEmpty(requestPayload["accountNumber"], "accountNumber")
				suite.Assert().NotEmpty(requestPayload["reference"], "reference")
				suite.Assert().NotEmpty(requestPayload["cardId"], "cardId")
				responseBody = `{
						"id": "1",
						"result": {
							"card": {
							"cardId": "ff32c9ded3774e109e002773e8822a0b",
							"postedDate": "2024-11-20T07:32:53.796Z",
							"updatedDate": "2024-11-20T07:37:46.326Z",
							"allowAtm": false,
							"allowEcommerce": false,
							"allowMoto": false,
							"allowPos": false,
							"allowTips": false,
							"allowPurchase": false,
							"allowRefund": false,
							"allowCashback": false,
							"allowWithdraw": false,
							"allowAuthAndCompletion": false,
							"smart": false,
							"checkAvsZip": false,
							"checkAvsAddr": false,
							"cvv": "",
							"limits": [
								{
								"type": "VOLUME",
								"value": "100000",
								"cycleLength": "0",
								"cycleType": "LIFE_TIME",
								"remaining": "100000"
								},
								{
								"type": "VOLUME",
								"value": "100000",
								"cycleLength": "0",
								"cycleType": "YEAR",
								"remaining": "100000"
								},
								{
								"type": "VOLUME",
								"value": "10000",
								"cycleLength": "0",
								"cycleType": "MONTH",
								"remaining": "10000"
								},
								{
								"type": "VOLUME",
								"value": "10000",
								"cycleLength": "0",
								"cycleType": "DAY",
								"remaining": "10000"
								},
								{
								"type": "VOLUME",
								"value": "10000",
								"cycleLength": "0",
								"cycleType": "PER_TRANSACTION",
								"remaining": "10000"
								},
								{
								"type": "COUNT",
								"value": "100",
								"cycleLength": "0",
								"cycleType": "LIFE_TIME",
								"remaining": "100"
								},
								{
								"type": "COUNT",
								"value": "100",
								"cycleLength": "0",
								"cycleType": "YEAR",
								"remaining": "100"
								},
								{
								"type": "COUNT",
								"value": "10",
								"cycleLength": "0",
								"cycleType": "MONTH",
								"remaining": "10"
								},
								{
								"type": "COUNT",
								"value": "1",
								"cycleLength": "0",
								"cycleType": "DAY",
								"remaining": "1"
								}
							],
							"transactionMade": false,
							"isReIssue": false,
							"isReplace": false
							},
							"api": {
							"type": "GET_CARD_LIMIT_ACK",
							"reference": "REFvisadps100036",
							"dateCreated": 1732090185,
							"originalReference": "visadps100036"
							}
						}
					}`

			default:
				suite.Require().Failf("Unknown method in request: %s", request.Method)
			}

		case ("ledger.ach.transfer"):
			suite.Assert().NotEmpty(requestPayload["channel"], "channel")
			suite.Assert().NotEmpty(requestPayload["transactionType"], "transactionType")

			switch requestPayload["transactionType"] {
			case "ACH_PULL":
				suite.Assert().NotEmpty(requestPayload["reference"], "reference")
				responseBody = `{
						"id": "1",
						"result": {
							"api": {
								"type": "ACH_PULL_ACK",
								"reference": "ledger.ach.transfer_ach_pull_1741988110548470029",
								"dateTime": "2025-03-14 21:35:10"
							},
							"account": {
								"accountId": "123456789012345",
								"balanceCents": 99999,
								"holdBalanceCents": 1,
								"status": "ACTIVE"
							},
							"transactionNumber": "QA00000000000000",
							"transactionStatus": "PENDING",
							"transactionAmountCents": 50000,
							"processId": "PL25031400013073"
						}
					}`

			case "ACH_OUT":
				suite.Assert().NotEmpty(requestPayload["reference"], "reference")
				responseBody = `{
						"id": "1",
						"result": {
							"api": {
								"type": "ACH_OUT_ACK",
								"reference": "ledger.ach.transfer_ach_out_1741989796955960346",
								"dateTime": "2025-03-14 22:03:17"
							},
							"account": {
								"accountId": "123456789012345",
								"balanceCents": 99999,
								"holdBalanceCents": 1,
								"status": "ACTIVE"
							},
							"transactionNumber": "QA00000000000001",
							"transactionStatus": "COMPLETED",
							"transactionAmountCents": 50000,
							"processId": "PL25031400013074"
						}
					}`

			default:
				suite.Require().Failf("Unknown method in request: %s", request.Method)
			}

		case ("ledger.transfer"):
			suite.Assert().NotEmpty(requestPayload["transactionType"], "transactionType")
			suite.Assert().NotEmpty(requestPayload["channel"], "channel")
			suite.Assert().NotEmpty(requestPayload["reference"], "reference")
			suite.Assert().NotEmpty(requestPayload["reason"], "reason")
			suite.Assert().NotEmpty(requestPayload["transactionAmount"], "transactionAmount")
			suite.Assert().NotEmpty(requestPayload["creditorAccount"], "creditorAccount")

			switch requestPayload["transactionType"] {
			case "PROVISIONAL_CREDIT":
				responseBody = `{
					"id": "1",
					"result": {
						"api": {
							"type": "PROVISIONAL_CREDIT_ACK",
							"reference": "ledger.paymentv2.provisional_credit_1758673957303644368",
							"dateTime": "2025-09-24 00:32:37"
						},
						"account": {
							"accountId": "500400039560945",
							"balanceCents": 8543,
							"status": "ACTIVE"
						},
						"transactionNumber": "QA00000000041051",
						"transactionStatus": "COMPLETED",
						"transactionAmountCents": 1457,
						"originalRequestBase64": "eyJjaGFubmVsIjoiQVBJIiwidHJhbnNhY3Rpb25UeXBlIjoiUFJPVklTSU9OQUxfQ1JFRElUIiwidHJhbnNhY3Rpb25EYXRlVGltZSI6IjIwMjUtMDktMjMgMjA6MzI6MzciLCJyZWZlcmVuY2UiOiJsZWRnZXIucGF5bWVudHYyLnByb3Zpc2lvbmFsX2NyZWRpdF8xNzU4NjczOTU3MzAzNjQ0MzY4IiwicmVhc29uIjoicHJvdmlzaW9uYWwgY3JlZGl0cyBmb3IgdHJhbnNhY3Rpb24gZGlzcHV0ZSIsInRyYW5zYWN0aW9uQW1vdW50Ijp7ImFtb3VudCI6IjE0NTciLCJjdXJyZW5jeSI6IlVTRCJ9LCJjcmVkaXRvckFjY291bnQiOnsiaWRlbnRpZmljYXRpb24iOiI1MDA0MDAwMzk1NjA5NDUiLCJpZGVudGlmaWNhdGlvblR5cGUiOiJBQ0NPVU5UX05VTUJFUiIsImlkZW50aWZpY2F0aW9uVHlwZTIiOiJDSEVDS0lORyJ9fQ==",
						"processId": "PL25092400033050"
					},
					"header": {
						"reference": "ledger.paymentv2.provisional_credit_1758673957303644368",
						"apiKey": "44a2e23b352f4d399be7e95fc6115c66",
						"signature": "MEUCIQC4ql+QhxxLko6NASqNFfq+xO7kx/vso+EYeqyqEBqwYwIgHKWjCvr9sJJSGy4i/KTZIZjaAjI1gIK8WO3q4u4pwYI="
					}
				}`
			default:
				suite.Require().Failf("Unknown method in request: %s", request.Method)
			}

		case ("TransactionService.Payment"):
			suite.Assert().NotEmpty(requestPayload["type"], "type")
			suite.Assert().NotEmpty(requestPayload["product"], "product")
			suite.Assert().NotEmpty(requestPayload["program"], "program")
			suite.Assert().NotEmpty(requestPayload["referenceId"], "referenceId")
			suite.Assert().NotEmpty(requestPayload["customerId"], "customerId")
			suite.Assert().NotEmpty(requestPayload["notes"], "notes")

			switch requestPayload["type"] {
			case "VOID":
				responseBody = `{
					"id": "1",
					"result": {
						"status": "COMPLETED",
						"TransactionID": "62620094",
						"transactionNumber": "QA00000000041052",
						"referenceID": "1758674315689",
						"isPartial": false
					},
					"jsonrpc": "2.0"
				}`
			default:
				suite.Require().Failf("Unknown method in request: %s", request.Method)
			}

		case ("TransactionService.ListTransactions"):
			suite.Assert().NotEmpty(requestPayload["accountNumber"], "accountNumber")
			emptyTransactions := `{
				    "id": "1",
					"error": {
						"code": "NOT_FOUND_TRANSACTION_ENTRIES",
						"message": "Missing transaction entries"
					},
					"jsonrpc": "2.0"
				}`
			presentTransactions := `{
					"id": "1",
					"result": {
						"totalDocs": 1,
						"accountTransactions": [
							{
								"type": "ACH_PULL",
								"ReferenceID": "ledger.ach.transfer_ach_pull_1741989935373707193",
								"timeStamp": "0001-01-01T00:00:00Z",
								"instructedAmount": {
									"amount": 12345,
									"currency": "USD"
								},
								"availableBalance": {
									"amount": 52692,
									"currency": "USD"
								},
								"holdBalance": {
									"amount": 0,
									"currency": "USD"
								},
								"ledgerBalance": {
									"amount": 52692,
									"currency": "USD"
								},
								"debtorAccount": {
									"accountNumber": "000000000000000",
									"party": {
										"name": "DONOR_FIRSTNAME"
									},
									"institutionId": "012345678",
									"institutionName": "BANK"
								},
								"creditorAccount": {
									"accountNumber": "000000000000001",
									"party": {
										"name": "Test",
										"address": {
											"line1": "123 Main St",
											"city": "Adak",
											"state": "AK",
											"country": "US",
											"zipCode": "11111"
										}
									},
									"institutionId": "012345678",
									"institutionName": "BANK",
									"customerName": "Test User",
									"customerID": "100000000000000",
									"nickName": "Default DreamFi Account"
								},
								"processID": "PL25031400013075",
								"status": "COMPLETED",
								"customerID": "100000000000000",
								"transactionID": "1415340",
								"credit": true,
								"autoFileProcess": false,
								"tokenAppFileUpload": false,
								"reason": "Tuesday Test 2",
								"transactionNumber": "QA00000000000002",
								"subTransactionType": "ACH_PULL",
								"transactionTypeDetails": "Transfer"
							}
						]
					}
				}`
			if requestPayload["accountNumber"].(string) == "emptyTransactions" {
				responseBody = emptyTransactions
			} else {
				responseBody = presentTransactions
			}
		case ("TransactionService.GetTransactionsByRef"):
			suite.Assert().NotEmpty(requestPayload["ReferenceId"], "ReferenceId")
			emptyTransaction := `{
				    "id": "1",
					"error": {
						"code": "NOT_FOUND_TRANSACTION",
						"message": "Transaction not found"
					},
					"jsonrpc": "2.0"
				}`
			voidTransaction := `{
					"id": "1",
					"result": {
						"type": "VOID",
						"ReferenceID": "1758050784463",
						"timeStamp": "2025-09-16T19:26:24.463Z",
						"instructedAmount": {
							"amount": 500,
							"currency": "USD"
						},
						"availableBalance": {
							"amount": 350,
							"currency": "USD"
						},
						"holdBalance": {
							"amount": 0,
							"currency": "USD"
						},
						"ledgerBalance": {
							"amount": 350,
							"currency": "USD"
						},
						"debtorAccount": {
							"accountNumber": "500400039005918",
							"party": {
								"name": "Test",
								"address": {
									"line1": "123 Main St",
									"line2": "Apartment 45",
									"city": "Adak",
									"state": "AK",
									"country": "US",
									"zipCode": "11111"
								}
							},
							"institutionId": "124303298",
							"institutionName": "XD Legder",
							"customerName": "Test User",
							"customerID": "100000000034052",
							"nickName": "Default DreamFi Account"
						},
						"creditorAccount": {},
						"processID": "PL25091500032086",
						"status": "COMPLETED",
						"customerID": "100000000039103",
						"transactionID": "58549136",
						"originalReferenceID": "3a64500ae25143dc9759f76bb139b702",
						"credit": false,
						"orginalTransactionType": "PROVISIONAL_CREDIT",
						"autoFileProcess": false,
						"tokenAppFileUpload": false,
						"transactionNumber": "QA00000000040471",
						"subTransactionType": "VOID"
					},
					"jsonrpc": "2.0"
				}`
			transactionExists := `{
					"id": "1",
					"result": {
						"type": "ACH_PULL",
						"ReferenceID": "ledger.ach.transfer_ach_pull_1755001708162912900",
						"timeStamp": "2025-08-12T17:58:28Z",
						"instructedAmount": {
							"amount": 7000,
							"currency": "USD"
						},
						"availableBalance": {
							"amount": 7000,
							"currency": "USD"
						},
						"holdBalance": {
							"amount": 0,
							"currency": "USD"
						},
						"ledgerBalance": {
							"amount": 7000,
							"currency": "USD"
						},
						"debtorAccount": {
							"accountNumber": "987546218371925",
							"party": {
								"name": "DONOR_FIRSTNAME"
							},
							"institutionId": "011002550",
							"institutionName": "EASTERN BANK"
						},
						"creditorAccount": {
							"accountNumber": "500400039005918",
							"party": {
								"name": "Test",
								"address": {
									"line1": "123 Main St",
									"line2": "Apartment 45",
									"city": "Adak",
									"state": "AK",
									"country": "US",
									"zipCode": "11111"
								}
							},
							"institutionId": "124303298",
							"institutionName": "XD Legder",
							"customerName": "Test User",
							"customerID": "100000000034052",
							"nickName": "Default DreamFi Account"
						},
						"processID": "PL25081200030040",
						"status": "COMPLETED",
						"customerID": "100000000034052",
						"transactionID": "41484993",
						"credit": true,
						"autoFileProcess": false,
						"tokenAppFileUpload": false,
						"transactionNumber": "QA00000000038044",
						"subTransactionType": "ACH_PULL",
						"transactionTypeDetails": "Transfer"
					},
					"jsonrpc": "2.0"
				}`
			switch requestPayload["ReferenceId"].(string) {
			case "emptyTransaction":
				responseBody = emptyTransaction
			case "voidTransaction":
				responseBody = voidTransaction
			default:
				responseBody = transactionExists
			}
		case ("StatementService.ListStatement"):
			suite.Assert().NotEmpty(requestPayload["pageNumber"], "pageNumber")
			suite.Assert().NotEmpty(requestPayload["pageSize"], "pageSize")
			responseBody = `{
					"id": "1",
					"result": {
						"statements": [
						{
							"id": "35123",
							"createdDate": "2025-07-02T01:17:17.934Z",
							"updatedDate": "2025-07-02T01:17:17.934Z",
							"customerId": "100000000029028",
							"accountId": "21281104",
							"md5": "d41d8cd98f00b204e9800998ecf8427e",
							"accountNumber": "500400009957614",
							"accountName": "Default DreamFi Account",
							"customerName": "Test User",
							"legalRepID": [
							{
								"ID": "21281103",
								"name": "Test User"
							}
							],
							"currency": "USD",
							"month": "June",
							"year": 2025,
							"lastDate": "2025-07-01T03:59:59.999Z",
							"fileType": "PDFV2"
						},
						{
							"id": "26428",
							"createdDate": "2025-06-24T01:18:50.735Z",
							"updatedDate": "2025-06-24T01:18:50.735Z",
							"customerId": "100000000029028",
							"accountId": "21281104",
							"md5": "d41d8cd98f00b204e9800998ecf8427e",
							"accountNumber": "500400009957614",
							"accountName": "Default DreamFi Account",
							"customerName": "Test User",
							"legalRepID": [
							{
								"ID": "21281103",
								"name": "Test User"
							}
							],
							"currency": "USD",
							"month": "May",
							"year": 2025,
							"lastDate": "2025-06-01T03:59:59.999Z",
							"fileType": "PDFV2"
						}
						],
						"totalCounts": 2
					}
				}`
		case ("StatementService.GetStatement"):
			suite.Assert().NotEmpty(requestPayload["id"], "statementId")
			responseBody = `{
					"id": "1",
					"result": {
						"id": "35123",
						"createdDate": "2025-07-02T01:17:17.934Z",
						"updatedDate": "2025-07-02T01:17:17.934Z",
						"customerId": "100000000029028",
						"accountId": "21281104",
						"pdfV2File": "JVBERi0xLjUKJcTl8uXrp/Og0MTGCjEgMCBvYmoKPDwKL1R5cGUgL0NhdGFsb2cKL1BhZ2VzIDIgMCBSCj4+CmVuZG9iagoyIDAgb2JqCjw8Ci9UeXBlIC9QYWdlcwovQ291bnQgMQovS2lkcyBbMyAwIFJdCj4+CmVuZG9iagozIDAgb2JqCjw8Ci9UeXBlIC9QYWdlCi9QYXJlbnQgMiAwIFIKL1Jlc291cmNlcyA8PAovRm9udCA8PAovRjEgNCAwIFIKPj4KPj4KL01lZGlhQm94IFswIDAgNjEyIDc5Ml0KL0NvbnRlbnRzIDUgMCBSCj4+CmVuZG9iago0IDAgb2JqCjw8Ci9UeXBlIC9Gb250Ci9TdWJ0eXBlIC9UeXBlMQovTmFtZSAvRjEKL0Jhc2VGb250IC9IZWx2ZXRpY2EKPj4KZW5kb2JqCjUgMCBvYmoKPDwKL0xlbmd0aCAxMgo+PgpzdHJlYW0KSGVsbG8gUERGClN0YXJ0eHJlZgowCiUlRU9G",
						"md5": "d41d8cd98f00b204e9800998ecf8427e",
						"accountNumber": "500400009957614",
						"accountName": "Default DreamFi Account",
						"customerName": "Test User",
						"legalRepID": [
						{
							"ID": "21281103",
							"name": "Test User"
						}
						],
						"currency": "USD",
						"month": "June",
						"year": 2025,
						"lastDate": "2025-07-01T03:59:59.999Z",
						"fileType": "PDFV2"
					},
					"jsonrpc": "2.0"
				}`

		default:
			suite.Require().Failf("Unknown method in request: %s", request.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(responseBody))
		if err != nil {
			suite.Require().NoError(err, "Response write should not return an error")
		}
	}))

	config.Config.Ledger.Endpoint = mockLedgerServer.URL + "/jsonrpc"
	config.Config.Ledger.CardsEndpoint = mockLedgerServer.URL + "/cardv2"
	config.Config.Ledger.PaymentsEndpoint = mockLedgerServer.URL + "/paymentv2"

	return mockLedgerServer
}
