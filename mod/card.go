package mod

type Bin struct {
	iinStart     uint32 `json:"iin_start"`
	iinEnd       uint32 `json:"iin_end"`
	numberLength uint8  `json:"number_length"`
	numberLuhn   string `json:"number_luhn"`
	schema       string `json:"schema"`    //mastercard, visa, unionpay, etc
	brand        string `json:"brand"`     //
	cardType     string `json:"card_type"` //卡类型, debit or credit
	prepaid      string `json:"prepaid"`
	country      string `json:"country"`      //国家, 英文名称
	countryCn    string `json:"country_cn"`   //国家, 中文名称
	bankName     string `json:"bank_name"`    //银行, 英文名称
	bankNameCn   string `json:"bank_name_cn"` //银行, 中文名称
}
