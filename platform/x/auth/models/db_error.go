package models

import "errors"

var XErrChkDeviceNotFound = errors.New("checkDevice not found user")
var XErrAuthUsernameNotFound = errors.New("Auth did not find user name")
var XErrAuthUserPasswordInCorrect = errors.New("Username or Password error!")
var XErrLoginAuthtokenNotFound = errors.New("AuthToken did not find, in login process!")
var XErrGetGateNotExist = errors.New("Gate Not Exist")
var XErrByBan = errors.New("user ban login by GM")
