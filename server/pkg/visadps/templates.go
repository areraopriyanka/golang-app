package visadps

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"process-api/pkg/clock"
	"strings"
	"time"
)

type ISO20022Message struct {
	Hdr  Header          `json:"Hdr"`
	Body TransactionBody `json:"Body"`
}

type Header struct {
	MsgFctn    string          `json:"MsgFctn"`
	CreDtTm    string          `json:"CreDtTm"`
	PrtcolVrsn string          `json:"PrtcolVrsn"`
	InitgPty   InitiatingParty `json:"InitgPty"`
}

type InitiatingParty struct {
	Id string `json:"Id"`
}

type TransactionBody struct {
	Tx          Transaction         `json:"Tx"`
	Cntxt       Context             `json:"Cntxt"`
	Envt        Environment         `json:"Envt"`
	PrcgRslt    *ProcessingResult   `json:"PrcgRslt,omitempty"`
	SplmtryData []SupplementaryData `json:"SplmtryData,omitempty"`
}

type Transaction struct {
	AcctFr           *Account                 `json:"AcctFr,omitempty"`
	AcctTo           *Account                 `json:"AcctTo,omitempty"`
	TxTp             string                   `json:"TxTp"`
	TxId             TransactionId            `json:"TxId"`
	AddtlData        []AdditionalData         `json:"AddtlData,omitempty"`
	FndsSvcs         *FundsServices           `json:"FndsSvcs,omitempty"`
	AddtlAmts        []AdditionalAmount       `json:"AddtlAmts,omitempty"`
	SpclPrgrmmQlfctn []SpecialProgrammeQualif `json:"SpclPrgrmmQlfctn,omitempty"`
	AltrnMsgRsn      []string                 `json:"AltrnMsgRsn,omitempty"`
	TxAmts           *TransactionAmounts      `json:"TxAmts,omitempty"`
}

type Account struct {
	AcctTp string `json:"AcctTp"`
}

type TransactionId struct {
	AcqrrRefData       string        `json:"AcqrrRefData,omitempty"`
	LifeCyclTracIdData LifeCycleData `json:"LifeCyclTracIdData"`
	SysTracAudtNb      string        `json:"SysTracAudtNb"`
	LclDtTm            string        `json:"LclDtTm"`
	CardIssrRefData    string        `json:"CardIssrRefData"`
	TrnsmssnDtTm       string        `json:"TrnsmssnDtTm"`
	RtrvlRefNb         string        `json:"RtrvlRefNb"`
}

type LifeCycleData struct {
	Id string `json:"Id"`
}

type AdditionalData struct {
	Val string `json:"Val"`
	Tp  string `json:"Tp"`
}

type FundsServices struct {
	FndgSvc FundingService `json:"FndgSvc"`
}

type FundingService struct {
	Desc    string `json:"Desc"`
	BizPurp string `json:"BizPurp"`
}

type SpecialProgrammeQualif struct {
	Dtl []SpecialProgrammeDetail `json:"Dtl,omitempty"`
}

type SpecialProgrammeDetail struct {
	Val string `json:"Val"`
	Nm  string `json:"Nm"`
}

type Payer struct {
	Cstmr PayerCustomer `json:"Cstmr"`
}

type PayerCustomer struct {
	Crdntls []PayerCredential `json:"Crdntls"`
	Adr     PayerAddress      `json:"Adr"`
	Nm      string            `json:"Nm"`
}

type PayerCredential struct {
	IdVal    string `json:"IdVal"`
	OthrIdCd string `json:"OthrIdCd"`
	IdCd     string `json:"IdCd"`
}

type PayerAddress struct {
	CtrySubDvsnMjr string `json:"CtrySubDvsnMjr"`
	Ctry           string `json:"Ctry"`
	AdrLine1       string `json:"AdrLine1"`
	TwnNm          string `json:"TwnNm"`
}

type AdditionalAmount struct {
	Amt AmountDetail `json:"Amt"`
	Tp  string       `json:"Tp"`
}

type AmountDetail struct {
	Ccy string  `json:"Ccy"`
	Amt float64 `json:"Amt"`
	Sgn bool    `json:"Sgn"`
}

type TransactionAmounts struct {
	RcncltnAmt     ReconciliationAmount `json:"RcncltnAmt"`
	AmtQlfr        string               `json:"AmtQlfr"`
	TxAmt          SimpleAmount         `json:"TxAmt"`
	DtldAmt        []DetailedAmount     `json:"DtldAmt"`
	CrdhldrBllgAmt SimpleAmount         `json:"CrdhldrBllgAmt"`
}

type ReconciliationAmount struct {
	QtnDt    string  `json:"QtnDt"`
	XchgRate float64 `json:"XchgRate"`
	Ccy      string  `json:"Ccy"`
	Amt      float64 `json:"Amt"`
}

