package visadps

import (
	"encoding/json"
	"process-api/pkg/clock"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestATMWithdrawalMatchesNetXD(t *testing.T) {
	// jq -r '.item[] | select(.name=="ATM2 Cash Withdrawal") | .request.body.raw' DPS_Txns_APIs.postman_collection.json
	const netxdJSON = `{"Hdr":{"MsgFctn":"REQU","CreDtTm":"2025-09-26T14:35:43Z","PrtcolVrsn":"2.49.1","InitgPty":{"Id":"VisaDPS"}},"Body":{"Tx":{"AcctFr":{"AcctTp":"00"},"TxTp":"01","TxId":{"AcqrrRefData":"412345","LifeCyclTracIdData":{"Id":"444948079490126"},"SysTracAudtNb":"S897344","LclDtTm":"2024-09-10T08:52:41","CardIssrRefData":"{\"CARD-ID\":\"v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14\",\"ProcessorLifeCycleId\":\"d8119928-4772-4f17-ab09-70bef3b9339b\",\"ProcessorTransactionId\":\"4627a9c181ce99082cb8d82b6f39c43da9bb3042260266c9d4bfb452aad3f258\"}","TrnsmssnDtTm":"2025-09-26T14:35:43Z","RtrvlRefNb":"R25092677441"},"AddtlData":[{"Val":"0","Tp":"VAUResult"},{"Val":"N","Tp":"CryptoPurchInd"},{"Val":"N","Tp":"DeferredAuthInd"}],"AddtlAmts":[{"Amt":{"Ccy":"840","Amt":0,"Sgn":true},"Tp":"OTHP"}],"AltrnMsgRsn":["0041"],"TxAmts":{"RcncltnAmt":{"QtnDt":"2024-09-10T00:00:00Z","XchgRate":1,"Ccy":"840","Amt":50},"AmtQlfr":"ACTL","TxAmt":{"Ccy":"840","Amt":50},"DtldAmt":[{"Amt":{"Amt":50},"OthrTp":"BASE","Tp":"OTHP"}],"CrdhldrBllgAmt":{"XchgRate":1,"Ccy":"840","Amt":50}}},"Cntxt":{"Vrfctn":[{"VrfctnRslt":[{"Rslt":"SUCC"}],"Tp":"NPIN"},{"VrfctnRslt":[{"RsltDtls":[{"Val":"0","Tp":"CAMReliability"}]}]},{"VrfctnInf":[{"Val":{"TxtVal":"Z"},"Tp":"Method"}],"Tp":"THDS"},{"VrfctnRslt":[{"Rslt":"SUCC"}],"Tp":"NVSC"}],"TxCntxt":{"MrchntCtgyCd":"5411","SttlmSvc":{"SttlmSvcApld":{"Tp":"VISAInternational"}},"CardPrgrmm":{"CardPrgrmmApld":{"Id":"VSV"}},"Rcncltn":{"Dt":"2024-09-11","Id":"VISAInternational"}},"PtOfSvcCntxt":{"CrdhldrPres":true,"CardDataNtryMd":"MGST","CardPres":true,"CrdhldrActvtd":true},"RskCntxt":[{"RskInptData":[{"Val":"2","Tp":"FalconScoreSource"},{"Val":"0165","Tp":"FalconScoreValue"},{"Val":"0","Tp":"FalconRespCode"},{"Val":"06","Tp":"FalconReason1"},{"Val":"08","Tp":"FalconReason2"},{"Val":"10","Tp":"FalconReason3"},{"Val":"09","Tp":"VisaRiskScore"},{"Val":"5A","Tp":"VisaRiskReason"},{"Val":"02","Tp":"VisaRiskCondCode1"},{"Val":"C2","Tp":"VisaRiskCondCode2"},{"Val":"00","Tp":"VisaRiskCondCode3"}]}]},"Envt":{"Accptr":{"NmAndLctn":"4 HOUSTON TEXANS CIR HOUSTON TXUS","ShrtNm":"Cash Withdrawal","Id":"               ","Adr":{"PstlCd":"80129    ","CtrySubDvsnMjr":"08","Ctry":"USA"}},"Termnl":{"Cpblties":{"CardRdngCpblty":["MGST"],"CrdhldrVrfctnCpblty":[{"Cpblty":"NPIN"}]},"TermnlId":{"Id":"ATM2    "},"Tp":"ATMT"},"Acqrr":{"Ctry":"840","Id":"59992960009"},"Sndr":{"Id":"11111111111"},"Card":{"PmtAcctRef":"V0010010000000000000000000001","XpryDt":"2709","PAN":"9999999999992059"}},"PrcgRslt":{"RsltData":{"RsltDtls":"00"},"ApprvlData":{"ApprvlCd":"250191"}},"SplmtryData":[{"Envlp":{"DPSProcessorData":{"LclDtTmOrgtr":"Y"}}}]}}`

	var expected map[string]any
	if err := json.Unmarshal([]byte(netxdJSON), &expected); err != nil {
		t.Fatalf("Failed to parse netxd JSON: %v", err)
	}

	hdr := expected["Hdr"].(map[string]any)
	body := expected["Body"].(map[string]any)
	tx := body["Tx"].(map[string]any)
	txId := tx["TxId"].(map[string]any)
	envt := body["Envt"].(map[string]any)
	card := envt["Card"].(map[string]any)

	var cardIssuerRef CardIssuerRefData
	if err := json.Unmarshal([]byte(txId["CardIssrRefData"].(string)), &cardIssuerRef); err != nil {
		t.Fatalf("Failed to unmarshal CardIssrRefData: %v", err)
	}

	cardId := cardIssuerRef.CardId
	last4 := card["PAN"].(string)[len(card["PAN"].(string))-4:]
	expiry := card["XpryDt"].(string)

	txAmts := tx["TxAmts"].(map[string]any)
	txAmt := txAmts["TxAmt"].(map[string]any)
	amountCents := int(txAmt["Amt"].(float64) * 100)
	crdhldrBllgAmt := txAmts["CrdhldrBllgAmt"].(map[string]any)
	billingAmountCents := int(crdhldrBllgAmt["Amt"].(float64) * 100)

	ids := TransactionIDs{
		CardExternalId:         cardId,
		ProcessorLifeCycleId:   cardIssuerRef.ProcessorLifeCycleId,
		ProcessorTransactionId: cardIssuerRef.ProcessorTransactionId,
		LifeCycleId:            txId["LifeCyclTracIdData"].(map[string]any)["Id"].(string),
		SysTracAudtNb:          txId["SysTracAudtNb"].(string),
		RtrvlRefNb:             txId["RtrvlRefNb"].(string),
		AcqrrRefData:           "412345",
	}

	refTime, err := time.Parse(time.RFC3339, hdr["CreDtTm"].(string))
	defer clock.Freeze(refTime)()
	if err != nil {
		t.Fatalf("Failed to parse reference time: %v", err)
	}

	localTime, err := time.Parse("2006-01-02T15:04:05", txId["LclDtTm"].(string))
	if err != nil {
		t.Fatalf("Failed to parse local time: %v", err)
	}

	message := NewMessageBuilder(cardId, last4, expiry).WithTransactionIDs(ids).WithLocalTime(localTime).BuildATMWithdrawal(amountCents, billingAmountCents)

	actualJson, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to serialize ISO20022Message into JSON: %v", err)
	}
	var actual map[string]any
	err = json.Unmarshal(actualJson, &actual)
	if err != nil {
		t.Fatalf("Failed to parse actualJson into map: %v", err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Payload mismatch (-expected +actual):\n%s", diff)
	}
}

func TestBuildATMDeposit(t *testing.T) {
	// jq -r '.item[] | select(.name=="PLUS8 Deposit") | .request.body.raw' DPS_Txns_APIs.postman_collection.json
	const netxdJSON = `{"Hdr":{"MsgFctn":"REQU","CreDtTm":"2025-09-26T14:35:49Z","PrtcolVrsn":"2.49.1","InitgPty":{"Id":"VisaDPS"}},"Body":{"Tx":{"AcctTo":{"AcctTp":"00"},"TxTp":"21","TxId":{"AcqrrRefData":"412345     ","LifeCyclTracIdData":{"Id":"730735475180276"},"SysTracAudtNb":"S897345","LclDtTm":"2024-09-17T15:50:05","CardIssrRefData":"{\"CARD-ID\":\"v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14\",\"ProcessorLifeCycleId\":\"40b96f7a-8393-4515-864f-0e5b93e3f681\",\"ProcessorTransactionId\":\"5fbabb8398013ae4929e9b2aadb7b7dd02b3b39d92c9098467ae6fd965771d13\"}","TrnsmssnDtTm":"2025-09-26T14:35:49Z","RtrvlRefNb":"R25092621408"},"AddtlData":[{"Val":"0","Tp":"VAUResult"},{"Val":"N","Tp":"CryptoPurchInd"},{"Val":"N","Tp":"DeferredAuthInd"}],"AddtlAmts":[{"Amt":{"Ccy":"840","Amt":0,"Sgn":true},"Tp":"OTHP"}],"SpclPrgrmmQlfctn":[{"Dtl":[{"Val":"875","Nm":"FPI"}]}],"TxAmts":{"RcncltnAmt":{"QtnDt":"2024-09-18T00:00:00Z","XchgRate":1,"Ccy":"840","Amt":80},"AmtQlfr":"ACTL","TxAmt":{"Ccy":"840","Amt":80},"DtldAmt":[{"Amt":{"Amt":80},"OthrTp":"BASE","Tp":"OTHP"}],"CrdhldrBllgAmt":{"XchgRate":1,"Ccy":"840","Amt":80}}},"Cntxt":{"Vrfctn":[{"VrfctnRslt":[{"Rslt":"SUCC"}],"Tp":"NPIN"},{"VrfctnRslt":[{"RsltDtls":[{"Val":"0","Tp":"CAMReliability"}]}]},{"VrfctnInf":[{"Val":{"TxtVal":"Z"},"Tp":"Method"}],"Tp":"THDS"},{"VrfctnRslt":[{"Rslt":"SUCC"}],"Tp":"NVSC"}],"TxCntxt":{"MrchntCtgyCd":"6011","SttlmSvc":{"SttlmSvcApld":{"Tp":"VISAInternational"}},"CardPrgrmm":{"CardPrgrmmApld":{"Id":"VSP"}},"Rcncltn":{"Dt":"2024-09-17","Id":"VISAInternational"}},"PtOfSvcCntxt":{"CrdhldrPres":true,"CardDataNtryMd":"MGST","CardPres":true,"CrdhldrActvtd":true},"RskCntxt":[{"RskInptData":[{"Val":"3","Tp":"FalconScoreSource"},{"Val":"0029","Tp":"FalconScoreValue"},{"Val":"0","Tp":"FalconRespCode"},{"Val":"34","Tp":"FalconReason1"},{"Val":"14","Tp":"FalconReason2"},{"Val":"77","Tp":"FalconReason3"},{"Val":"09","Tp":"VisaRiskScore"},{"Val":"5A","Tp":"VisaRiskReason"},{"Val":"02","Tp":"VisaRiskCondCode1"},{"Val":"C2","Tp":"VisaRiskCondCode2"},{"Val":"00","Tp":"VisaRiskCondCode3"}]}]},"Envt":{"Accptr":{"NmAndLctn":"4 HOUSTON TEXANS CIR HOUSTON TXUS","ShrtNm":"Deposit","Id":"               ","Adr":{"PstlCd":"80129    ","CtrySubDvsnMjr":"08","Ctry":"USA"}},"Termnl":{"Cpblties":{"CardRdngCpblty":["MGST"],"CrdhldrVrfctnCpblty":[{"Cpblty":"NPIN"}]},"TermnlId":{"Id":"PLUS8   "},"Tp":"ATMT"},"Acqrr":{"Ctry":"840","Id":"59992960009"},"Sndr":{"Id":"11111111111"},"Card":{"PmtAcctRef":"V0010010000000000000000000001","XpryDt":"2709","PAN":"9999999999999958"}},"SplmtryData":[{"Envlp":{"DPSProcessorData":{"LclDtTmOrgtr":"Y"}}}]}}`

	var expected map[string]any
	if err := json.Unmarshal([]byte(netxdJSON), &expected); err != nil {
		t.Fatalf("Failed to parse netxd JSON: %v", err)
	}

	hdr := expected["Hdr"].(map[string]any)
	body := expected["Body"].(map[string]any)
	tx := body["Tx"].(map[string]any)
	txId := tx["TxId"].(map[string]any)
	envt := body["Envt"].(map[string]any)
	card := envt["Card"].(map[string]any)

	var cardIssuerRef CardIssuerRefData
	if err := json.Unmarshal([]byte(txId["CardIssrRefData"].(string)), &cardIssuerRef); err != nil {
		t.Fatalf("Failed to unmarshal CardIssrRefData: %v", err)
	}

	cardId := cardIssuerRef.CardId
	last4 := card["PAN"].(string)[len(card["PAN"].(string))-4:]
	expiry := card["XpryDt"].(string)

	txAmts := tx["TxAmts"].(map[string]any)
	txAmt := txAmts["TxAmt"].(map[string]any)
	amountCents := int(txAmt["Amt"].(float64) * 100)
	crdhldrBllgAmt := txAmts["CrdhldrBllgAmt"].(map[string]any)
	billingAmountCents := int(crdhldrBllgAmt["Amt"].(float64) * 100)

	ids := TransactionIDs{
		CardExternalId:         cardId,
		ProcessorLifeCycleId:   cardIssuerRef.ProcessorLifeCycleId,
		ProcessorTransactionId: cardIssuerRef.ProcessorTransactionId,
		LifeCycleId:            txId["LifeCyclTracIdData"].(map[string]any)["Id"].(string),
		SysTracAudtNb:          txId["SysTracAudtNb"].(string),
		RtrvlRefNb:             txId["RtrvlRefNb"].(string),
		AcqrrRefData:           "412345     ",
	}

	refTime, err := time.Parse(time.RFC3339, hdr["CreDtTm"].(string))
	defer clock.Freeze(refTime)()
	if err != nil {
		t.Fatalf("Failed to parse reference time: %v", err)
	}

	localTime, err := time.Parse("2006-01-02T15:04:05", txId["LclDtTm"].(string))
	if err != nil {
		t.Fatalf("Failed to parse local time: %v", err)
	}

	message := NewMessageBuilder(cardId, last4, expiry).WithTransactionIDs(ids).WithLocalTime(localTime).BuildATMDeposit(amountCents, billingAmountCents)

	actualJson, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to serialize ISO20022Message into JSON: %v", err)
	}
	var actual map[string]any
	err = json.Unmarshal(actualJson, &actual)
	if err != nil {
		t.Fatalf("Failed to parse actualJson into map: %v", err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Payload mismatch (-expected +actual):\n%s", diff)
	}
}

func TestBuildPurchase(t *testing.T) {
	// jq -r '.item[] | select(.name=="POS10 Purchase") | .request.body.raw' DPS_Txns_APIs.postman_collection.json
	const netxdJSON = `{"Hdr":{"MsgFctn":"REQU","CreDtTm":"2025-09-26T14:35:54Z","PrtcolVrsn":"2.49.1","InitgPty":{"Id":"VisaDPS"}},"Body":{"Tx":{"AcctFr":{"AcctTp":"00"},"TxTp":"00","TxId":{"AcqrrRefData":"412345     ","LifeCyclTracIdData":{"Id":"992064955384422"},"SysTracAudtNb":"S897347","LclDtTm":"2024-09-10T08:53:29","CardIssrRefData":"{\"CARD-ID\":\"v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14\",\"ProcessorLifeCycleId\":\"d321edbf-3653-4561-a697-6ac05c09a93f\",\"ProcessorTransactionId\":\"3c7e644cf8c1e91d22aaa6d5c3ff79f555fe7d9dc9e26fc21a0cf0f31750b209\"}","TrnsmssnDtTm":"2025-09-26T14:35:54Z","RtrvlRefNb":"R25092669347"},"AddtlData":[{"Val":"0","Tp":"VAUResult"},{"Val":"N","Tp":"CryptoPurchInd"},{"Val":"N","Tp":"DeferredAuthInd"}],"AddtlAmts":[{"Amt":{"Ccy":"840","Amt":0,"Sgn":true},"Tp":"OTHP"}],"AltrnMsgRsn":["0031"],"TxAmts":{"RcncltnAmt":{"QtnDt":"2024-09-11T00:00:00Z","XchgRate":1,"Ccy":"840","Amt":13.5},"AmtQlfr":"ACTL","TxAmt":{"Ccy":"840","Amt":13.5},"DtldAmt":[{"Amt":{"Amt":13.5},"OthrTp":"BASE","Tp":"OTHP"}],"CrdhldrBllgAmt":{"XchgRate":1,"Ccy":"840","Amt":10.6}}},"Cntxt":{"Vrfctn":[{"VrfctnRslt":[{"RsltDtls":[{"Val":"2","Tp":"CAVV Result Code"}],"Tp":"AuthenticationValue","Rslt":"SUCC"}],"SubTp":"Visa","VrfctnInf":[{"Val":{"TxtVal":"Z"},"Tp":"Method"}],"Tp":"THDS"},{"VrfctnRslt":[{"RsltDtls":[{"Val":"0","Tp":"CAMReliability"}]}]}],"TxCntxt":{"MrchntCtgyCd":"5947","SttlmSvc":{"SttlmSvcApld":{"Tp":"VISAInternational"}},"CardPrgrmm":{"CardPrgrmmApld":{"Id":"VSN"}},"Rcncltn":{"Dt":"2024-09-10","Id":"VISAInternational"}},"PtOfSvcCntxt":{"CardDataNtryMd":"KEEN","CrdhldrActvtd":true,"EComrcInd":true,"EComrcData":[{"Val":"5","Tp":"ECI"}]},"RskCntxt":[{"RskInptData":[{"Val":"3","Tp":"FalconScoreSource"},{"Val":"0469","Tp":"FalconScoreValue"},{"Val":"0","Tp":"FalconRespCode"},{"Val":"08","Tp":"FalconReason1"},{"Val":"06","Tp":"FalconReason2"},{"Val":"02","Tp":"FalconReason3"},{"Val":"09","Tp":"VisaRiskScore"},{"Val":"5A","Tp":"VisaRiskReason"},{"Val":"02","Tp":"VisaRiskCondCode1"},{"Val":"C2","Tp":"VisaRiskCondCode2"},{"Val":"00","Tp":"VisaRiskCondCode3"}]}]},"Envt":{"Accptr":{"NmAndLctn":"DNH*GODADDY#3521500092 TEMPE        AZUS","Id":"Authorization  ","Adr":{"PstlCd":"         ","Ctry":"USA"}},"Termnl":{"Cpblties":{"CardRdngCpblty":["MLEY"],"CrdhldrVrfctnCpblty":[{"Cpblty":"UNSP"}]},"TermnlId":{"Id":"CNP8    "},"Tp":"POST"},"Acqrr":{"Ctry":"840","Id":"59000000204"},"Sndr":{"Id":"11111111111"},"Card":{"PmtAcctRef":"V0010010000000000000000000001","XpryDt":"2709","PAN":"9999999999992059","CardPdctTp":"F"}},"PrcgRslt":{"RsltData":{"RsltDtls":"00"},"ApprvlData":{"ApprvlCd":"250191"}},"SplmtryData":[{"Envlp":{"DPSProcessorData":{"LclDtTmOrgtr":"Y"}}}]}}`

	var expected map[string]any
	if err := json.Unmarshal([]byte(netxdJSON), &expected); err != nil {
		t.Fatalf("Failed to parse netxd JSON: %v", err)
	}

	body := expected["Body"].(map[string]any)
	tx := body["Tx"].(map[string]any)
	txId := tx["TxId"].(map[string]any)

	cardId := "v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14"
	last4 := "2059"
	expiry := "2709"
	amountCents := 1350

	txAmts := tx["TxAmts"].(map[string]any)
	crdhldrBllgAmt := txAmts["CrdhldrBllgAmt"].(map[string]any)
	billingAmountCents := int(crdhldrBllgAmt["Amt"].(float64) * 100)

	txCntxt := body["Cntxt"].(map[string]any)["TxCntxt"].(map[string]any)
	mcc := txCntxt["MrchntCtgyCd"].(string)

	ids := TransactionIDs{
		CardExternalId:         cardId,
		ProcessorLifeCycleId:   "d321edbf-3653-4561-a697-6ac05c09a93f",
		ProcessorTransactionId: "3c7e644cf8c1e91d22aaa6d5c3ff79f555fe7d9dc9e26fc21a0cf0f31750b209",
		LifeCycleId:            txId["LifeCyclTracIdData"].(map[string]any)["Id"].(string),
		SysTracAudtNb:          txId["SysTracAudtNb"].(string),
		RtrvlRefNb:             txId["RtrvlRefNb"].(string),
		AcqrrRefData:           "412345     ",
	}

	refTime, err := time.Parse(time.RFC3339, "2025-09-26T14:35:54Z")
	defer clock.Freeze(refTime)()
	if err != nil {
		t.Fatalf("Failed to parse reference time: %v", err)
	}

	localTime, err := time.Parse("2006-01-02T15:04:05", txId["LclDtTm"].(string))
	if err != nil {
		t.Fatalf("Failed to parse local time: %v", err)
	}

	message := NewMessageBuilder(cardId, last4, expiry).WithTransactionIDs(ids).WithLocalTime(localTime).BuildPurchase(amountCents, billingAmountCents, mcc)

	actualJson, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to serialize ISO20022Message into JSON: %v", err)
	}
	var actual map[string]any
	err = json.Unmarshal(actualJson, &actual)
	if err != nil {
		t.Fatalf("Failed to parse actualJson into map: %v", err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Payload mismatch (-expected +actual):\n%s", diff)
	}
}

func TestBuildMerchandiseReturn(t *testing.T) {
	// jq -r '.item[] | select(.name=="POS12 merchendise Return") | .request.body.raw' DPS_Txns_APIs.postman_collection.json
	const netxdJSON = `{"Hdr":{"MsgFctn":"REQU","CreDtTm":"2025-09-26T14:35:57Z","PrtcolVrsn":"2.49.1","InitgPty":{"Id":"VisaDPS"}},"Body":{"Tx":{"AcctFr":{"AcctTp":"00"},"TxTp":"20","TxId":{"AcqrrRefData":"412345     ","LifeCyclTracIdData":{"Id":"992064955384422"},"SysTracAudtNb":"S897348","LclDtTm":"2024-09-10T08:53:29","CardIssrRefData":"{\"CARD-ID\":\"v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14\",\"ProcessorLifeCycleId\":\"3bfec337-bfdc-4107-b8a9-c2c36fd5e0d5\",\"ProcessorTransactionId\":\"ff5d6f4d8cca2caf838d44ad900e2eed8aa9056db503b3e2188ad8a8fb8d2165\"}","TrnsmssnDtTm":"2025-09-26T14:35:57Z","RtrvlRefNb":"R25092682918"},"AddtlData":[{"Val":"0","Tp":"VAUResult"},{"Val":"N","Tp":"CryptoPurchInd"},{"Val":"N","Tp":"DeferredAuthInd"}],"AddtlAmts":[{"Amt":{"Ccy":"840","Amt":0,"Sgn":true},"Tp":"OTHP"}],"AltrnMsgRsn":["0031"],"TxAmts":{"RcncltnAmt":{"QtnDt":"2024-09-11T00:00:00Z","XchgRate":1,"Ccy":"840","Amt":13.5},"AmtQlfr":"ACTL","TxAmt":{"Ccy":"840","Amt":13.5},"DtldAmt":[{"Amt":{"Amt":13.5},"OthrTp":"BASE","Tp":"OTHP"}],"CrdhldrBllgAmt":{"XchgRate":1,"Ccy":"840","Amt":12.6}}},"Cntxt":{"Vrfctn":[{"VrfctnRslt":[{"RsltDtls":[{"Val":"2","Tp":"CAVV Result Code"}],"Tp":"AuthenticationValue","Rslt":"SUCC"}],"SubTp":"Visa","VrfctnInf":[{"Val":{"TxtVal":"Z"},"Tp":"Method"}],"Tp":"THDS"},{"VrfctnRslt":[{"RsltDtls":[{"Val":"0","Tp":"CAMReliability"}]}]}],"TxCntxt":{"MrchntCtgyCd":"5947","SttlmSvc":{"SttlmSvcApld":{"Tp":"VISAInternational"}},"CardPrgrmm":{"CardPrgrmmApld":{"Id":"VSN"}},"Rcncltn":{"Dt":"2024-09-10","Id":"VISAInternational"}},"PtOfSvcCntxt":{"CardDataNtryMd":"KEEN","CrdhldrActvtd":true,"EComrcInd":true,"EComrcData":[{"Val":"5","Tp":"ECI"}]},"RskCntxt":[{"RskInptData":[{"Val":"3","Tp":"FalconScoreSource"},{"Val":"0469","Tp":"FalconScoreValue"},{"Val":"0","Tp":"FalconRespCode"},{"Val":"08","Tp":"FalconReason1"},{"Val":"06","Tp":"FalconReason2"},{"Val":"02","Tp":"FalconReason3"},{"Val":"09","Tp":"VisaRiskScore"},{"Val":"5A","Tp":"VisaRiskReason"},{"Val":"02","Tp":"VisaRiskCondCode1"},{"Val":"C2","Tp":"VisaRiskCondCode2"},{"Val":"00","Tp":"VisaRiskCondCode3"}]}]},"Envt":{"Accptr":{"NmAndLctn":"DNH*GODADDY#3521500092 TEMPE        AZUS","Id":"Authorization  ","Adr":{"PstlCd":"         ","Ctry":"USA"}},"Termnl":{"Cpblties":{"CardRdngCpblty":["MLEY"],"CrdhldrVrfctnCpblty":[{"Cpblty":"UNSP"}]},"TermnlId":{"Id":"CNP8    "},"Tp":"POST"},"Acqrr":{"Ctry":"840","Id":"59000000204"},"Sndr":{"Id":"11111111111"},"Card":{"PmtAcctRef":"V0010010000000000000000000001","XpryDt":"2709","PAN":"9999999999992059","CardPdctTp":"F"}},"PrcgRslt":{"RsltData":{"RsltDtls":"00"},"ApprvlData":{"ApprvlCd":"250191"}},"SplmtryData":[{"Envlp":{"DPSProcessorData":{"LclDtTmOrgtr":"Y"}}}]}}`

	var expected map[string]any
	if err := json.Unmarshal([]byte(netxdJSON), &expected); err != nil {
		t.Fatalf("Failed to parse netxd JSON: %v", err)
	}

	body := expected["Body"].(map[string]any)
	tx := body["Tx"].(map[string]any)
	txId := tx["TxId"].(map[string]any)

	cardId := "v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14"
	last4 := "2059"
	expiry := "2709"
	amountCents := 1350

	txAmts := tx["TxAmts"].(map[string]any)
	crdhldrBllgAmt := txAmts["CrdhldrBllgAmt"].(map[string]any)
	billingAmountCents := int(crdhldrBllgAmt["Amt"].(float64) * 100)

	ids := TransactionIDs{
		CardExternalId:         cardId,
		ProcessorLifeCycleId:   "3bfec337-bfdc-4107-b8a9-c2c36fd5e0d5",
		ProcessorTransactionId: "ff5d6f4d8cca2caf838d44ad900e2eed8aa9056db503b3e2188ad8a8fb8d2165",
		LifeCycleId:            txId["LifeCyclTracIdData"].(map[string]any)["Id"].(string),
		SysTracAudtNb:          txId["SysTracAudtNb"].(string),
		RtrvlRefNb:             txId["RtrvlRefNb"].(string),
		AcqrrRefData:           "412345     ",
	}

	refTime, err := time.Parse(time.RFC3339, "2025-09-26T14:35:57Z")
	defer clock.Freeze(refTime)()
	if err != nil {
		t.Fatalf("Failed to parse reference time: %v", err)
	}

	localTime, err := time.Parse("2006-01-02T15:04:05", txId["LclDtTm"].(string))
	if err != nil {
		t.Fatalf("Failed to parse local time: %v", err)
	}

	message := NewMessageBuilder(cardId, last4, expiry).WithTransactionIDs(ids).WithLocalTime(localTime).BuildMerchandiseReturn(amountCents, billingAmountCents)

	actualJson, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to serialize ISO20022Message into JSON: %v", err)
	}
	var actual map[string]any
	err = json.Unmarshal(actualJson, &actual)
	if err != nil {
		t.Fatalf("Failed to parse actualJson into map: %v", err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Payload mismatch (-expected +actual):\n%s", diff)
	}
}

func TestBuildBillPayDebit(t *testing.T) {
	// jq -r '.item[] | select(.name=="BillPayDebit") | .request.body.raw' DPS_Txns_APIs.postman_collection.json
	const netxdJSON = `{"Hdr":{"MsgFctn":"ADVC","CreDtTm":"2025-09-26T14:36:21Z","PrtcolVrsn":"2.49.1","InitgPty":{"Id":"VisaDPS"}},"Body":{"Tx":{"AcctFr":{"AcctTp":"00"},"TxTp":"50","TxId":{"AcqrrRefData":"412345     ","LifeCyclTracIdData":{"Id":"992064955384422"},"SysTracAudtNb":"S897353","LclDtTm":"2024-09-10T08:53:29","CardIssrRefData":"{\"CARD-ID\":\"v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14\",\"ProcessorLifeCycleId\":\"a18580b7-da21-4f41-9b31-2b1606760174\",\"ProcessorTransactionId\":\"1f90b259556146b5984bcc70670aaa36a696b0644a945d6c076fc3277c414c4c\"}","TrnsmssnDtTm":"2025-09-26T14:36:21Z","RtrvlRefNb":"R25092633964"},"AddtlData":[{"Val":"0","Tp":"VAUResult"},{"Val":"N","Tp":"CryptoPurchInd"},{"Val":"N","Tp":"DeferredAuthInd"}],"AddtlAmts":[{"Amt":{"Ccy":"840","Amt":0,"Sgn":true},"Tp":"OTHP"}],"AltrnMsgRsn":["0031"],"TxAmts":{"RcncltnAmt":{"QtnDt":"2024-09-11T00:00:00Z","XchgRate":1,"Ccy":"840","Amt":13.5},"AmtQlfr":"ACTL","TxAmt":{"Ccy":"840","Amt":13.5},"DtldAmt":[{"Amt":{"Amt":13.5},"OthrTp":"BASE","Tp":"OTHP"}],"CrdhldrBllgAmt":{"XchgRate":1,"Ccy":"840","Amt":13}}},"Cntxt":{"Vrfctn":[{"VrfctnRslt":[{"RsltDtls":[{"Val":"2","Tp":"CAVV Result Code"}],"Tp":"AuthenticationValue","Rslt":"SUCC"}],"SubTp":"Visa","VrfctnInf":[{"Val":{"TxtVal":"Z"},"Tp":"Method"}],"Tp":"THDS"},{"VrfctnRslt":[{"RsltDtls":[{"Val":"0","Tp":"CAMReliability"}]}]}],"TxCntxt":{"MrchntCtgyCd":"5947","SttlmSvc":{"SttlmSvcApld":{"Tp":"VISAInternational"}},"CardPrgrmm":{"CardPrgrmmApld":{"Id":"VSN"}},"Rcncltn":{"Dt":"2024-09-10","Id":"VISAInternational"}},"PtOfSvcCntxt":{"CardDataNtryMd":"KEEN","CrdhldrActvtd":true,"EComrcInd":true,"EComrcData":[{"Val":"5","Tp":"ECI"}]},"RskCntxt":[{"RskInptData":[{"Val":"3","Tp":"FalconScoreSource"},{"Val":"0469","Tp":"FalconScoreValue"},{"Val":"0","Tp":"FalconRespCode"},{"Val":"08","Tp":"FalconReason1"},{"Val":"06","Tp":"FalconReason2"},{"Val":"02","Tp":"FalconReason3"},{"Val":"09","Tp":"VisaRiskScore"},{"Val":"5A","Tp":"VisaRiskReason"},{"Val":"02","Tp":"VisaRiskCondCode1"},{"Val":"C2","Tp":"VisaRiskCondCode2"},{"Val":"00","Tp":"VisaRiskCondCode3"}]}]},"Envt":{"Accptr":{"NmAndLctn":"DNH*GODADDY#3521500092 TEMPE        AZUS","Id":"Authorization  ","Adr":{"PstlCd":"         ","Ctry":"USA"}},"Termnl":{"Cpblties":{"CardRdngCpblty":["MLEY"],"CrdhldrVrfctnCpblty":[{"Cpblty":"UNSP"}]},"TermnlId":{"Id":"CNP8    "},"Tp":"POST"},"Acqrr":{"Ctry":"840","Id":"59000000204"},"Sndr":{"Id":"11111111111"},"Card":{"PmtAcctRef":"V0010010000000000000000000001","XpryDt":"2709","PAN":"9999999999992059","CardPdctTp":"F"}},"PrcgRslt":{"RsltData":{"RsltDtls":"00"},"ApprvlData":{"ApprvlCd":"250191"}},"SplmtryData":[{"Envlp":{"DPSProcessorData":{"LclDtTmOrgtr":"Y"}}}]}}`

	var expected map[string]any
	if err := json.Unmarshal([]byte(netxdJSON), &expected); err != nil {
		t.Fatalf("Failed to parse netxd JSON: %v", err)
	}

	body := expected["Body"].(map[string]any)
	tx := body["Tx"].(map[string]any)
	txId := tx["TxId"].(map[string]any)

	cardId := "v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14"
	last4 := "2059"
	expiry := "2709"
	amountCents := 1350

	txAmts := tx["TxAmts"].(map[string]any)
	crdhldrBllgAmt := txAmts["CrdhldrBllgAmt"].(map[string]any)
	billingAmountCents := int(crdhldrBllgAmt["Amt"].(float64) * 100)

	ids := TransactionIDs{
		CardExternalId:         cardId,
		ProcessorLifeCycleId:   "a18580b7-da21-4f41-9b31-2b1606760174",
		ProcessorTransactionId: "1f90b259556146b5984bcc70670aaa36a696b0644a945d6c076fc3277c414c4c",
		LifeCycleId:            txId["LifeCyclTracIdData"].(map[string]any)["Id"].(string),
		SysTracAudtNb:          txId["SysTracAudtNb"].(string),
		RtrvlRefNb:             txId["RtrvlRefNb"].(string),
		AcqrrRefData:           "412345     ",
	}

	refTime, err := time.Parse(time.RFC3339, "2025-09-26T14:36:21Z")
	defer clock.Freeze(refTime)()
	if err != nil {
		t.Fatalf("Failed to parse reference time: %v", err)
	}

	localTime, err := time.Parse("2006-01-02T15:04:05", txId["LclDtTm"].(string))
	if err != nil {
		t.Fatalf("Failed to parse local time: %v", err)
	}

	message := NewMessageBuilder(cardId, last4, expiry).WithTransactionIDs(ids).WithLocalTime(localTime).BuildBillPayDebit(amountCents, billingAmountCents)

	actualJson, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to serialize ISO20022Message into JSON: %v", err)
	}
	var actual map[string]any
	err = json.Unmarshal(actualJson, &actual)
	if err != nil {
		t.Fatalf("Failed to parse actualJson into map: %v", err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Payload mismatch (-expected +actual):\n%s", diff)
	}
}

func TestBuildBillPayCredit(t *testing.T) {
	// jq -r '.item[] | select(.name=="BillPayCredit") | .request.body.raw' DPS_Txns_APIs.postman_collection.json
	const netxdJSON = `{"Hdr":{"MsgFctn":"REQU","CreDtTm":"2025-09-26T14:36:23Z","PrtcolVrsn":"2.49.1","InitgPty":{"Id":"VisaDPS"}},"Body":{"Tx":{"AcctTo":{"AcctTp":"00"},"TxTp":"55","TxId":{"AcqrrRefData":"412345     ","LifeCyclTracIdData":{"Id":"424237150563117"},"SysTracAudtNb":"S897354","LclDtTm":"2024-09-10T08:53:23","CardIssrRefData":"{\"CARD-ID\":\"v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14\",\"ProcessorLifeCycleId\":\"9f7429b1-8d78-4dca-b81a-316db128646c\",\"ProcessorTransactionId\":\"4295accd03ebfd9df6c0d1c6244f56148aa7ea008482938f90d4c52bf81c3c5b\"}","TrnsmssnDtTm":"2025-09-26T14:36:23Z","RtrvlRefNb":"R25092678329"},"AddtlData":[{"Val":"0","Tp":"VAUResult"},{"Val":"N","Tp":"CryptoPurchInd"},{"Val":"N","Tp":"DeferredAuthInd"},{"Val":"$PP","Tp":"MONEYXFERTYPE"}],"FndsSvcs":{"FndgSvc":{"Desc":"Person to person","BizPurp":"$PP"}},"AddtlAmts":[{"Amt":{"Ccy":"840","Amt":0,"Sgn":true},"Tp":"OTHP"}],"SpclPrgrmmQlfctn":[{"Dtl":[{"Val":"421","Nm":"FPI"},{"Val":"0","Nm":"RA"}]},{}],"TxAmts":{"RcncltnAmt":{"QtnDt":"2024-09-10T00:00:00Z","XchgRate":1,"Ccy":"840","Amt":55},"AmtQlfr":"ACTL","TxAmt":{"Ccy":"840","Amt":55},"DtldAmt":[{"Amt":{"Amt":55},"OthrTp":"BASE","Tp":"OTHP"}],"CrdhldrBllgAmt":{"XchgRate":1,"Ccy":"840","Amt":1}}},"Cntxt":{"Vrfctn":[{"VrfctnRslt":[{"RsltDtls":[{"Val":"0","Tp":"CAMReliability"}]}]},{"SubTp":"Visa","VrfctnInf":[{"Val":{"TxtVal":"Z"},"Tp":"Method"}],"Tp":"THDS"},{"Tp":"CPSG"}],"TxCntxt":{"MrchntCtgyCd":"4829","SttlmSvc":{"SttlmSvcApld":{"Tp":"VISAInternational"}},"CardPrgrmm":{"CardPrgrmmApld":{"Id":"VSN"}},"Rcncltn":{"Dt":"2024-09-10","Id":"VISAInternational"}},"PtOfSvcCntxt":{"CrdhldrPres":true,"AttnddInd":true,"CardDataNtryMd":"KEEN","CrdhldrActvtd":true},"RskCntxt":[{"RskInptData":[{"Val":"3","Tp":"FalconScoreSource"},{"Val":"0000","Tp":"FalconScoreValue"},{"Val":"0","Tp":"FalconRespCode"},{"Val":"00","Tp":"FalconReason1"},{"Val":"00","Tp":"FalconReason2"},{"Val":"00","Tp":"FalconReason3"},{"Val":"09","Tp":"VisaRiskScore"},{"Val":"5A","Tp":"VisaRiskReason"},{"Val":"02","Tp":"VisaRiskCondCode1"},{"Val":"C2","Tp":"VisaRiskCondCode2"},{"Val":"00","Tp":"VisaRiskCondCode3"}]}]},"Envt":{"Pyer":{"Cstmr":{"Crdntls":[{"IdVal":"1234567890123456","OthrIdCd":"REFERENCE#","IdCd":"OTHP"}],"Adr":{"CtrySubDvsnMjr":"CO","Ctry":"USA","AdrLine1":"1340 Pennsylvania St","TwnNm":"Denver"},"Nm":"Molly Brown"}},"Accptr":{"NmAndLctn":"PAYPAL 4029311133 CAUS","Id":"Original Credit","Adr":{"PstlCd":"80203    ","CtrySubDvsnMjr":"08","Ctry":"USA"}},"Termnl":{"Cpblties":{"CardRdngCpblty":["UNSP"],"CrdhldrVrfctnCpblty":[{"Cpblty":"UNSP"}]},"TermnlId":{"Id":"VD6     "},"Tp":"POST"},"Acqrr":{"Ctry":"840","Id":"59000000204"},"Sndr":{"Id":"11111111111"},"Card":{"PmtAcctRef":"V0010010000000000000000000001","XpryDt":"2709","PAN":"9999999999992059"}},"SplmtryData":[{"Envlp":{"DPSProcessorData":{"LclDtTmOrgtr":"Y"}}}]}}`

	var expected map[string]any
	if err := json.Unmarshal([]byte(netxdJSON), &expected); err != nil {
		t.Fatalf("Failed to parse netxd JSON: %v", err)
	}

	body := expected["Body"].(map[string]any)
	tx := body["Tx"].(map[string]any)
	txId := tx["TxId"].(map[string]any)

	cardId := "v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14"
	last4 := "2059"
	expiry := "2709"
	amountCents := 5500

	txAmts := tx["TxAmts"].(map[string]any)
	crdhldrBllgAmt := txAmts["CrdhldrBllgAmt"].(map[string]any)
	billingAmountCents := int(crdhldrBllgAmt["Amt"].(float64) * 100)

	ids := TransactionIDs{
		CardExternalId:         cardId,
		ProcessorLifeCycleId:   "9f7429b1-8d78-4dca-b81a-316db128646c",
		ProcessorTransactionId: "4295accd03ebfd9df6c0d1c6244f56148aa7ea008482938f90d4c52bf81c3c5b",
		LifeCycleId:            txId["LifeCyclTracIdData"].(map[string]any)["Id"].(string),
		SysTracAudtNb:          txId["SysTracAudtNb"].(string),
		RtrvlRefNb:             txId["RtrvlRefNb"].(string),
		AcqrrRefData:           "412345     ",
	}

	refTime, err := time.Parse(time.RFC3339, "2025-09-26T14:36:23Z")
	defer clock.Freeze(refTime)()
	if err != nil {
		t.Fatalf("Failed to parse reference time: %v", err)
	}

	localTime, err := time.Parse("2006-01-02T15:04:05", txId["LclDtTm"].(string))
	if err != nil {
		t.Fatalf("Failed to parse local time: %v", err)
	}

	message := NewMessageBuilder(cardId, last4, expiry).WithTransactionIDs(ids).WithLocalTime(localTime).BuildBillPayCredit(amountCents, billingAmountCents)

	actualJson, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to serialize ISO20022Message into JSON: %v", err)
	}
	var actual map[string]any
	err = json.Unmarshal(actualJson, &actual)
	if err != nil {
		t.Fatalf("Failed to parse actualJson into map: %v", err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Payload mismatch (-expected +actual):\n%s", diff)
	}
}

func TestBuildP2PSend(t *testing.T) {
	// jq -r '.item[] | select(.name=="CARD_TRANSFER") | .request.body.raw' DPS_Txns_APIs.postman_collection.json
	const netxdJSON = `{"Hdr":{"MsgFctn":"REQU","CreDtTm":"2025-09-26T14:36:33Z","PrtcolVrsn":"2.49.1","InitgPty":{"Id":"VisaDPS"}},"Body":{"Tx":{"AcctFr":{"AcctTp":"00"},"TxTp":"10","TxId":{"AcqrrRefData":"412345     ","LifeCyclTracIdData":{"Id":"992064955384422"},"SysTracAudtNb":"S897358","LclDtTm":"2024-09-10T08:53:29","CardIssrRefData":"{\"CARD-ID\":\"v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14\",\"ProcessorLifeCycleId\":\"5c02ca46-166b-4e69-9128-090d1f5d4753\",\"ProcessorTransactionId\":\"e7026d4fad9b586d0a42ca8badb7180087a81472127cb97f1854b0c5bd06834a\"}","TrnsmssnDtTm":"2025-09-26T14:36:33Z","RtrvlRefNb":"R25092605370"},"AddtlData":[{"Val":"0","Tp":"VAUResult"},{"Val":"N","Tp":"CryptoPurchInd"},{"Val":"N","Tp":"DeferredAuthInd"}],"AddtlAmts":[{"Amt":{"Ccy":"840","Amt":0,"Sgn":true},"Tp":"OTHP"}],"AltrnMsgRsn":["0031"],"TxAmts":{"RcncltnAmt":{"QtnDt":"2024-09-11T00:00:00Z","XchgRate":1,"Ccy":"840","Amt":13.5},"AmtQlfr":"ACTL","TxAmt":{"Ccy":"840","Amt":13.5},"DtldAmt":[{"Amt":{"Amt":13.5},"OthrTp":"BASE","Tp":"OTHP"}],"CrdhldrBllgAmt":{"XchgRate":1,"Ccy":"840","Amt":2}}},"Cntxt":{"Vrfctn":[{"VrfctnRslt":[{"RsltDtls":[{"Val":"2","Tp":"CAVV Result Code"}],"Tp":"AuthenticationValue","Rslt":"SUCC"}],"SubTp":"Visa","VrfctnInf":[{"Val":{"TxtVal":"Z"},"Tp":"Method"}],"Tp":"THDS"},{"VrfctnRslt":[{"RsltDtls":[{"Val":"0","Tp":"CAMReliability"}]}]}],"TxCntxt":{"MrchntCtgyCd":"5921","SttlmSvc":{"SttlmSvcApld":{"Tp":"VISAInternational"}},"CardPrgrmm":{"CardPrgrmmApld":{"Id":"VSN"}},"Rcncltn":{"Dt":"2024-09-10","Id":"VISAInternational"}},"PtOfSvcCntxt":{"CardDataNtryMd":"KEEN","CrdhldrActvtd":true,"EComrcInd":true,"EComrcData":[{"Val":"5","Tp":"ECI"}]},"RskCntxt":[{"RskInptData":[{"Val":"3","Tp":"FalconScoreSource"},{"Val":"0469","Tp":"FalconScoreValue"},{"Val":"0","Tp":"FalconRespCode"},{"Val":"08","Tp":"FalconReason1"},{"Val":"06","Tp":"FalconReason2"},{"Val":"02","Tp":"FalconReason3"},{"Val":"09","Tp":"VisaRiskScore"},{"Val":"5A","Tp":"VisaRiskReason"},{"Val":"02","Tp":"VisaRiskCondCode1"},{"Val":"C2","Tp":"VisaRiskCondCode2"},{"Val":"00","Tp":"VisaRiskCondCode3"}]}]},"Envt":{"Accptr":{"NmAndLctn":"DNH*GODADDY#3521500092 TEMPE        AZUS","Id":"Authorization  ","Adr":{"PstlCd":"         ","Ctry":"USA"}},"Termnl":{"Cpblties":{"CardRdngCpblty":["MLEY"],"CrdhldrVrfctnCpblty":[{"Cpblty":"UNSP"}]},"TermnlId":{"Id":"CNP8    "},"Tp":"POST"},"Acqrr":{"Ctry":"840","Id":"59000000204"},"Sndr":{"Id":"11111111111"},"Card":{"PmtAcctRef":"V0010010000000000000000000001","XpryDt":"2709","PAN":"9999999999992059","CardPdctTp":"F"}},"PrcgRslt":{"RsltData":{"RsltDtls":"00"},"ApprvlData":{"ApprvlCd":"250191"}},"SplmtryData":[{"Envlp":{"DPSProcessorData":{"LclDtTmOrgtr":"Y"}}}]}}`

	var expected map[string]any
	if err := json.Unmarshal([]byte(netxdJSON), &expected); err != nil {
		t.Fatalf("Failed to parse netxd JSON: %v", err)
	}

	body := expected["Body"].(map[string]any)
	tx := body["Tx"].(map[string]any)
	txId := tx["TxId"].(map[string]any)

	cardId := "v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14"
	last4 := "2059"
	expiry := "2709"
	amountCents := 1350

	txAmts := tx["TxAmts"].(map[string]any)
	crdhldrBllgAmt := txAmts["CrdhldrBllgAmt"].(map[string]any)
	billingAmountCents := int(crdhldrBllgAmt["Amt"].(float64) * 100)

	ids := TransactionIDs{
		CardExternalId:         cardId,
		ProcessorLifeCycleId:   "5c02ca46-166b-4e69-9128-090d1f5d4753",
		ProcessorTransactionId: "e7026d4fad9b586d0a42ca8badb7180087a81472127cb97f1854b0c5bd06834a",
		LifeCycleId:            txId["LifeCyclTracIdData"].(map[string]any)["Id"].(string),
		SysTracAudtNb:          txId["SysTracAudtNb"].(string),
		RtrvlRefNb:             txId["RtrvlRefNb"].(string),
		AcqrrRefData:           "412345     ",
	}

	refTime, err := time.Parse(time.RFC3339, "2025-09-26T14:36:33Z")
	defer clock.Freeze(refTime)()
	if err != nil {
		t.Fatalf("Failed to parse reference time: %v", err)
	}

	localTime, err := time.Parse("2006-01-02T15:04:05", txId["LclDtTm"].(string))
	if err != nil {
		t.Fatalf("Failed to parse local time: %v", err)
	}

	message := NewMessageBuilder(cardId, last4, expiry).WithTransactionIDs(ids).WithLocalTime(localTime).BuildP2PSend(amountCents, billingAmountCents)

	actualJson, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to serialize ISO20022Message into JSON: %v", err)
	}
	var actual map[string]any
	err = json.Unmarshal(actualJson, &actual)
	if err != nil {
		t.Fatalf("Failed to parse actualJson into map: %v", err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Payload mismatch (-expected +actual):\n%s", diff)
	}
}

func TestBuildP2PReceive(t *testing.T) {
	// jq -r '.item[] | select(.name=="PREPAID_CARD_DEPOSIT") | .request.body.raw' DPS_Txns_APIs.postman_collection.json
	const netxdJSON = `{"Hdr":{"MsgFctn":"REQU","CreDtTm":"2025-09-26T14:36:26Z","PrtcolVrsn":"2.49.1","InitgPty":{"Id":"VisaDPS"}},"Body":{"Tx":{"AcctFr":{"AcctTp":"00"},"TxTp":"28","TxId":{"AcqrrRefData":"412345     ","LifeCyclTracIdData":{"Id":"992064955384422"},"SysTracAudtNb":"S897355","LclDtTm":"2024-09-10T08:53:29","CardIssrRefData":"{\"CARD-ID\":\"v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14\",\"ProcessorLifeCycleId\":\"1d0bae2d-ffd9-4a1e-afa6-fca50fc12dbf\",\"ProcessorTransactionId\":\"1174a7d7b33ab980aa3c68c7176e22ec495aae4b0d14a7fc439eae662f2126de\"}","TrnsmssnDtTm":"2025-09-26T14:36:26Z","RtrvlRefNb":"R25092617915"},"AddtlData":[{"Val":"0","Tp":"VAUResult"},{"Val":"N","Tp":"CryptoPurchInd"},{"Val":"N","Tp":"DeferredAuthInd"}],"AddtlAmts":[{"Amt":{"Ccy":"840","Amt":0,"Sgn":true},"Tp":"OTHP"}],"AltrnMsgRsn":["0031"],"TxAmts":{"RcncltnAmt":{"QtnDt":"2024-09-11T00:00:00Z","XchgRate":1,"Ccy":"840","Amt":13.5},"AmtQlfr":"ACTL","TxAmt":{"Ccy":"840","Amt":13.5},"DtldAmt":[{"Amt":{"Amt":13.5},"OthrTp":"BASE","Tp":"OTHP"}],"CrdhldrBllgAmt":{"XchgRate":1,"Ccy":"840","Amt":11}}},"Cntxt":{"Vrfctn":[{"VrfctnRslt":[{"RsltDtls":[{"Val":"2","Tp":"CAVV Result Code"}],"Tp":"AuthenticationValue","Rslt":"SUCC"}],"SubTp":"Visa","VrfctnInf":[{"Val":{"TxtVal":"Z"},"Tp":"Method"}],"Tp":"THDS"},{"VrfctnRslt":[{"RsltDtls":[{"Val":"0","Tp":"CAMReliability"}]}]}],"TxCntxt":{"MrchntCtgyCd":"5921","SttlmSvc":{"SttlmSvcApld":{"Tp":"VISAInternational"}},"CardPrgrmm":{"CardPrgrmmApld":{"Id":"VSN"}},"Rcncltn":{"Dt":"2024-09-10","Id":"VISAInternational"}},"PtOfSvcCntxt":{"CardDataNtryMd":"KEEN","CrdhldrActvtd":true,"EComrcInd":true,"EComrcData":[{"Val":"5","Tp":"ECI"}]},"RskCntxt":[{"RskInptData":[{"Val":"3","Tp":"FalconScoreSource"},{"Val":"0469","Tp":"FalconScoreValue"},{"Val":"0","Tp":"FalconRespCode"},{"Val":"08","Tp":"FalconReason1"},{"Val":"06","Tp":"FalconReason2"},{"Val":"02","Tp":"FalconReason3"},{"Val":"09","Tp":"VisaRiskScore"},{"Val":"5A","Tp":"VisaRiskReason"},{"Val":"02","Tp":"VisaRiskCondCode1"},{"Val":"C2","Tp":"VisaRiskCondCode2"},{"Val":"00","Tp":"VisaRiskCondCode3"}]}]},"Envt":{"Accptr":{"NmAndLctn":"DNH*GODADDY#3521500092 TEMPE        AZUS","Id":"Authorization  ","Adr":{"PstlCd":"         ","Ctry":"USA"}},"Termnl":{"Cpblties":{"CardRdngCpblty":["MLEY"],"CrdhldrVrfctnCpblty":[{"Cpblty":"UNSP"}]},"TermnlId":{"Id":"CNP8    "},"Tp":"POST"},"Acqrr":{"Ctry":"840","Id":"59000000204"},"Sndr":{"Id":"11111111111"},"Card":{"PmtAcctRef":"V0010010000000000000000000001","XpryDt":"2709","PAN":"9999999999992059","CardPdctTp":"F"}},"PrcgRslt":{"RsltData":{"RsltDtls":"00"},"ApprvlData":{"ApprvlCd":"250191"}},"SplmtryData":[{"Envlp":{"DPSProcessorData":{"LclDtTmOrgtr":"Y"}}}]}}`

	var expected map[string]any
	if err := json.Unmarshal([]byte(netxdJSON), &expected); err != nil {
		t.Fatalf("Failed to parse netxd JSON: %v", err)
	}

	body := expected["Body"].(map[string]any)
	tx := body["Tx"].(map[string]any)
	txId := tx["TxId"].(map[string]any)

	cardId := "v-401-f89af56d-41e0-46a1-85ac-af00abd6ab14"
	last4 := "2059"
	expiry := "2709"
	amountCents := 1350

	txAmts := tx["TxAmts"].(map[string]any)
	crdhldrBllgAmt := txAmts["CrdhldrBllgAmt"].(map[string]any)
	billingAmountCents := int(crdhldrBllgAmt["Amt"].(float64) * 100)

	ids := TransactionIDs{
		CardExternalId:         cardId,
		ProcessorLifeCycleId:   "1d0bae2d-ffd9-4a1e-afa6-fca50fc12dbf",
		ProcessorTransactionId: "1174a7d7b33ab980aa3c68c7176e22ec495aae4b0d14a7fc439eae662f2126de",
		LifeCycleId:            txId["LifeCyclTracIdData"].(map[string]any)["Id"].(string),
		SysTracAudtNb:          txId["SysTracAudtNb"].(string),
		RtrvlRefNb:             txId["RtrvlRefNb"].(string),
		AcqrrRefData:           "412345     ",
	}

	refTime, err := time.Parse(time.RFC3339, "2025-09-26T14:36:26Z")
	defer clock.Freeze(refTime)()
	if err != nil {
		t.Fatalf("Failed to parse reference time: %v", err)
	}

	localTime, err := time.Parse("2006-01-02T15:04:05", txId["LclDtTm"].(string))
	if err != nil {
		t.Fatalf("Failed to parse local time: %v", err)
	}

	message := NewMessageBuilder(cardId, last4, expiry).WithTransactionIDs(ids).WithLocalTime(localTime).BuildP2PReceive(amountCents, billingAmountCents)

	actualJson, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to serialize ISO20022Message into JSON: %v", err)
	}
	var actual map[string]any
	err = json.Unmarshal(actualJson, &actual)
	if err != nil {
		t.Fatalf("Failed to parse actualJson into map: %v", err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Payload mismatch (-expected +actual):\n%s", diff)
	}
}
