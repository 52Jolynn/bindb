package mod

type BinData struct {
	IinStart     uint32    `json:"iin_start"`
	IinEnd       uint32    `json:"iin_end"`
	NumberLength uint8     `json:"number_length"`
	NumberLuhn   string    `json:"number_luhn"`
	Prepaid      string    `json:"prepaid"`
	Status       BinStatus `json:"status"`
	BaseBinData
}

type BaseBinData struct {
	Schema   string `json:"schema"`    //mastercard, visa, unionpay, etc
	Brand    string `json:"brand"`     //
	CardType string `json:"card_type"` //卡类型, debit or credit
	Country  string `json:"country"`   //国家, 英文名称
	BankName string `json:"bank_name"` //银行, 英文名称
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
