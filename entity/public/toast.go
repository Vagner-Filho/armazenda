package entity_public

import (
	"encoding/json"
)

type ToastType uint8

const (
	SuccessToast ToastType = iota
	WarningToast
	ErrorToast
	InfoToast
)

type Toast struct {
	Message string
	Hint    string
	Type    ToastType
}

func makeToast(message string, hint string, toastType ToastType) Toast {
	return Toast{
		Message: message,
		Hint:    hint,
		Type:    toastType,
	}
}

func GetSuccessToast(message string, hint string) Toast {
	return makeToast(message, hint, SuccessToast)
}

func GetWarningToast(message string, hint string) Toast {
	return makeToast(message, hint, WarningToast)
}

func GetErrorToast(message string, hint string) Toast {
	return makeToast(message, hint, ErrorToast)
}

func GetInfoToast(message string, hint string) Toast {
	return makeToast(message, hint, InfoToast)
}

func (t *Toast) ToJson() []byte {
	marshalled, _ := json.Marshal(map[string]Toast{"toast": *t})
	return marshalled
}