type SimpleAmount struct {
	XchgRate float64 `json:"XchgRate,omitempty"`
	Ccy      string  `json:"Ccy"`
	Amt      float64 `json:"Amt"`
}

type DetailedAmount struct {
	Amt    SimpleAmountNoCcy `json:"Amt"`
	OthrTp string            `json:"OthrTp,omitempty"`
	Tp     string            `json:"Tp"`
}

type SimpleAmountNoCcy struct {
	Amt float64 `json:"Amt"`
}

type Context struct {
	Vrfctn       []Verification        `json:"Vrfctn,omitempty"`
	TxCntxt      TransactionContext    `json:"TxCntxt"`
	PtOfSvcCntxt PointOfServiceContext `json:"PtOfSvcCntxt"`
	RskCntxt     any                   `json:"RskCntxt,omitempty"`
}

type Verification struct {
	SubTp      string               `json:"SubTp,omitempty"`
	VrfctnInf  []VerificationInfo   `json:"VrfctnInf,omitempty"`
	VrfctnRslt []VerificationResult `json:"VrfctnRslt,omitempty"`
	Tp         string               `json:"Tp,omitempty"`
}

type VerificationInfo struct {
	Val ValueWrapper `json:"Val"`
	Tp  string       `json:"Tp"`
}

type ValueWrapper struct {
	TxtVal string `json:"TxtVal"`
}

type VerificationResult struct {
	Rslt     string         `json:"Rslt,omitempty"`
	RsltDtls []ResultDetail `json:"RsltDtls,omitempty"`
	Tp       string         `json:"Tp,omitempty"`
}

type ResultDetail struct {
	Val string `json:"Val"`
	Tp  string `json:"Tp"`
}

type TransactionContext struct {
	MrchntCtgyCd string            `json:"MrchntCtgyCd"`
	SttlmSvc     SettlementService `json:"SttlmSvc"`
	CardPrgrmm   CardProgramme     `json:"CardPrgrmm"`
	Rcncltn      Reconciliation    `json:"Rcncltn"`
}

type SettlementService struct {
	SttlmSvcApld SettlementServiceApplied `json:"SttlmSvcApld"`
}

type SettlementServiceApplied struct {
	Tp string `json:"Tp"`
}

type CardProgramme struct {
	CardPrgrmmApld CardProgrammeApplied `json:"CardPrgrmmApld"`
}

type CardProgrammeApplied struct {
	Id string `json:"Id"`
}

type Reconciliation struct {
	Dt string `json:"Dt"`
	Id string `json:"Id"`
}

type PointOfServiceContext struct {
	CrdhldrPres    bool   `json:"CrdhldrPres,omitempty"`
	AttnddInd      bool   `json:"AttnddInd,omitempty"`
	CardDataNtryMd string `json:"CardDataNtryMd"`
	CardPres       bool   `json:"CardPres,omitempty"`
	CrdhldrActvtd  bool   `json:"CrdhldrActvtd,omitempty"`
	EComrcInd      bool   `json:"EComrcInd,omitempty"`
	EComrcData     []any  `json:"EComrcData,omitempty"`
}

type Environment struct {
	Accptr Acceptor `json:"Accptr"`
	Termnl Terminal `json:"Termnl"`
	Acqrr  Acquirer `json:"Acqrr"`
	Sndr   Sender   `json:"Sndr"`
	Card   Card     `json:"Card"`
	Pyer   *Payer   `json:"Pyer,omitempty"`
}

type Acceptor struct {
	NmAndLctn string  `json:"NmAndLctn"`
	ShrtNm    string  `json:"ShrtNm,omitempty"`
	Id        string  `json:"Id"`
	Adr       Address `json:"Adr"`
}

type Address struct {
	PstlCd         string `json:"PstlCd"`
	CtrySubDvsnMjr string `json:"CtrySubDvsnMjr,omitempty"`
	Ctry           string `json:"Ctry"`
}

type Terminal struct {
	Cpblties Capabilities `json:"Cpblties"`
	TermnlId TerminalId   `json:"TermnlId"`
	OthrTp   string       `json:"OthrTp,omitempty"`
	Tp       string       `json:"Tp"`
}

type Capabilities struct {
	CardRdngCpblty      []string                           `json:"CardRdngCpblty"`
	CrdhldrVrfctnCpblty []CardholderVerificationCapability `json:"CrdhldrVrfctnCpblty"`
}

type CardholderVerificationCapability struct {
	Cpblty string `json:"Cpblty"`
}

type TerminalId struct {
	Id string `json:"Id"`
}

type Acquirer struct {
	Ctry string `json:"Ctry"`
	Id   string `json:"Id"`
}

type Sender struct {
	Id string `json:"Id"`
}

type Card struct {
	PmtAcctRef string `json:"PmtAcctRef"`
	XpryDt     string `json:"XpryDt"`
	PAN        string `json:"PAN"`
	CardPdctTp string `json:"CardPdctTp,omitempty"`
}

