package database

type Account string

func NewAccount(value string) Account {
	return Account(value)
}

type Tx struct {
	From  Account `json:"from"`
	To    Account `json:"to"`
	Value uint    `json:"value"`
	Data  string  `json:"data"`
}

func NewTx(from, to Account, val uint, data string) Tx {
	return Tx{
		From:  from,
		To:    to,
		Value: val,
		Data:  data,
	}
}

func (t Tx) IsReward() bool {
	return t.Data == "reward"
}
