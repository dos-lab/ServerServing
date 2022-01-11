package service

import "regexp"

type Validator struct{}

var validator = Validator{}

func (v Validator) ValidateAccountPassword(pwd string) bool {
	reg := regexp.MustCompile(`^[a-zA-Z][0-9a-zA-Z~!@#$%^&*?]{5,14}$`)
	return reg.MatchString(pwd)
}
