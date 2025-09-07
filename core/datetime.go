package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	Text "github.com/linkdotnet/golang-stringbuilder"
	"github.com/sderkacs/exi-go/utils"
)

func DateTimeParse(cal string, kind DateTimeType) (*DateTimeValue, error) {
	cal = strings.TrimSpace(cal)

	sYear := 0
	sMonthDay := 0
	sTime := 0
	sFractionalSecs := 0
	var sPresenceTimezone bool
	var sTimezone int
	var err error

	sb := Text.StringBuilder{}
	sb.Append(cal)

	switch kind {
	// gYear Year, [Time-Zone]
	case DateTimeGYear:
		sYear, err = dateTimeParseYear(&sb)
		if err != nil {
			return nil, err
		}
	case DateTimeGYearMonth:
		sYear, err = dateTimeParseYear(&sb)
		if err != nil {
			return nil, err
		}
		if err := dateTimeCheckCharacter(&sb, '-'); err != nil {
			return nil, err
		}
		sMonthDay, err = dateTimeParseMonth(&sb)
		if err != nil {
			return nil, err
		}
		sMonthDay *= DateTimeValue_MonthMultiplicator
	case DateTimeDate:
		sYear, err = dateTimeParseYear(&sb)
		if err != nil {
			return nil, err
		}
		if err := dateTimeCheckCharacter(&sb, '-'); err != nil {
			return nil, err
		}
		sMonthDay, err = dateTimeParseMonthDay(&sb)
		if err != nil {
			return nil, err
		}
		if err := dateTimeCheckCharacter(&sb, 'T'); err != nil {
			return nil, err
		}
		// No break!
		fallthrough
	case DateTimeTime:
		sTime, err = dateTimeParseTime(&sb)
		if err != nil {
			return nil, err
		}
		if sb.Len() > 0 && sb.RuneAt(0) == '.' {
			if err := sb.Remove(0, 1); err != nil {
				return nil, err
			}
			digits := dateTimeCountDigits(&sb)
			tmp, err := sb.Substring(0, digits)
			if err != nil {
				return nil, err
			}
			sb2 := Text.StringBuilder{}
			fracSec, err := strconv.ParseInt(sb2.Append(tmp).Reverse().ToString(), 10, 32)
			if err != nil {
				return nil, err
			}
			sFractionalSecs = int(fracSec)

			if err := sb.Remove(0, digits); err != nil {
				return nil, err
			}
		}
	case DateTimeGMonth:
		if err := dateTimeCheckCharacter(&sb, '-'); err != nil {
			return nil, err
		}
		if err := dateTimeCheckCharacter(&sb, '-'); err != nil {
			return nil, err
		}
		sMonthDay, err = dateTimeParseMonth(&sb)
		if err != nil {
			return nil, err
		}
		sMonthDay *= DateTimeValue_MonthMultiplicator

		if sb.Len() > 1 && sb.RuneAt(0) == sb.RuneAt(1) && sb.RuneAt(0) == '-' {
			if err := dateTimeCheckCharacter(&sb, '-'); err != nil {
				return nil, err
			}
			if err := dateTimeCheckCharacter(&sb, '-'); err != nil {
				return nil, err
			}
		}
	case DateTimeGMonthDay:
		if err := dateTimeCheckCharacter(&sb, '-'); err != nil {
			return nil, err
		}
		if err := dateTimeCheckCharacter(&sb, '-'); err != nil {
			return nil, err
		}
		sMonthDay, err = dateTimeParseMonth(&sb)
		if err != nil {
			return nil, err
		}
	case DateTimeGDay:
		if err := dateTimeCheckCharacter(&sb, '-'); err != nil {
			return nil, err
		}
		if err := dateTimeCheckCharacter(&sb, '-'); err != nil {
			return nil, err
		}
		if err := dateTimeCheckCharacter(&sb, '-'); err != nil {
			return nil, err
		}
		sMonthDay, err = dateTimeParseDay(&sb)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported date time type: %d", kind)
	}

	// [TimeZone]
	// lexical representation of a timezone: (('+' | '-') hh ':' mm) |
	// 'Z',
	// where
	// * hh is a two-digit numeral (with leading zeros as required) that
	// represents the hours,
	// * mm is a two-digit numeral that represents the minutes,
	// * '+' indicates a nonnegative duration,
	// * '-' indicates a nonpositive duration.
	//
	// TimeZone TZHours * 64 + TZMinutes (896 = 14 * 64)

	// plus, minus, Z or nothing ?
	if sb.Len() == 0 {
		sPresenceTimezone = false
		sTimezone = 0
	} else if sb.Len() == 1 && sb.RuneAt(0) == 'Z' {
		if err := sb.Remove(0, 1); err != nil {
			return nil, err
		}
		sPresenceTimezone = true
		sTimezone = 0
	} else {
		sPresenceTimezone = true
		var multiplicator int

		if sb.RuneAt(0) == '+' {
			multiplicator = 1
		} else if sb.RuneAt(0) == '-' {
			multiplicator = -1
		} else {
			return nil, fmt.Errorf("unexpected character while parsing: %c", sb.RuneAt(0))
		}

		tmp, err := sb.Substring(1, 3)
		if err != nil {
			return nil, err
		}
		hours, err := strconv.ParseInt(tmp, 10, 32)
		if err != nil {
			return nil, err
		}

		tmp, err = sb.Substring(4, 6)
		if err != nil {
			return nil, err
		}
		minutes, err := strconv.ParseInt(tmp, 10, 32)
		if err != nil {
			return nil, err
		}

		sTimezone = multiplicator * (int(hours)*DateTimeValue_SecondsInMinute + int(minutes))
	}

	return NewDateTimeValue(kind, sYear, sMonthDay, sTime, sFractionalSecs, sPresenceTimezone, sTimezone), nil
}

