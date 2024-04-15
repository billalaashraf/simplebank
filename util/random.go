package util

import (
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomString(n int) string {
	var sb strings.Builder
	length := len(alphabet)
	for i := 0; i < n; i++ {
		char := alphabet[rand.Intn(length)]
		sb.WriteByte(char)
	}
	return sb.String()
}

func RandomOwner() string {
	return RandomString(6)
}

func RandomMoney(min int64, max int64) pgtype.Numeric {
	amount := RandomInt(min, max)
	return FromIntToPgNumeric(amount)
}

func FromIntToPgNumeric(amount int64) pgtype.Numeric {
	var rt pgtype.Numeric
	strValue := strconv.Itoa(int(amount));
	err := rt.Scan(strValue)
	if err != nil {
		log.Fatal(err)
	}
	return rt
}

func RandomCurrency() string {
	currencies := []string{"USD", "EUR", "CAD"}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}