type ProcessingResult struct {
	RsltData   ResultData   `json:"RsltData"`
	ApprvlData ApprovalData `json:"ApprvlData"`
}

type ResultData struct {
	RsltDtls string `json:"RsltDtls"`
}

type ApprovalData struct {
	ApprvlCd string `json:"ApprvlCd"`
}

type SupplementaryData struct {
	Envlp Envelope `json:"Envlp"`
}

type Envelope struct {
	DPSProcessorData DPSProcessorData `json:"DPSProcessorData"`
}

type DPSProcessorData struct {
	LclDtTmOrgtr string `json:"LclDtTmOrgtr"`
}

type CardIssuerRefData struct {
	CardId                 string `json:"CARD-ID"`
	ProcessorLifeCycleId   string `json:"ProcessorLifeCycleId"`
	ProcessorTransactionId string `json:"ProcessorTransactionId"`
}

type TransactionIDs struct {
	CardExternalId         string
	ProcessorLifeCycleId   string
	ProcessorTransactionId string
	LifeCycleId            string
	SysTracAudtNb          string
	RtrvlRefNb             string
	AcqrrRefData           string
}

type MessageBuilder struct {
	cardExternalId string
	last4          string
	expiryDate     string
	transactionIDs *TransactionIDs
	localTime      *time.Time
}

func NewMessageBuilder(cardExternalId, last4, expiryDate string) MessageBuilder {
	return MessageBuilder{
		cardExternalId: cardExternalId,
		last4:          last4,
		expiryDate:     expiryDate,
	}
}

func (b MessageBuilder) WithTransactionIDs(ids TransactionIDs) MessageBuilder {
	if ids.AcqrrRefData == "" {
		ids.AcqrrRefData = "412345     "
	}
	b.transactionIDs = &ids
	return b
}

func (b MessageBuilder) WithLocalTime(t time.Time) MessageBuilder {
	b.localTime = &t
	return b
}

func (b MessageBuilder) getLocalTime() time.Time {
	if b.localTime != nil {
		return *b.localTime
	}
	return clock.Now()
}

func (b MessageBuilder) lclDtTm() string {
	return formatLclDtTm(b.getLocalTime())
}

func (b MessageBuilder) trnsmssnDtTm() string {
	return formatTrnsmssnDtTm(clock.Now())
}

func (b MessageBuilder) txId() TransactionId {
	ids := b.getTransactionIDs()
	return buildTransactionId(ids, b.lclDtTm(), b.trnsmssnDtTm())
}

func (b MessageBuilder) getTransactionIDs() TransactionIDs {
	if b.transactionIDs != nil {
		ids := *b.transactionIDs
		if ids.CardExternalId == "" {
			ids.CardExternalId = b.cardExternalId
		}
		return ids
	}
	return generateTransactionIDs(b.cardExternalId, clock.Now())
}

func (b MessageBuilder) card() Card {
	return Card{
		PmtAcctRef: "V0010010000000000000000000001",
		XpryDt:     formatExpiryDate(b.expiryDate),
		PAN:        maskPAN(b.last4),
	}
}

func (b MessageBuilder) cardWithProductType(productType string) Card {
	card := b.card()
	card.CardPdctTp = productType
	return card
}

func standardAddtlData() []AdditionalData {
	return []AdditionalData{
		{Val: "0", Tp: "VAUResult"},
		{Val: "N", Tp: "CryptoPurchInd"},
		{Val: "N", Tp: "DeferredAuthInd"},
	}
}

func buildRskCntxt(falconSource, falconValue, falconReason1, falconReason2, falconReason3 string) []map[string]any {
	return []map[string]any{
		{
			"RskInptData": []map[string]string{
				{"Val": falconSource, "Tp": "FalconScoreSource"},
				{"Val": falconValue, "Tp": "FalconScoreValue"},
				{"Val": "0", "Tp": "FalconRespCode"},
				{"Val": falconReason1, "Tp": "FalconReason1"},
				{"Val": falconReason2, "Tp": "FalconReason2"},
				{"Val": falconReason3, "Tp": "FalconReason3"},
				{"Val": "09", "Tp": "VisaRiskScore"},
				{"Val": "5A", "Tp": "VisaRiskReason"},
				{"Val": "02", "Tp": "VisaRiskCondCode1"},
				{"Val": "C2", "Tp": "VisaRiskCondCode2"},
				{"Val": "00", "Tp": "VisaRiskCondCode3"},
			},
		},
	}
}

func generateTransactionIDs(cardExternalId string, now time.Time) TransactionIDs {
	return TransactionIDs{
		CardExternalId:         cardExternalId,
		ProcessorLifeCycleId:   generateUUID(),
		ProcessorTransactionId: generateProcessorTransactionId(),
		LifeCycleId:            generateLifeCycleId(),
		SysTracAudtNb:          generateSysTracAudtNb(),
		RtrvlRefNb:             generateRetrievalRefNb(now),
		AcqrrRefData:           "412345     ",
	}
}

