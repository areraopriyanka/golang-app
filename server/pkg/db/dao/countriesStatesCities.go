package dao

type CountriesDao struct {
	CountryCode string `json:"code" gorm:"column:country_code"`
	CountryName string `json:"name" gorm:"column:country_name"`
}

// TableName overrides the table name used by CountriesDao to `countries`
func (CountriesDao) TableName() string {
	return "countries"
}

type StatesDao struct {
	StateCode string `json:"stateCode" gorm:"column:state_code"`
	StateName string `json:"stateName" gorm:"column:state_name"`
}

// TableName overrides the table name used by StatesDao to `states`
func (StatesDao) TableName() string {
	return "states"
}

type CitiesDao struct {
	CityName string `json:"cityName" gorm:"column:city_name"`
}

// TableName overrides the table name used by CitiesDao to `cities`
func (CitiesDao) TableName() string {
	return "cities"
}