func dateTimeParseYear(sb *Text.StringBuilder) (int, error) {
	var sYear string
	var len int
	var err error

	if sb.RuneAt(0) == '-' {
		sYear, err = sb.Substring(0, 5)
		if err != nil {
			return -1, err
		}
		len = 5
	} else {
		sYear, err = sb.Substring(0, 4)
		if err != nil {
			return -1, err
		}
		len = 4
	}
	year, err := strconv.ParseInt(sYear, 10, 32)
	if err != nil {
		return -1, err
	}

	if err := sb.Remove(0, len); err != nil {
		return -1, err
	}

	return int(year), nil
}

func dateTimeParseMonth(sb *Text.StringBuilder) (int, error) {
	sMonth, err := sb.Substring(0, 2)
	if err != nil {
		return -1, err
	}
	month, err := strconv.ParseInt(sMonth, 10, 32)
	if err != nil {
		return -1, err
	}

	if err := sb.Remove(0, 2); err != nil {
		return -1, err
	}

	return int(month), nil
}

func dateTimeParseDay(sb *Text.StringBuilder) (int, error) {
	sDay, err := sb.Substring(0, 2)
	if err != nil {
		return -1, err
	}
	day, err := strconv.ParseInt(sDay, 10, 32)
	if err != nil {
		return -1, err
	}

	if err := sb.Remove(0, 2); err != nil {
		return -1, err
	}

	return int(day), nil
}

func dateTimeCheckCharacter(sb *Text.StringBuilder, c rune) error {
	if sb.Len() > 0 && sb.RuneAt(0) == c {
		if err := sb.Remove(0, 1); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unexpected character while parsing")
	}

	return nil
}

func dateTimeParseMonthDay(sb *Text.StringBuilder) (int, error) {
	month, err := dateTimeParseMonth(sb)
	if err != nil {
		return -1, err
	}
	if err := dateTimeCheckCharacter(sb, '-'); err != nil {
		return -1, err
	}
	day, err := dateTimeParseDay(sb)
	if err != nil {
		return -1, err
	}

	return int(month)*DateTimeValue_MonthMultiplicator + int(day), nil
}

