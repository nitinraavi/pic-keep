package controllers

import "net/mail"

func valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
