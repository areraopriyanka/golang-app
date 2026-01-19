package handler

import (
	"log/slog"
	"os"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/resource/mcc"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapCodeToMerchantCategory(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logging.Logger = logger

	mccTestCases := map[string]string{
		"3509": "Marriott",
		"3530": "Renaissance Hotels",
		"4121": "Taxicabs And Limousines",
		"4899": "Cable, Satellite And Other Pay Television/Radio/Streaming Services",
		"5311": "Department Stores",
		"5399": "Miscellaneous General Merchandise",
		"5411": "Grocery Stores And Supermarkets",
		"5462": "Bakeries",
		"5499": "Miscellaneous Food Stores - Convenience Stores And Specialty Markets",
		"5541": "Service Stations (With Or Without Ancillary Services)",
		"5542": "Automated Fuel Dispensers",
		"5552": "Electric Vehicle Charging",
		"5655": "Sports And Riding Apparel Stores",
		"5812": "Eating Places And Restaurants",
		"5813": "Drinking Places (Alcoholic Beverages) - Bars, Taverns, Nightclubs, Cocktail Lounges, And Discotheques",
		"5814": "Fast Food Restaurants",
		"5818": "Digital Goods – Large Digital Goods Merchant",
		"5912": "Drug Stores And Pharmacies",
		"5921": "Package Stores – Beer, Wine, And Liquor",
		"5941": "Sporting Goods Stores",
		"5942": "Book Stores",
		"5947": "Gift, Card, Novelty And Souvenir Shops",
		"5968": "Direct Marketing - Continuity/Subscription Merchant",
		"5999": "Miscellaneous And Specialty Retail Shops",
		"6010": "Financial Institutions - Manual Cash Disbursements",
		"6011": "Financial Institutions – Automated Cash Disbursements",
		"7311": "Advertising Services",
		"7523": "Parking Lots, Parking Meters And Garages",
		"7542": "Car Washes",
		"7800": "Government-Owned Lotteries (Us Region Only)",
		"8099": "Medical Services And Health Practitioners (Not Elsewhere Classified)",
	}

	t.Run("Return the expected merchant category for given mcc codes", func(t *testing.T) {
		for code, expectedCategory := range mccTestCases {
			category, found := mcc.GetCategory(code)
			assert.True(t, found)
			assert.Equal(t, expectedCategory, category)
		}
	})
}

