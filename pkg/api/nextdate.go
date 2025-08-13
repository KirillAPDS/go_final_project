package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func afterNow(date, now time.Time) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := now.Date()
	return y1 > y2 || (y1 == y2 && m1 > m2) || (y1 == y2 && m1 == m2 && d1 > d2)
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("no repeat rule")
	}

	start, err := time.Parse(layout, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %w", err)
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", errors.New("invalid repeat rule")
	}

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("missing day interval")
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil || n < 1 || n > 400 {
			return "", errors.New("invalid day interval")
		}
		for {
			start = start.AddDate(0, 0, n)
			if afterNow(start, now) {
				return start.Format(layout), nil
			}
		}

	case "y":
		for {
			start = start.AddDate(1, 0, 0)
			if afterNow(start, now) {
				return start.Format(layout), nil
			}
		}

	case "w":
		if len(parts) < 2 {
			return "", errors.New("missing week days")
		}
		var days [8]bool
		for _, p := range strings.Split(parts[1], ",") {
			day, err := strconv.Atoi(p)
			if err != nil || day < 1 || day > 7 {
				return "", errors.New("invalid weekday")
			}
			days[day] = true
		}

		date := start
		for i := 0; ; i++ {
			date = date.AddDate(0, 0, 1)
			weekday := int(date.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			if days[weekday] && afterNow(date, now) {
				return date.Format(layout), nil
			}
		}

		//return "", errors.New("no matching weekday found")

	case "m":
		if len(parts) < 2 {
			return "", errors.New("missing month days")
		}
		dayParts := strings.Split(parts[1], ",")
		var monthParts []string
		if len(parts) == 3 {
			monthParts = strings.Split(parts[2], ",")
		}

		var dayMask [32]bool
		for _, p := range dayParts {
			day, err := strconv.Atoi(p)
			if err != nil || (day < -2 || day == 0 || day > 31) {
				return "", errors.New("invalid month day")
			}
			if day == -1 {
				dayMask[31] = true
			} else if day == -2 {
				dayMask[30] = true
			} else {
				dayMask[day] = true
			}
		}

		var monthMask [13]bool
		if len(monthParts) == 0 {
			for i := 1; i <= 12; i++ {
				monthMask[i] = true
			}
		} else {
			for _, p := range monthParts {
				m, err := strconv.Atoi(p)
				if err != nil || m < 1 || m > 12 {
					return "", errors.New("invalid month number")
				}
				monthMask[m] = true
			}
		}

		for i := 0; i < 366*5; i++ {
			date := start.AddDate(0, 0, i)
			d := date.Day()
			m := int(date.Month())
			if !monthMask[m] {
				continue
			}

			last := lastDayOfMonth(date)
			if (d <= 29 && dayMask[d]) ||
				(d == 31 && date.Day() == 31 && dayMask[31]) ||
				(d == last-1 && dayMask[30]) {
				if afterNow(date, now) {
					return date.Format(layout), nil
				}
			}
		}

		return "", errors.New("no matching month day found")

	default:
		return "", errors.New("unsupported rule")
	}
}

func lastDayOfMonth(t time.Time) int {
	t = t.AddDate(0, 1, -t.Day())
	return t.Day()
}
