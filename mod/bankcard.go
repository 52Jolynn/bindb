package mod

type BinData struct {
	Id           int64     `json:"id"`
	IinStart     uint32    `json:"iin_start"`
	IinEnd       uint32    `json:"iin_end"`
	NumberLength int8      `json:"number_length"`
	NumberLuhn   string    `json:"number_luhn"`
	Prepaid      string    `json:"prepaid"`
	Status       BinStatus `json:"status"`
	BaseBinData
}

type BaseBinData struct {
	Schema    string `json:"schema"`     //mastercard, visa, unionpay, etc
	Brand     string `json:"brand"`      //运营商名称
	CardType  string `json:"card_type"`  //卡类型, debit or credit
	Country   string `json:"country"`    //国家, 英文名称
	BankName  string `json:"bank_name"`  //银行, 英文名称
	BankLogo  string `json:"bank_logo"`  //银行logo, url
	BankUrl   string `json:"bank_url"`   //银行官网
	BankPhone string `json:"bank_phone"` //银行服务电话
	BankCity  string `json:"bank_city"`  //银行所在城市
}

type SimpleBinData struct {
	BaseBinData
	BankNameCn string `json:"bank_name_cn"` //银行, 中文名称
	CountryCn  string `json:"country_cn"`   //国家, 中文名称
}

type BinStatus uint8

const (
	//近似
	BinStatusApproximate = 1
	//确切
	BinStatusTruly = 2
)