func TestParsedCardAcceptor(t *testing.T) {
	tests := []struct {
		raw    string
		parsed CardAcceptor
	}{
		{
			raw: "100492136              GRANADA        ES",
			parsed: CardAcceptor{
				Merchant: "100492136",
				City:     "GRANADA",
				State:    "",
				Country:  "ES",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "11095 NE EVERGREEN     HILLSBORO    ORUS",
			parsed: CardAcceptor{
				Merchant: "11095 NE EVERGREEN",
				City:     "HILLSBORO",
				State:    "OR",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "11095 NE Evergreen PkwyHillsboro    ORUS",
			parsed: CardAcceptor{
				Merchant: "11095 NE Evergreen Pkwy",
				City:     "Hillsboro",
				State:    "OR",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "1139 WHITE HORSE RD    VOORHEES     NJUS",
			parsed: CardAcceptor{
				Merchant: "1139 WHITE HORSE RD",
				City:     "VOORHEES",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "1139 White Horse Rd    Voorhees     NJUS",
			parsed: CardAcceptor{
				Merchant: "1139 White Horse Rd",
				City:     "Voorhees",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "120 BROADWAY           DENVER       COUS",
			parsed: CardAcceptor{
				Merchant: "120 BROADWAY",
				City:     "DENVER",
				State:    "CO",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "15 HAVERHILL RD        AMESBURY     MAUS",
			parsed: CardAcceptor{
				Merchant: "15 HAVERHILL RD",
				City:     "AMESBURY",
				State:    "MA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "175 W ARMY TRAIL       GLENDALE     ILUS",
			parsed: CardAcceptor{
				Merchant: "175 W ARMY TRAIL",
				City:     "GLENDALE",
				State:    "IL",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "2 PARK  AVE            PORTLAND     MEUS",
			parsed: CardAcceptor{
				Merchant: "2 PARK  AVE",
				City:     "PORTLAND",
				State:    "ME",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "221 DELSEA DR N        GLASSBORO    NJUS",
			parsed: CardAcceptor{
				Merchant: "221 DELSEA DR N",
				City:     "GLASSBORO",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "2563 15TH ST STE 105   DENVER       COUS",
			parsed: CardAcceptor{
				Merchant: "2563 15TH ST STE 105",
				City:     "DENVER",
				State:    "CO",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "265 LAFAYETTE ROAD     SEABROOK     NHUS",
			parsed: CardAcceptor{
				Merchant: "265 LAFAYETTE ROAD",
				City:     "SEABROOK",
				State:    "NH",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "295 MAIN ST            BRIDGTON     MEUS",
			parsed: CardAcceptor{
				Merchant: "295 MAIN ST",
				City:     "BRIDGTON",
				State:    "ME",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "3000 S LAS VEGAS BLVD  LAS VEGAS    NVUS",
			parsed: CardAcceptor{
				Merchant: "3000 S LAS VEGAS BLVD",
				City:     "LAS VEGAS",
				State:    "NV",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "3025 S LAS VEGAS BL    LAS VEGAS    NVUS",
			parsed: CardAcceptor{
				Merchant: "3025 S LAS VEGAS BL",
				City:     "LAS VEGAS",
				State:    "NV",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "500 BELLVIEW AVE       CRESTED BUTTECOUS",
			parsed: CardAcceptor{
				Merchant: "500 BELLVIEW AVE",
				City:     "CRESTED BUTTE",
				State:    "CO",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "51 Wharf Street EMV    Portland     MEUS",
			parsed: CardAcceptor{
				Merchant: "51 Wharf Street EMV",
				City:     "Portland",
				State:    "ME",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "51450 - SPECTRUM 4     PHILADELPHIA PAUS",
			parsed: CardAcceptor{
				Merchant: "51450 - SPECTRUM 4",
				City:     "PHILADELPHIA",
				State:    "PA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "540 N SCHMALE RD       CAROL STREAM ILUS",
			parsed: CardAcceptor{
				Merchant: "540 N SCHMALE RD",
				City:     "CAROL STREAM",
				State:    "IL",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "600 DELSEA DR N        GLASSBORO    NJUS",
			parsed: CardAcceptor{
				Merchant: "600 DELSEA DR N",
				City:     "GLASSBORO",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "620 LAFAYETTE RD       HAMPTON      NHUS",
			parsed: CardAcceptor{
				Merchant: "620 LAFAYETTE RD",
				City:     "HAMPTON",
				State:    "NH",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "62522 9481 SPEEDWAY    ROSEBURG     ORUS",
			parsed: CardAcceptor{
				Merchant: "62522 9481 SPEEDWAY",
				City:     "ROSEBURG",
				State:    "OR",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "650 W. CUTHBERT BLV    HADDON TOWNSHNJUS",
			parsed: CardAcceptor{
				Merchant: "650 W. CUTHBERT BLV",
				City:     "HADDON TOWNSH",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "650 W. CUTHBERT BLVD   HADDON TOWNSHNJUS",
			parsed: CardAcceptor{
				Merchant: "650 W. CUTHBERT BLVD",
				City:     "HADDON TOWNSH",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "665 CONGRESS ST        PORTLAND     MEUS",
			parsed: CardAcceptor{
				Merchant: "665 CONGRESS ST",
				City:     "PORTLAND",
				State:    "ME",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "675 WOODBURY GLASSBORO SEWELL       NJUS",
			parsed: CardAcceptor{
				Merchant: "675 WOODBURY GLASSBORO",
				City:     "SEWELL",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "733 N MAIN ST          GLASSBORO    NJUS",
			parsed: CardAcceptor{
				Merchant: "733 N MAIN ST",
				City:     "GLASSBORO",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "9420 W LAKE MEAD BL    LAS VEGAS    NVUS",
			parsed: CardAcceptor{
				Merchant: "9420 W LAKE MEAD BL",
				City:     "LAS VEGAS",
				State:    "NV",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "ALIMENTACION LI        GRANADA        ES",
			parsed: CardAcceptor{
				Merchant: "ALIMENTACION LI",
				City:     "GRANADA",
				State:    "",
				Country:  "ES",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "AMAZON MKTPLACE PMTS   Amzn.com/billWAUS",
			parsed: CardAcceptor{
				Merchant: "AMAZON MKTPLACE PMTS",
				City:     "",
				State:    "WA",
				Country:  "US",
				Website:  "Amzn.com/bill",
				Phone:    "",
			},
		},
		{
			raw: "AMZN Digital           888-802-3080 WAUS",
			parsed: CardAcceptor{
				Merchant: "AMZN Digital",
				City:     "",
				State:    "WA",
				Country:  "US",
				Website:  "",
				Phone:    "888-802-3080",
			},
		},
		{
			raw: "Amazon.com             Amzn.com/billWAUS",
			parsed: CardAcceptor{
				Merchant: "Amazon.com",
				City:     "",
				State:    "WA",
				Country:  "US",
				Website:  "Amzn.com/bill",
				Phone:    "",
			},
		},
		{
			raw: "BACCI PIZZA - TAYLOR   Chicago      ILUS",
			parsed: CardAcceptor{
				Merchant: "BACCI PIZZA - TAYLOR",
				City:     "Chicago",
				State:    "IL",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "BIG APPLE #1026        PORTLAND     MEUS",
			parsed: CardAcceptor{
				Merchant: "BIG APPLE #1026",
				City:     "PORTLAND",
				State:    "ME",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "BLACKWOOD CARWASH      BLACKWOOD    NJUS",
			parsed: CardAcceptor{
				Merchant: "BLACKWOOD CARWASH",
				City:     "BLACKWOOD",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "BP#9173998LAKE BP INC. ROSELLE      ILUS",
			parsed: CardAcceptor{
				Merchant: "BP#9173998LAKE BP INC.",
				City:     "ROSELLE",
				State:    "IL",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "Big Head Vending, LLC  Greenwood VilCOUS",
			parsed: CardAcceptor{
				Merchant: "Big Head Vending, LLC",
				City:     "Greenwood Vil",
				State:    "CO",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "Borough of Glassboro   Glassboro    NJUS",
			parsed: CardAcceptor{
				Merchant: "Borough of Glassboro",
				City:     "Glassboro",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "CELLARDOOR WINERY      PORTLAND     MEUS",
			parsed: CardAcceptor{
				Merchant: "CELLARDOOR WINERY",
				City:     "PORTLAND",
				State:    "ME",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "CHARGEPOINT 5          CAMPBELL     CAUS",
			parsed: CardAcceptor{
				Merchant: "CHARGEPOINT 5",
				City:     "CAMPBELL",
				State:    "CA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "CHEVRON 0210567        STOCKBRIDGE  GAUS",
			parsed: CardAcceptor{
				Merchant: "CHEVRON 0210567",
				City:     "STOCKBRIDGE",
				State:    "GA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "CIRCLE K 07501         AMESBURY     MAUS",
			parsed: CardAcceptor{
				Merchant: "CIRCLE K 07501",
				City:     "AMESBURY",
				State:    "MA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "CPI*CASCO BAY FOOD AND LEWISTON     MEUS",
			parsed: CardAcceptor{
				Merchant: "CPI*CASCO BAY FOOD AND",
				City:     "LEWISTON",
				State:    "ME",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "CRESCENT MOON COFFEE & MULLICA HILL NJUS",
			parsed: CardAcceptor{
				Merchant: "CRESCENT MOON COFFEE &",
				City:     "MULLICA HILL",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "CVS/PHARMACY #10 10879-Fort Worth   TXUS",
			parsed: CardAcceptor{
				Merchant: "CVS/PHARMACY #10 10879-",
				City:     "Fort Worth",
				State:    "TX",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "Casco Bay Food n Bev   Lewiston     MEUS",
			parsed: CardAcceptor{
				Merchant: "Casco Bay Food n Bev",
				City:     "Lewiston",
				State:    "ME",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "DELAWARE NORTH LOGAN F EAST BOSTON  MAUS",
			parsed: CardAcceptor{
				Merchant: "DELAWARE NORTH LOGAN F",
				City:     "EAST BOSTON",
				State:    "MA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "DELTA AIRLINES ONBOARD ATLANTA      GAUS",
			parsed: CardAcceptor{
				Merchant: "DELTA AIRLINES ONBOARD",
				City:     "ATLANTA",
				State:    "GA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "DELUCA'S AUTO WASH     SALISBURY    MAUS",
			parsed: CardAcceptor{
				Merchant: "DELUCA'S AUTO WASH",
				City:     "SALISBURY",
				State:    "MA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "EL CORTE INGLES DEPARTAGRANADA        ES",
			parsed: CardAcceptor{
				Merchant: "EL CORTE INGLES DEPARTA",
				City:     "GRANADA",
				State:    "",
				Country:  "ES",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "ELITE SPORTSWEAR       READING      PAUS",
			parsed: CardAcceptor{
				Merchant: "ELITE SPORTSWEAR",
				City:     "READING",
				State:    "PA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "GELATISSIMO ARTISAN GE NEW CANAAN   CTUS",
			parsed: CardAcceptor{
				Merchant: "GELATISSIMO ARTISAN GE",
				City:     "NEW CANAAN",
				State:    "CT",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "GOOGLE *ANDROID TEMP   cc@google.comCAUS",
			parsed: CardAcceptor{
				Merchant: "GOOGLE *ANDROID TEMP",
				City:     "",
				State:    "CA",
				Country:  "US",
				Website:  "google.com",
				Phone:    "",
			},
		},
		{
			raw: "HANNAFORD #8351        PORTLAND     MEUS",
			parsed: CardAcceptor{
				Merchant: "HANNAFORD #8351",
				City:     "PORTLAND",
				State:    "ME",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "HOMETOWN PANTRY        BLOOMINGDALE ILUS",
			parsed: CardAcceptor{
				Merchant: "HOMETOWN PANTRY",
				City:     "BLOOMINGDALE",
				State:    "IL",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "INOVIA LLC  HILLSBORO  HILLSBORO    ORUS",
			parsed: CardAcceptor{
				Merchant: "INOVIA LLC  HILLSBORO",
				City:     "HILLSBORO",
				State:    "OR",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "JOE'S SUPER VARIETY    PORTLAND     MEUS",
			parsed: CardAcceptor{
				Merchant: "JOE'S SUPER VARIETY",
				City:     "PORTLAND",
				State:    "ME",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "Jones Landing          Portland     MEUS",
			parsed: CardAcceptor{
				Merchant: "Jones Landing",
				City:     "Portland",
				State:    "ME",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "LA CALLE TAQUERIA      DENVER       COUS",
			parsed: CardAcceptor{
				Merchant: "LA CALLE TAQUERIA",
				City:     "DENVER",
				State:    "CO",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "LOGAN PKG MASSPORT EMV EAST BOSTON  MAUS",
			parsed: CardAcceptor{
				Merchant: "LOGAN PKG MASSPORT EMV",
				City:     "EAST BOSTON",
				State:    "MA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "MARRIOTT SAVANNAH RIVE SAVANNAH     GAUS",
			parsed: CardAcceptor{
				Merchant: "MARRIOTT SAVANNAH RIVE",
				City:     "SAVANNAH",
				State:    "GA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "Meijer Express 198     BLOOMINGDALE ILUS",
			parsed: CardAcceptor{
				Merchant: "Meijer Express 198",
				City:     "BLOOMINGDALE",
				State:    "IL",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "Nyx*8336322778 ELECTRIFHUNT VALLEY  MDUS",
			parsed: CardAcceptor{
				Merchant: "Nyx*8336322778 ELECTRIF",
				City:     "HUNT VALLEY",
				State:    "MD",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "PNC BANK CASH ADV 312  WEST DEPTFORDNJUS",
			parsed: CardAcceptor{
				Merchant: "PNC BANK CASH ADV 312",
				City:     "WEST DEPTFORD",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "PP*Delish Cakes        BLOOMINGDALE ILUS",
			parsed: CardAcceptor{
				Merchant: "PP*Delish Cakes",
				City:     "BLOOMINGDALE",
				State:    "IL",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "QVC*9257672755         800-367-9444 PAUS",
			parsed: CardAcceptor{
				Merchant: "QVC*9257672755",
				City:     "",
				State:    "PA",
				Country:  "US",
				Website:  "",
				Phone:    "800-367-9444",
			},
		},
		{
			raw: "REI #50 LAKEWOOD       LAKEWOOD     COUS",
			parsed: CardAcceptor{
				Merchant: "REI #50 LAKEWOOD",
				City:     "LAKEWOOD",
				State:    "CO",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "RENAISSANCE ATLANTA    ATLANTA      GAUS",
			parsed: CardAcceptor{
				Merchant: "RENAISSANCE ATLANTA",
				City:     "ATLANTA",
				State:    "GA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "SAV Savannah News      Savannah     GAUS",
			parsed: CardAcceptor{
				Merchant: "SAV Savannah News",
				City:     "Savannah",
				State:    "GA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "SHAWS 2491             NEWBURYPORT  MAUS",
			parsed: CardAcceptor{
				Merchant: "SHAWS 2491",
				City:     "NEWBURYPORT",
				State:    "MA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "SHELL/SHELL            ENGLEWOOD    COUS",
			parsed: CardAcceptor{
				Merchant: "SHELL/SHELL",
				City:     "ENGLEWOOD",
				State:    "CO",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "SHOPRITE GLASSBORO S1  GLASSBORO    NJUS",
			parsed: CardAcceptor{
				Merchant: "SHOPRITE GLASSBORO S1",
				City:     "GLASSBORO",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "SHOPRITE OF MULLICA HILMULLICA HIL  NJUS",
			parsed: CardAcceptor{
				Merchant: "SHOPRITE OF MULLICA HIL",
				City:     "MULLICA HIL",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "SPLASH AND DASH        BEND         ORUS",
			parsed: CardAcceptor{
				Merchant: "SPLASH AND DASH",
				City:     "BEND",
				State:    "OR",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "SPO*HIGHGROUNDSCOFFEEROGLASSBORO    NJUS",
			parsed: CardAcceptor{
				Merchant: "SPO*HIGHGROUNDSCOFFEERO",
				City:     "GLASSBORO",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "SPOTIFY                NEW YORK     NYUS",
			parsed: CardAcceptor{
				Merchant: "SPOTIFY",
				City:     "NEW YORK",
				State:    "NY",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "SQ *LONE PINE COFFEE R Bend         ORUS",
			parsed: CardAcceptor{
				Merchant: "SQ *LONE PINE COFFEE R",
				City:     "Bend",
				State:    "OR",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "SQ *PORRON CELLARS     Hood River   ORUS",
			parsed: CardAcceptor{
				Merchant: "SQ *PORRON CELLARS",
				City:     "Hood River",
				State:    "OR",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "SQ *THE SPARROW BAKERY Portland     ORUS",
			parsed: CardAcceptor{
				Merchant: "SQ *THE SPARROW BAKERY",
				City:     "Portland",
				State:    "OR",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "STARBUCKS 14026        BEAVERTON    ORUS",
			parsed: CardAcceptor{
				Merchant: "STARBUCKS 14026",
				City:     "BEAVERTON",
				State:    "OR",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "SUNOCO 0509739900      SALISBURY    MAUS",
			parsed: CardAcceptor{
				Merchant: "SUNOCO 0509739900",
				City:     "SALISBURY",
				State:    "MA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "Subway 7885            Portland     ORUS",
			parsed: CardAcceptor{
				Merchant: "Subway 7885",
				City:     "Portland",
				State:    "OR",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "TASTY CORNER CHINESE R CHICAGO      ILUS",
			parsed: CardAcceptor{
				Merchant: "TASTY CORNER CHINESE R",
				City:     "CHICAGO",
				State:    "IL",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "TAXI GRANADA LIC. 172  GRANADA        ES",
			parsed: CardAcceptor{
				Merchant: "TAXI GRANADA LIC. 172",
				City:     "GRANADA",
				State:    "",
				Country:  "ES",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "THE DUSTLAND BAR       LAS VEGAS    NVUS",
			parsed: CardAcceptor{
				Merchant: "THE DUSTLAND BAR",
				City:     "LAS VEGAS",
				State:    "NV",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "THE STORE AND DELI     CRESTED BUTTECOUS",
			parsed: CardAcceptor{
				Merchant: "THE STORE AND DELI",
				City:     "CRESTED BUTTE",
				State:    "CO",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "THE THIRSTY PIG        Portland     MEUS",
			parsed: CardAcceptor{
				Merchant: "THE THIRSTY PIG",
				City:     "Portland",
				State:    "ME",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "THORNTONS #221345      HANOVER PARK ILUS",
			parsed: CardAcceptor{
				Merchant: "THORNTONS #221345",
				City:     "HANOVER PARK",
				State:    "IL",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "TROLLEY STOP GIFTS     SAVANNAH     GAUS",
			parsed: CardAcceptor{
				Merchant: "TROLLEY STOP GIFTS",
				City:     "SAVANNAH",
				State:    "GA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "TST* FORTUNE BAR       AMESBURY     MAUS",
			parsed: CardAcceptor{
				Merchant: "TST* FORTUNE BAR",
				City:     "AMESBURY",
				State:    "MA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "TST* LE PAIN QUOTIDIEN RYE          CTUS",
			parsed: CardAcceptor{
				Merchant: "TST* LE PAIN QUOTIDIEN",
				City:     "RYE",
				State:    "CT",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "TST* SILVER STAMP      LAS VEGAS    NVUS",
			parsed: CardAcceptor{
				Merchant: "TST* SILVER STAMP",
				City:     "LAS VEGAS",
				State:    "NV",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "Virgin Valley Cab      LAS VEGAS    NVUS",
			parsed: CardAcceptor{
				Merchant: "Virgin Valley Cab",
				City:     "LAS VEGAS",
				State:    "NV",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "WALGREENS STORE 15 HAVEAMESBURY     MAUS",
			parsed: CardAcceptor{
				Merchant: "WALGREENS STORE 15 HAVE",
				City:     "AMESBURY",
				State:    "MA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "WALGREENS STORE 223 NE BEND         ORUS",
			parsed: CardAcceptor{
				Merchant: "WALGREENS STORE 223 NE",
				City:     "BEND",
				State:    "OR",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "WALGREENS STORE 270 W LBLOOMINGDALE ILUS",
			parsed: CardAcceptor{
				Merchant: "WALGREENS STORE 270 W L",
				City:     "BLOOMINGDALE",
				State:    "IL",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "WALGREENS STORE 300 W BSAVANNAH     GAUS",
			parsed: CardAcceptor{
				Merchant: "WALGREENS STORE 300 W B",
				City:     "SAVANNAH",
				State:    "GA",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "WALGREENS STORE 9420 W LAS VEGAS    NVUS",
			parsed: CardAcceptor{
				Merchant: "WALGREENS STORE 9420 W",
				City:     "LAS VEGAS",
				State:    "NV",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "WAWA 8336              CHERRY HILL  NJUS",
			parsed: CardAcceptor{
				Merchant: "WAWA 8336",
				City:     "CHERRY HILL",
				State:    "NJ",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
		{
			raw: "WMT PLUS Aug 2025      BENTONVILLE  ARUS",
			parsed: CardAcceptor{
				Merchant: "WMT PLUS Aug 2025",
				City:     "BENTONVILLE",
				State:    "AR",
				Country:  "US",
				Website:  "",
				Phone:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			result := parseCardAcceptor(tt.raw)

			assert.Equal(t, tt.parsed, *result,
				"parsedValue should match expectedParsedValue")
		})
	}
}

func TestMergeTransactions(t *testing.T) {
	preAuth := ledger.ListTransactionsByAccountResultTransaction{
		ProcessID: "txn-123",
		Type:      "PRE_AUTH",
		TimeStamp: "2025-10-24T08:06:01Z",
		InstructedAmount: struct {
			Amount   int64  `json:"amount"`
			Currency string `json:"currency"`
		}{Amount: 1359, Currency: "USD"},
	}

	completion := ledger.ListTransactionsByAccountResultTransaction{
		ProcessID: "txn-123",
		Type:      "COMPLETION",
		TimeStamp: "2025-10-24T08:07:00Z",
		InstructedAmount: struct {
			Amount   int64  `json:"amount"`
			Currency string `json:"currency"`
		}{Amount: 1359, Currency: "USD"},
	}

	merged, completionTimestamps := MergeTransactions([]ledger.ListTransactionsByAccountResultTransaction{preAuth, completion})

	assert.Equal(t, 1, len(merged))                                          // only one transaction remains after merging
	assert.Equal(t, "COMPLETION", merged[0].Type)                            // the final transaction type is "COMPLETION"
	assert.Equal(t, "2025-10-24T08:06:01Z", merged[0].TimeStamp)             // timestamp is from PRE_AUTH, not COMPLETION
	assert.Equal(t, "2025-10-24T08:06:01Z", completionTimestamps["txn-123"]) // completionTimestamps map stores the same timestamp as the final merged transaction
}