func buildTransactionId(ids TransactionIDs, lclDtTm, trnsmssnDtTm string) TransactionId {
	cardIssuerRef := CardIssuerRefData{
		CardId:                 ids.CardExternalId,
		ProcessorLifeCycleId:   ids.ProcessorLifeCycleId,
		ProcessorTransactionId: ids.ProcessorTransactionId,
	}
	cardIssuerRefJSON, err := json.Marshal(cardIssuerRef)
	if err != nil {
		panic(fmt.Sprintf("json.Marshal failure: %v", err))
	}

	return TransactionId{
		AcqrrRefData:       ids.AcqrrRefData,
		LifeCyclTracIdData: LifeCycleData{Id: ids.LifeCycleId},
		SysTracAudtNb:      ids.SysTracAudtNb,
		LclDtTm:            lclDtTm,
		CardIssrRefData:    string(cardIssuerRefJSON),
		TrnsmssnDtTm:       trnsmssnDtTm,
		RtrvlRefNb:         ids.RtrvlRefNb,
	}
}

func generateRandomDigits(n int) string {
	result := make([]byte, n)
	for i := range n {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			panic(fmt.Sprintf("crypto/rand failure: %v", err))
		}
		result[i] = byte('0' + num.Int64())
	}
	return string(result)
}

func generateLifeCycleId() string {
	return generateRandomDigits(15)
}

func generateSysTracAudtNb() string {
	return "S" + generateRandomDigits(6)
}

func generateRetrievalRefNb(now time.Time) string {
	return fmt.Sprintf("R%02d%02d%02d%s", now.Year()%100, int(now.Month()), now.Day(), generateRandomDigits(5))
}

func generateUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failure: %v", err))
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func generateProcessorTransactionId() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failure: %v", err))
	}
	return fmt.Sprintf("%x", b)
}

func maskPAN(last4 string) string {
	return "999999999999" + last4
}

func formatExpiryDate(expiryDate string) string {
	parts := strings.Split(expiryDate, "/")
	if len(parts) == 2 {
		return parts[0] + parts[1]
	}
	return expiryDate
}

func formatCreDtTmWithMillis(t time.Time) string {
	ms := t.UnixMilli() % 1000
	if ms == 0 {
		return t.UTC().Format(time.RFC3339)
	}
	return fmt.Sprintf("%s.%03dZ", t.UTC().Format("2006-01-02T15:04:05"), ms)
}

func formatLclDtTm(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
}