// Time ((Hour * 64) + Minutes) * 64 + seconds
func dateTimeParseTime(sb *Text.StringBuilder) (int, error) {
	// Hour
	sHour, err := sb.Substring(0, 2)
	if err != nil {
		return -1, err
	}
	hour, err := strconv.ParseInt(sHour, 10, 32)
	if err != nil {
		return -1, err
	}
	if err := sb.Remove(0, 2); err != nil {
		return -1, err
	}

	// Colon
	if err := dateTimeCheckCharacter(sb, ':'); err != nil {
		return -1, err
	}

	// Minute
	sMinutes, err := sb.Substring(0, 2)
	if err != nil {
		return -1, err
	}
	minutes, err := strconv.ParseInt(sMinutes, 10, 32)
	if err != nil {
		return -1, err
	}
	if err := sb.Remove(0, 2); err != nil {
		return -1, err
	}

	// Colon
	if err := dateTimeCheckCharacter(sb, ':'); err != nil {
		return -1, err
	}

	// Second
	sSeconds, err := sb.Substring(0, 2)
	if err != nil {
		return -1, err
	}
	seconds, err := strconv.ParseInt(sSeconds, 10, 32)
	if err != nil {
		return -1, err
	}
	if err := sb.Remove(0, 2); err != nil {
		return -1, err
	}

	return ((int(hour)*DateTimeValue_SecondsInMinute)+int(minutes))*DateTimeValue_SecondsInMinute + int(seconds), nil
}

func dateTimeCountDigits(sb *Text.StringBuilder) int {
	len := sb.Len()
	idx := 0

	for idx < len && unicode.IsDigit(sb.RuneAt(idx)) {
		idx++
	}

	return idx
}

/*
 * Encode Date-Time as a sequence of values representing the individual
 * components of the Date-Time.
 */
func DateTimeParseTime(time *time.Time, kind DateTimeType) (*DateTimeValue, error) {
	sYear := 0
	sMonthDay := 0
	sTime := 0
	sFractionalSecs := 0
	sPresenceTimezone := false
	sTimezone := 0

	switch kind {
	case DateTimeGYear, DateTimeGYearMonth, DateTimeDate:
		sYear = time.Year()
		sMonthDay = dateTimeGetMonthDay(time)
	case DateTimeDateTime:
		sYear = time.Year()
		sMonthDay = dateTimeGetMonthDay(time)
	case DateTimeTime:
		sTime = dateTimeGetTime(time)
		sFractionalSecs = time.Nanosecond() * 1_000_000
	case DateTimeGMonth, DateTimeGMonthDay, DateTimeGDay:
		sMonthDay = dateTimeGetMonthDay(time)
	default:
		return nil, fmt.Errorf("unsupported date time type: %d", kind)
	}

	sTimezone = dateTimeGetTimeZoneInMinutesOffset(time)
	if sTimezone != 0 {
		sPresenceTimezone = true
	}

	return NewDateTimeValue(kind, sYear, sMonthDay, sTime, sFractionalSecs, sPresenceTimezone, sTimezone), nil
}

func dateTimeGetMonthDay(time *time.Time) int {
	month := time.Month() + 1
	day := time.Day()

	return int(month)*DateTimeValue_MonthMultiplicator + int(day)
}

func dateTimeGetTime(time *time.Time) int {
	t := time.Hour()
	t *= DateTimeValue_SecondsInMinute
	t += time.Minute()
	t *= DateTimeValue_SecondsInMinute
	t += time.Second()

	return t
}

func dateTimeGetTimeZoneInMinutesOffset(time *time.Time) int {
	_, offset := time.Zone()
	return offset/(1000*60) + DateTimeValue_TimeZoneOffsetInMinutes
}

func dateTimeGetTimeZoneInMillisecs(minutes int) int {
	return minutes / (1000 * 60)
}

func dateTimeSetMonthDay(monthDay int, t time.Time) time.Time {
	month := monthDay / DateTimeValue_MonthMultiplicator
	day := monthDay - month*DateTimeValue_MonthMultiplicator

	return time.Date(t.Year(), time.Month(month), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), nil)
}

