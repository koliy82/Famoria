package common

import (
	"fmt"
)

func FormattedScore(score int64) string {
	if score < 0 {
		return fmt.Sprintf("%d", score)
	}
	switch {
	case score >= 1e15:
		return fmt.Sprintf("%.2fP", float64(score)/1e15) // Квадриллионы (P)
	case score >= 1e12:
		return fmt.Sprintf("%.2fT", float64(score)/1e12) // Триллионы (T)
	case score >= 1e9:
		return fmt.Sprintf("%.2fB", float64(score)/1e9) // Миллиарды (B)
	case score >= 1e6:
		return fmt.Sprintf("%.2fM", float64(score)/1e6) // Миллионы (M)
	case score >= 1e3:
		return fmt.Sprintf("%.2fK", float64(score)/1e3) // Тысячи (K)
	default:
		return fmt.Sprintf("%d", score) // Менее тысячи
	}
}

// GetSaleScore применяет скидку sale (0 < sale < 1) к целочисленной цене score и
// возвращает указатель на новое значение. Возвращает nil, если скидка не применима.
func GetSaleScore(score int64, sale float64) *int64 {
	if sale <= 0 || sale >= 1 {
		return nil
	}
	v := int64(float64(score) * (1 - sale))
	return &v
}
