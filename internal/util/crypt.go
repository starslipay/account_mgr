package util

import (
	"crypto/md5"
	"fmt"
	"strconv"
)

func GenMD5(input string) string {
	return fmt.Sprintf("%X", md5.Sum([]byte(input)))
}

func GenC2BDeductToken(transaction_id string, merchant_uid, uid, amount int64) string {
	md5Str := GenMD5(transaction_id + "|" + strconv.FormatInt(merchant_uid, 10) +
		"|" + strconv.FormatInt(uid, 10) + "|" + strconv.FormatInt(amount, 10))
	return md5Str
}
