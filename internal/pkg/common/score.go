package common

import (
	"fmt"
	"math"
	"strings"
)

type Score struct {
	Mantissa int64 `bson:"mantissa"`
	Exponent int   `bson:"exponent"`
}

func (u *Score) GetSaleScore(sale float64) *Score {
	if sale <= 0 || sale >= 1 {
		return nil
	}
	discountedMantissa := float64(u.Mantissa) * (1 - sale)

	discountedScore := &Score{
		Mantissa: int64(discountedMantissa),
		Exponent: u.Exponent,
	}

	discountedScore.normalize()
	return discountedScore
}

func (u *Score) GetFormattedScore() string {
	if u.Exponent < 3 {
		return formatSmallNumber(u.Mantissa * int64Pow(10, u.Exponent))
	}

	sign := ""
	if u.Mantissa < 0 {
		sign = "-"
	}

	mantissaStr := fmt.Sprintf("%d", absInt64(u.Mantissa))
	mantissaLength := len(mantissaStr)

	if mantissaLength > 1 {
		u.Exponent += mantissaLength - 1
	}

	mantissa := float64(absInt64(u.Mantissa)) / math.Pow(10, float64(mantissaLength-1))
	mantissaStr = fmt.Sprintf("%.2f", mantissa)
	mantissaStr = strings.TrimRight(mantissaStr, "0")
	mantissaStr = strings.TrimRight(mantissaStr, ".")

	return fmt.Sprintf("%s%se%d", sign, mantissaStr, u.Exponent)
}

func formatSmallNumber(n int64) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}

	mod := math.Abs(float64(n))
	switch {
	case mod >= 1e15:
		return fmt.Sprintf("%s%.2fP", sign, float64(n)/1e15) // Квадриллионы (P)
	case mod >= 1e12:
		return fmt.Sprintf("%s%.2fT", sign, float64(n)/1e12) // Триллионы (T)
	case mod >= 1e9:
		return fmt.Sprintf("%s%.2fB", sign, float64(n)/1e9) // Миллиарды (B)
	case mod >= 1e6:
		return fmt.Sprintf("%s%.2fM", sign, float64(n)/1e6) // Миллионы (M)
	case mod >= 1e3:
		return fmt.Sprintf("%s%.2fK", sign, float64(n)/1e3) // Тысячи (K)
	default:
		return fmt.Sprintf("%s%d", sign, n) // Менее тысячи
	}
}

func int64Pow(base, exp int) int64 {
	result := int64(1)
	for exp > 0 {
		result *= int64(base)
		exp--
	}
	return result
}

func (u *Score) Increase(increment uint64) {
	u.Mantissa += int64(increment)
	u.normalize()
}

func (u *Score) Decrease(decrement uint64) {
	u.Mantissa -= int64(decrement)
	u.normalize()
}

func (u *Score) Minus(other *Score) {
	if u.Exponent > other.Exponent {
		diff := u.Exponent - other.Exponent
		otherMantissa := other.Mantissa * int64Pow(10, diff)
		u.Mantissa -= otherMantissa
	} else if u.Exponent < other.Exponent {
		diff := other.Exponent - u.Exponent
		uMantissa := u.Mantissa * int64Pow(10, diff)
		u.Mantissa = uMantissa - other.Mantissa
		u.Exponent = other.Exponent
	} else {
		u.Mantissa -= other.Mantissa
	}
	u.normalize()
}

func (u *Score) Plus(other *Score) {
	if u.Exponent > other.Exponent {
		diff := u.Exponent - other.Exponent
		otherMantissa := other.Mantissa * int64Pow(10, diff)
		u.Mantissa += otherMantissa
	} else if u.Exponent < other.Exponent {
		diff := other.Exponent - u.Exponent
		uMantissa := u.Mantissa * int64Pow(10, diff)
		u.Mantissa = uMantissa + other.Mantissa
		u.Exponent = other.Exponent
	} else {
		u.Mantissa += other.Mantissa
	}
	u.normalize()
}

func (u *Score) IsBiggerOrEquals(other *Score) bool {
	if u.Exponent > other.Exponent {
		return true
	}
	if u.Exponent < other.Exponent {
		return false
	}
	return u.Mantissa >= other.Mantissa
}

func (u *Score) Multiply(factor float64) {
	discountedMantissa := float64(u.Mantissa) * factor
	u.Mantissa = int64(discountedMantissa)
	u.normalize()
}

func absInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

func (u *Score) normalize() {
	for u.Mantissa >= 1e18 {
		u.Mantissa /= 10
		u.Exponent++
	}
	for u.Mantissa < 1e17 && u.Exponent > 0 {
		u.Mantissa *= 10
		u.Exponent--
	}
	if u.Mantissa == 0 {
		u.Exponent = 0
	}
}
