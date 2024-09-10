package common

import (
	"fmt"
	"math"
	"strings"
)

type Score struct {
	Mantissa int64 `bson:"mantissa" default:"0"`
	Exponent int   `bson:"exponent" default:"0"`
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

	for discountedScore.Mantissa < 1e17 && discountedScore.Exponent > 0 {
		discountedScore.Mantissa *= 10
		discountedScore.Exponent--
	}

	if discountedScore.Mantissa < 0 {
		discountedScore.Mantissa = u.Mantissa
		discountedScore.Exponent = u.Exponent
	}

	return discountedScore
}

func (u *Score) GetFormattedScore() string {
	if u.Exponent < 3 {
		return formatSmallNumber(u.Mantissa * int64Pow(10, u.Exponent))
	}

	mantissaStr := fmt.Sprintf("%d", u.Mantissa)
	mantissaLength := len(mantissaStr)

	if mantissaLength > 1 {
		u.Exponent += mantissaLength - 1
	}

	mantissa := float64(u.Mantissa) / math.Pow(10, float64(mantissaLength-1))
	mantissaStr = fmt.Sprintf("%.2f", mantissa)
	mantissaStr = strings.TrimRight(mantissaStr, "0")
	mantissaStr = strings.TrimRight(mantissaStr, ".")

	return fmt.Sprintf("%se%d", mantissaStr, u.Exponent)
}

func formatSmallNumber(n int64) string {
	mod := math.Abs(float64(n))
	switch {
	case mod >= 1e15:
		return fmt.Sprintf("%.2fP", float64(n)/1e15) // Квадриллионы (P)
	case mod >= 1e12:
		return fmt.Sprintf("%.2fT", float64(n)/1e12) // Триллионы (T)
	case mod >= 1e9:
		return fmt.Sprintf("%.2fB", float64(n)/1e9) // Миллиарды (B)
	case mod >= 1e6:
		return fmt.Sprintf("%.2fM", float64(n)/1e6) // Миллионы (M)
	case mod >= 1e3:
		return fmt.Sprintf("%.2fK", float64(n)/1e3) // Тысячи (K)
	default:
		return fmt.Sprintf("%d", n) // Менее тысячи
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
	for increment >= 1e18 {
		u.Mantissa += int64(increment / 1e18)
		u.Exponent++
		if u.Mantissa >= 1e18 {
			u.Mantissa /= 10
			u.Exponent++
		}
		increment %= 1e18
	}

	u.Mantissa += int64(increment)

	if u.Mantissa >= 1e18 {
		u.Mantissa /= 10
		u.Exponent++
	}

	if u.Mantissa < 0 {
		u.Mantissa = 0
	}
}

func (u *Score) Decrease(decrement uint64) {
	for decrement >= 1e18 {
		u.Mantissa -= int64(decrement / 1e18)
		u.Exponent--
		if u.Mantissa < 0 {
			u.Mantissa *= 10
			u.Exponent--
		}
		decrement %= 1e18
	}

	u.Mantissa -= int64(decrement)

	if u.Mantissa < 0 {
		u.Mantissa = 0
	}
	if u.Exponent < 0 {
		u.Exponent = 0
	}
}

func (u *Score) Minus(other *Score) {
	// Приводим мантиссу и экспоненту к одному уровню для вычитания
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
		// Если экспоненты равны, вычитаем мантиссы
		u.Mantissa -= other.Mantissa
	}

	// Нормализация результата
	for u.Mantissa < 0 && u.Exponent > 0 {
		u.Mantissa *= 10
		u.Exponent--
	}

	if u.Mantissa < 0 {
		u.Mantissa = 0
		u.Exponent = 0
	}
}

func (u *Score) Plus(other *Score) {
	// Приводим мантиссу и экспоненту к одному уровню для сложения
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
		// Если экспоненты равны, складываем мантиссы
		u.Mantissa += other.Mantissa
	}

	// Нормализация результата
	for u.Mantissa >= 1e18 {
		u.Mantissa /= 10
		u.Exponent++
	}
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
