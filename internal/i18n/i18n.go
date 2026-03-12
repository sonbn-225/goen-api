package i18n

import (
	"strings"
)

type Bundle map[string]string

var bundleEN = Bundle{
	"trade_buy":  "Trade buy",
	"trade_sell": "Trade sell",
	"trade_fee":  "Trade fee",
	"trade_tax":  "Trade tax",
	"savings_name": "Savings",
	"savings_deposit": "Savings deposit",
	"rotating_savings_prefix": "RotatingSavings",
}

var bundleVI = Bundle{
	"trade_buy":  "Giao dịch mua",
	"trade_sell": "Giao dịch bán",
	"trade_fee":  "Phí giao dịch",
	"trade_tax":  "Thuế giao dịch",
	"savings_name": "Tiết kiệm",
	"savings_deposit": "Gửi tiết kiệm",
	"rotating_savings_prefix": "Hụi/Họ",
}

func T(lang, key string) string {
	lang = strings.ToLower(lang)
	if strings.HasPrefix(lang, "vi") {
		if v, ok := bundleVI[key]; ok {
			return v
		}
	}
	if v, ok := bundleEN[key]; ok {
		return v
	}
	return key
}
