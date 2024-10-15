package tools

import "regexp"

var nameRegex *regexp.Regexp

func init() {
	r, err := regexp.Compile("^[0-9A-Za-z_-]+(.[0-9A-Za-z_-]+)*$")
	if err != nil {
		panic(err)
	}
	nameRegex = r
}

// ValidateName is used to validate group, key, selector keypair and other fields
func ValidateName(str string) bool {
	return nameRegex.MatchString(str)
}
