package model

type TgInlineButton struct {
	DisplayName string
	Value       string
}

type TgRowButtons []TgInlineButton

type UserDataRecord struct {
	UserID int64
	Form   string
}

type UserDataFormRecord struct {
	Form string
}