func formatTrnsmssnDtTm(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func BuildATMWithdrawal(cardExternalId, last4, expiryDate string, amountCents, cardholderBillingAmountCents int) ISO20022Message {
	return NewMessageBuilder(cardExternalId, last4, expiryDate).BuildATMWithdrawal(amountCents, cardholderBillingAmountCents)
}

func (b MessageBuilder) BuildATMWithdrawal(amountCents, cardholderBillingAmountCents int) ISO20022Message {
	localTime := b.getLocalTime()
	amount := float64(amountCents) / 100.0
	billingAmount := float64(cardholderBillingAmountCents) / 100.0

	return ISO20022Message{
		Hdr: buildHeader("REQU", clock.Now()),
		Body: TransactionBody{
			Tx: Transaction{
				AcctFr:    &Account{AcctTp: "00"},
				TxTp:      "01",
				TxId:      b.txId(),
				AddtlData: standardAddtlData(),
				AddtlAmts: []AdditionalAmount{
					{Amt: AmountDetail{Ccy: "840", Amt: 0, Sgn: true}, Tp: "OTHP"},
				},
				AltrnMsgRsn: []string{"0041"},
				TxAmts: &TransactionAmounts{
					RcncltnAmt: ReconciliationAmount{
						QtnDt:    localTime.Format("2006-01-02") + "T00:00:00Z",
						XchgRate: 1,
						Ccy:      "840",
						Amt:      amount,
					},
					AmtQlfr: "ACTL",
					TxAmt:   SimpleAmount{Ccy: "840", Amt: amount},
					DtldAmt: []DetailedAmount{
						{Amt: SimpleAmountNoCcy{Amt: amount}, OthrTp: "BASE", Tp: "OTHP"},
					},
					CrdhldrBllgAmt: SimpleAmount{XchgRate: 1, Ccy: "840", Amt: billingAmount},
				},
			},
			Cntxt: Context{
				Vrfctn: []Verification{
					{VrfctnRslt: []VerificationResult{{Rslt: "SUCC"}}, Tp: "NPIN"},
					{VrfctnRslt: []VerificationResult{{RsltDtls: []ResultDetail{{Val: "0", Tp: "CAMReliability"}}}}, Tp: ""},
					{VrfctnInf: []VerificationInfo{{Val: ValueWrapper{TxtVal: "Z"}, Tp: "Method"}}, Tp: "THDS"},
					{VrfctnRslt: []VerificationResult{{Rslt: "SUCC"}}, Tp: "NVSC"},
				},
				TxCntxt: TransactionContext{
					MrchntCtgyCd: "5411",
					SttlmSvc:     SettlementService{SttlmSvcApld: SettlementServiceApplied{Tp: "VISAInternational"}},
					CardPrgrmm:   CardProgramme{CardPrgrmmApld: CardProgrammeApplied{Id: "VSV"}},
					Rcncltn:      Reconciliation{Dt: localTime.AddDate(0, 0, 1).Format("2006-01-02"), Id: "VISAInternational"},
				},
				PtOfSvcCntxt: PointOfServiceContext{
					CrdhldrPres:    true,
					CardDataNtryMd: "MGST",
					CardPres:       true,
					CrdhldrActvtd:  true,
				},
				RskCntxt: buildRskCntxt("2", "0165", "06", "08", "10"),
			},
			Envt: Environment{
				Accptr: Acceptor{
					NmAndLctn: "4 HOUSTON TEXANS CIR HOUSTON TXUS",
					ShrtNm:    "Cash Withdrawal",
					Id:        "               ",
					Adr:       Address{PstlCd: "80129    ", CtrySubDvsnMjr: "08", Ctry: "USA"},
				},
				Termnl: Terminal{
					Cpblties: Capabilities{
						CardRdngCpblty:      []string{"MGST"},
						CrdhldrVrfctnCpblty: []CardholderVerificationCapability{{Cpblty: "NPIN"}},
					},
					TermnlId: TerminalId{Id: "ATM2    "},
					Tp:       "ATMT",
				},
				Acqrr: Acquirer{Ctry: "840", Id: "59992960009"},
				Sndr:  Sender{Id: "11111111111"},
				Card:  b.card(),
			},
			PrcgRslt: &ProcessingResult{
				RsltData:   ResultData{RsltDtls: "00"},
				ApprvlData: ApprovalData{ApprvlCd: "250191"},
			},
			SplmtryData: []SupplementaryData{
				{Envlp: Envelope{DPSProcessorData: DPSProcessorData{LclDtTmOrgtr: "Y"}}},
			},
		},
	}
}

func BuildATMDeposit(cardExternalId, last4, expiryDate string, amountCents, cardholderBillingAmountCents int) ISO20022Message {
	return NewMessageBuilder(cardExternalId, last4, expiryDate).BuildATMDeposit(amountCents, cardholderBillingAmountCents)
}

func (b MessageBuilder) BuildATMDeposit(amountCents, cardholderBillingAmountCents int) ISO20022Message {
	localTime := b.getLocalTime()
	amount := float64(amountCents) / 100.0
	billingAmount := float64(cardholderBillingAmountCents) / 100.0

	return ISO20022Message{
		Hdr: buildHeader("REQU", clock.Now()),
		Body: TransactionBody{
			Tx: Transaction{
				AcctTo:    &Account{AcctTp: "00"},
				TxTp:      "21",
				TxId:      b.txId(),
				AddtlData: standardAddtlData(),
				AddtlAmts: []AdditionalAmount{
					{Amt: AmountDetail{Ccy: "840", Amt: 0, Sgn: true}, Tp: "OTHP"},
				},
				SpclPrgrmmQlfctn: []SpecialProgrammeQualif{
					{
						Dtl: []SpecialProgrammeDetail{
							{Val: "875", Nm: "FPI"},
						},
					},
				},
				TxAmts: &TransactionAmounts{
					RcncltnAmt: ReconciliationAmount{
						QtnDt:    localTime.AddDate(0, 0, 1).Format("2006-01-02") + "T00:00:00Z",
						XchgRate: 1,
						Ccy:      "840",
						Amt:      amount,
					},
					AmtQlfr: "ACTL",
					TxAmt:   SimpleAmount{Ccy: "840", Amt: amount},
					DtldAmt: []DetailedAmount{
						{Amt: SimpleAmountNoCcy{Amt: amount}, OthrTp: "BASE", Tp: "OTHP"},
					},
					CrdhldrBllgAmt: SimpleAmount{XchgRate: 1, Ccy: "840", Amt: billingAmount},
				},
			},
			Cntxt: Context{
				Vrfctn: []Verification{
					{VrfctnRslt: []VerificationResult{{Rslt: "SUCC"}}, Tp: "NPIN"},
					{VrfctnRslt: []VerificationResult{{RsltDtls: []ResultDetail{{Val: "0", Tp: "CAMReliability"}}}}, Tp: ""},
					{VrfctnInf: []VerificationInfo{{Val: ValueWrapper{TxtVal: "Z"}, Tp: "Method"}}, Tp: "THDS"},
					{VrfctnRslt: []VerificationResult{{Rslt: "SUCC"}}, Tp: "NVSC"},
				},
				TxCntxt: TransactionContext{
					MrchntCtgyCd: "6011",
					SttlmSvc:     SettlementService{SttlmSvcApld: SettlementServiceApplied{Tp: "VISAInternational"}},
					CardPrgrmm:   CardProgramme{CardPrgrmmApld: CardProgrammeApplied{Id: "VSP"}},
					Rcncltn:      Reconciliation{Dt: localTime.Format("2006-01-02"), Id: "VISAInternational"},
				},
				PtOfSvcCntxt: PointOfServiceContext{
					CrdhldrPres:    true,
					CardDataNtryMd: "MGST",
					CardPres:       true,
					CrdhldrActvtd:  true,
				},
				RskCntxt: buildRskCntxt("3", "0029", "34", "14", "77"),
			},
			Envt: Environment{
				Accptr: Acceptor{
					NmAndLctn: "4 HOUSTON TEXANS CIR HOUSTON TXUS",
					ShrtNm:    "Deposit",
					Id:        "               ",
					Adr:       Address{PstlCd: "80129    ", CtrySubDvsnMjr: "08", Ctry: "USA"},
				},
				Termnl: Terminal{
					Cpblties: Capabilities{
						CardRdngCpblty:      []string{"MGST"},
						CrdhldrVrfctnCpblty: []CardholderVerificationCapability{{Cpblty: "NPIN"}},
					},
					TermnlId: TerminalId{Id: "PLUS8   "},
					Tp:       "ATMT",
				},
				Acqrr: Acquirer{Ctry: "840", Id: "59992960009"},
				Sndr:  Sender{Id: "11111111111"},
				Card:  b.card(),
			},
			SplmtryData: []SupplementaryData{
				{Envlp: Envelope{DPSProcessorData: DPSProcessorData{LclDtTmOrgtr: "Y"}}},
			},
		},
	}
}

func (b MessageBuilder) buildEcommerceBase(amountCents, cardholderBillingAmountCents int,
	msgFctn, txType, mcc string,
) ISO20022Message {
	localTime := b.getLocalTime()
	amount := float64(amountCents) / 100.0
	billingAmount := float64(cardholderBillingAmountCents) / 100.0

	return ISO20022Message{
		Hdr: buildHeader(msgFctn, clock.Now()),
		Body: TransactionBody{
			Tx: Transaction{
				AcctFr:    &Account{AcctTp: "00"},
				TxTp:      txType,
				TxId:      b.txId(),
				AddtlData: standardAddtlData(),
				AddtlAmts: []AdditionalAmount{
					{Amt: AmountDetail{Ccy: "840", Amt: 0, Sgn: true}, Tp: "OTHP"},
				},
				AltrnMsgRsn: []string{"0031"},
				TxAmts: &TransactionAmounts{
					RcncltnAmt: ReconciliationAmount{
						QtnDt:    localTime.AddDate(0, 0, 1).Format("2006-01-02") + "T00:00:00Z",
						XchgRate: 1,
						Ccy:      "840",
						Amt:      amount,
					},
					AmtQlfr: "ACTL",
					TxAmt:   SimpleAmount{Ccy: "840", Amt: amount},
					DtldAmt: []DetailedAmount{
						{Amt: SimpleAmountNoCcy{Amt: amount}, OthrTp: "BASE", Tp: "OTHP"},
					},
					CrdhldrBllgAmt: SimpleAmount{XchgRate: 1, Ccy: "840", Amt: billingAmount},
				},
			},
			Cntxt: Context{
				Vrfctn: []Verification{
					{
						VrfctnRslt: []VerificationResult{{
							RsltDtls: []ResultDetail{{Val: "2", Tp: "CAVV Result Code"}},
							Tp:       "AuthenticationValue",
							Rslt:     "SUCC",
						}},
						SubTp:     "Visa",
						VrfctnInf: []VerificationInfo{{Val: ValueWrapper{TxtVal: "Z"}, Tp: "Method"}},
						Tp:        "THDS",
					},
					{VrfctnRslt: []VerificationResult{{RsltDtls: []ResultDetail{{Val: "0", Tp: "CAMReliability"}}}}, Tp: ""},
				},
				TxCntxt: TransactionContext{
					MrchntCtgyCd: mcc,
					SttlmSvc:     SettlementService{SttlmSvcApld: SettlementServiceApplied{Tp: "VISAInternational"}},
					CardPrgrmm:   CardProgramme{CardPrgrmmApld: CardProgrammeApplied{Id: "VSN"}},
					Rcncltn:      Reconciliation{Dt: localTime.Format("2006-01-02"), Id: "VISAInternational"},
				},
				PtOfSvcCntxt: PointOfServiceContext{
					CardDataNtryMd: "KEEN",
					CrdhldrActvtd:  true,
					EComrcInd:      true,
					EComrcData:     []any{map[string]string{"Val": "5", "Tp": "ECI"}},
				},
				RskCntxt: buildRskCntxt("3", "0469", "08", "06", "02"),
			},
			Envt: Environment{
				Accptr: Acceptor{
					NmAndLctn: "DNH*GODADDY#3521500092 TEMPE        AZUS",
					Id:        "Authorization  ",
					Adr:       Address{PstlCd: "         ", Ctry: "USA"},
				},
				Termnl: Terminal{
					Cpblties: Capabilities{
						CardRdngCpblty:      []string{"MLEY"},
						CrdhldrVrfctnCpblty: []CardholderVerificationCapability{{Cpblty: "UNSP"}},
					},
					TermnlId: TerminalId{Id: "CNP8    "},
					Tp:       "POST",
				},
				Acqrr: Acquirer{Ctry: "840", Id: "59000000204"},
				Sndr:  Sender{Id: "11111111111"},
				Card:  b.cardWithProductType("F"),
			},
			PrcgRslt: &ProcessingResult{
				RsltData:   ResultData{RsltDtls: "00"},
				ApprvlData: ApprovalData{ApprvlCd: "250191"},
			},
			SplmtryData: []SupplementaryData{
				{Envlp: Envelope{DPSProcessorData: DPSProcessorData{LclDtTmOrgtr: "Y"}}},
			},
		},
	}
}

func BuildPurchase(cardExternalId, last4, expiryDate string, amountCents, cardholderBillingAmountCents int, mcc string) ISO20022Message {
	return NewMessageBuilder(cardExternalId, last4, expiryDate).BuildPurchase(amountCents, cardholderBillingAmountCents, mcc)
}

func (b MessageBuilder) BuildPurchase(amountCents, cardholderBillingAmountCents int, mcc string) ISO20022Message {
	return b.buildEcommerceBase(amountCents, cardholderBillingAmountCents,
		"REQU", "00", mcc)
}

func BuildMerchandiseReturn(cardExternalId, last4, expiryDate string, amountCents, cardholderBillingAmountCents int) ISO20022Message {
	return NewMessageBuilder(cardExternalId, last4, expiryDate).BuildMerchandiseReturn(amountCents, cardholderBillingAmountCents)
}

func (b MessageBuilder) BuildMerchandiseReturn(amountCents, cardholderBillingAmountCents int) ISO20022Message {
	return b.buildEcommerceBase(amountCents, cardholderBillingAmountCents,
		"REQU", "20", "5947")
}

func BuildBillPayDebit(cardExternalId, last4, expiryDate string, amountCents, cardholderBillingAmountCents int) ISO20022Message {
	return NewMessageBuilder(cardExternalId, last4, expiryDate).BuildBillPayDebit(amountCents, cardholderBillingAmountCents)
}

func (b MessageBuilder) BuildBillPayDebit(amountCents, cardholderBillingAmountCents int) ISO20022Message {
	return b.buildEcommerceBase(amountCents, cardholderBillingAmountCents,
		"ADVC", "50", "5947")
}

func (b MessageBuilder) buildBillPayCreditBase(amountCents, cardholderBillingAmountCents int) ISO20022Message {
	localTime := b.getLocalTime()
	amount := float64(amountCents) / 100.0
	billingAmount := float64(cardholderBillingAmountCents) / 100.0

	return ISO20022Message{
		Hdr: buildHeader("REQU", clock.Now()),
		Body: TransactionBody{
			Tx: Transaction{
				AcctTo: &Account{AcctTp: "00"},
				TxTp:   "55",
				TxId:   b.txId(),
				AddtlData: append(standardAddtlData(),
					AdditionalData{Val: "$PP", Tp: "MONEYXFERTYPE"},
				),
				FndsSvcs: &FundsServices{
					FndgSvc: FundingService{
						Desc:    "Person to person",
						BizPurp: "$PP",
					},
				},
				AddtlAmts: []AdditionalAmount{
					{Amt: AmountDetail{Ccy: "840", Amt: 0, Sgn: true}, Tp: "OTHP"},
				},
				SpclPrgrmmQlfctn: []SpecialProgrammeQualif{
					{
						Dtl: []SpecialProgrammeDetail{
							{Val: "421", Nm: "FPI"},
							{Val: "0", Nm: "RA"},
						},
					},
					{},
				},
				TxAmts: &TransactionAmounts{
					RcncltnAmt: ReconciliationAmount{
						QtnDt:    localTime.Format("2006-01-02") + "T00:00:00Z",
						XchgRate: 1,
						Ccy:      "840",
						Amt:      amount,
					},
					AmtQlfr: "ACTL",
					TxAmt:   SimpleAmount{Ccy: "840", Amt: amount},
					DtldAmt: []DetailedAmount{
						{Amt: SimpleAmountNoCcy{Amt: amount}, OthrTp: "BASE", Tp: "OTHP"},
					},
					CrdhldrBllgAmt: SimpleAmount{XchgRate: 1, Ccy: "840", Amt: billingAmount},
				},
			},
			Cntxt: Context{
				Vrfctn: []Verification{
					{VrfctnRslt: []VerificationResult{{RsltDtls: []ResultDetail{{Val: "0", Tp: "CAMReliability"}}}}, Tp: ""},
					{
						SubTp:     "Visa",
						VrfctnInf: []VerificationInfo{{Val: ValueWrapper{TxtVal: "Z"}, Tp: "Method"}},
						Tp:        "THDS",
					},
					{Tp: "CPSG"},
				},
				TxCntxt: TransactionContext{
					MrchntCtgyCd: "4829",
					SttlmSvc:     SettlementService{SttlmSvcApld: SettlementServiceApplied{Tp: "VISAInternational"}},
					CardPrgrmm:   CardProgramme{CardPrgrmmApld: CardProgrammeApplied{Id: "VSN"}},
					Rcncltn:      Reconciliation{Dt: localTime.Format("2006-01-02"), Id: "VISAInternational"},
				},
				PtOfSvcCntxt: PointOfServiceContext{
					CrdhldrPres:    true,
					AttnddInd:      true,
					CardDataNtryMd: "KEEN",
					CrdhldrActvtd:  true,
				},
				RskCntxt: buildRskCntxt("3", "0000", "00", "00", "00"),
			},
			Envt: Environment{
				Pyer: &Payer{
					Cstmr: PayerCustomer{
						Crdntls: []PayerCredential{
							{IdVal: "1234567890123456", OthrIdCd: "REFERENCE#", IdCd: "OTHP"},
						},
						Adr: PayerAddress{
							CtrySubDvsnMjr: "CO",
							Ctry:           "USA",
							AdrLine1:       "1340 Pennsylvania St",
							TwnNm:          "Denver",
						},
						Nm: "Molly Brown",
					},
				},
				Accptr: Acceptor{
					NmAndLctn: "PAYPAL 4029311133 CAUS",
					Id:        "Original Credit",
					Adr:       Address{PstlCd: "80203    ", CtrySubDvsnMjr: "08", Ctry: "USA"},
				},
				Termnl: Terminal{
					Cpblties: Capabilities{
						CardRdngCpblty:      []string{"UNSP"},
						CrdhldrVrfctnCpblty: []CardholderVerificationCapability{{Cpblty: "UNSP"}},
					},
					TermnlId: TerminalId{Id: "VD6     "},
					Tp:       "POST",
				},
				Acqrr: Acquirer{Ctry: "840", Id: "59000000204"},
				Sndr:  Sender{Id: "11111111111"},
				Card:  b.card(),
			},
			SplmtryData: []SupplementaryData{
				{Envlp: Envelope{DPSProcessorData: DPSProcessorData{LclDtTmOrgtr: "Y"}}},
			},
		},
	}
}

func BuildBillPayCredit(cardExternalId, last4, expiryDate string, amountCents, cardholderBillingAmountCents int) ISO20022Message {
	return NewMessageBuilder(cardExternalId, last4, expiryDate).BuildBillPayCredit(amountCents, cardholderBillingAmountCents)
}

func (b MessageBuilder) BuildBillPayCredit(amountCents, cardholderBillingAmountCents int) ISO20022Message {
	return b.buildBillPayCreditBase(amountCents, cardholderBillingAmountCents)
}

func BuildP2PSend(cardExternalId, last4, expiryDate string, amountCents, cardholderBillingAmountCents int) ISO20022Message {
	return NewMessageBuilder(cardExternalId, last4, expiryDate).BuildP2PSend(amountCents, cardholderBillingAmountCents)
}

func (b MessageBuilder) BuildP2PSend(amountCents, cardholderBillingAmountCents int) ISO20022Message {
	return b.buildEcommerceBase(amountCents, cardholderBillingAmountCents,
		"REQU", "10", "5921")
}

func BuildP2PReceive(cardExternalId, last4, expiryDate string, amountCents, cardholderBillingAmountCents int) ISO20022Message {
	return NewMessageBuilder(cardExternalId, last4, expiryDate).BuildP2PReceive(amountCents, cardholderBillingAmountCents)
}

func (b MessageBuilder) BuildP2PReceive(amountCents, cardholderBillingAmountCents int) ISO20022Message {
	return b.buildEcommerceBase(amountCents, cardholderBillingAmountCents,
		"REQU", "28", "5921")
}

func buildHeader(messageFunction string, now time.Time) Header {
	return Header{
		MsgFctn:    messageFunction,
		CreDtTm:    formatCreDtTmWithMillis(now),
		PrtcolVrsn: "2.49.1",
		InitgPty:   InitiatingParty{Id: "VisaDPS"},
	}
}