func dateTimeSetTime(timeValue int, t time.Time) time.Time {
	// ((Hour * 64) + Minutes) * 64 + seconds
	hour := timeValue / DateTimeValue_SecondsInHour
	timeValue -= hour * DateTimeValue_SecondsInHour
	minute := timeValue / DateTimeValue_SecondsInMinute
	timeValue -= minute * DateTimeValue_SecondsInMinute

	return time.Date(t.Year(), t.Month(), t.Day(), hour, minute, timeValue, t.Nanosecond(), nil)
}

func dateTimeSetTimezone(tz int, t time.Time) time.Time {
	loc := time.FixedZone("GMT", tz)
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

func dateTimeAppendYear(ca []rune, index *int, year int) {
	if year < 0 {
		ca[*index] = '-'
		*index++
		year = -year
	}

	if year > 999 {
		*index += 4
	} else if year > 99 {
		ca[*index] = '0'
		*index += 4
	} else if year > 9 {
		ca[*index] = '0'
		*index++
		ca[*index] = '0'
		*index += 3
	} else {
		ca[*index] = '0'
		*index++
		ca[*index] = '0'
		*index++
		ca[*index] = '0'
		*index += 2
	}

	utils.Itos32(year, index, ca)
}

func dateTimeAppendTwoDigits(ca []rune, index *int, i int) {
	if i > 9 {
		*index += 2
	} else {
		ca[*index] = '0'
		*index += 2
	}
	utils.Itos32(i, index, ca)
}

func dateTimeAppendMonth(ca []rune, index *int, monthDay int) {
	month := monthDay / DateTimeValue_MonthMultiplicator

	*index--
	ca[*index] = '-'

	dateTimeAppendTwoDigits(ca, index, month)
}

func dateTimeAppendMonthDay(ca []rune, index *int, monthDay int) {
	// monthDay: Month * 32 + Day

	// month & day
	month := monthDay / DateTimeValue_MonthMultiplicator
	day := monthDay - (month * DateTimeValue_MonthMultiplicator)

	// -MM-DD
	ca[*index] = '-'
	*index++
	dateTimeAppendTwoDigits(ca, index, month)
	ca[*index] = '-'
	*index++
	dateTimeAppendTwoDigits(ca, index, day)
}

func dateTimeAppendDay(ca []rune, index *int, day int) {
	dateTimeAppendTwoDigits(ca, index, day)
}

func dateTimeAppendTime(ca []rune, index *int, time int) {
	// time = ( ( hour * 64) + minutes ) * 64 + seconds

	hour := time / DateTimeValue_SecondsInHour
	time -= hour * DateTimeValue_SecondsInHour
	minutes := time / DateTimeValue_SecondsInMinute
	seconds := time - minutes*DateTimeValue_SecondsInMinute

	dateTimeAppendTwoDigits(ca, index, hour)
	ca[*index] = ':'
	*index++
	dateTimeAppendTwoDigits(ca, index, minutes)
	ca[*index] = ':'
	*index++
	dateTimeAppendTwoDigits(ca, index, seconds)
}

func dateTimeAppendFractionalSeconds(ca []rune, index *int, fracSecs, sLen int) {
	if fracSecs > 0 {
		ca[*index] = '.'
		*index++

		i := *index
		utils.ItosReverse32(fracSecs, &i, ca)
		*index += i
	}
}

func dateTimeAppendTimezone(ca []rune, index *int, tz int) {
	if tz == 0 {
		// per default 'Z'
		ca[*index] = 'Z'
		*index++
	} else {
		// +/-
		if tz < 0 {
			ca[*index] = '-'
			*index++
			tz *= -1
		} else {
			ca[*index] = '+'
			*index++
		}

		hours := tz / 64
		dateTimeAppendTwoDigits(ca, index, hours)

		ca[*index] = ':'
		*index++

		minutes := tz - (hours * 64)
		dateTimeAppendTwoDigits(ca, index, minutes)
	}
}
