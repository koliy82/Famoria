package bot

import (
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Me struct {
	ID        int64
	Username  string
	FirstName string
	LastName  string
	IsBot     bool
}

type PrintBotInfo interface {
	Print(log *zap.Logger)
}

func (m Me) Print(log *zap.Logger) {
	title := "Telegram Bot"
	author := "by Koliy82"
	centerItem := "          "

	horizontalWall := "│😈│"
	leftUpChar := "┏━━┯"
	rightUpChar := "┯━━┓"
	verticalWall := "━"
	leftDownChar := "┗━━┷"
	rightDownChar := "┷━━┛"
	spaceLength := 2

	labelsAndValues := []struct {
		label string
		value string
	}{
		{"[ID]", strconv.FormatInt(m.ID, 10)},
	}

	if m.Username != "" {
		labelsAndValues = append(labelsAndValues, struct {
			label string
			value string
		}{"[Username]", "@" + m.Username})
	}

	labelsAndValues = append(labelsAndValues, struct {
		label string
		value string
	}{"[FirstName]", m.FirstName})

	if m.LastName != "" {
		labelsAndValues = append(labelsAndValues, struct {
			label string
			value string
		}{"[Last Name]", m.LastName})
	}

	labelsAndValues = append(labelsAndValues, struct {
		label string
		value string
	}{"[Is Bot]", strconv.FormatBool(m.IsBot)})

	header, bodyLines, footer := buildBox(
		labelsAndValues,
		title,
		author,
		centerItem,
		horizontalWall,
		leftUpChar,
		rightUpChar,
		verticalWall,
		leftDownChar,
		rightDownChar,
		spaceLength,
	)

	log.Info(header)
	for _, line := range bodyLines {
		log.Info(line)
	}
	log.Info(footer)
}

func buildBox(
	labelsAndValues []struct {
		label string
		value string
	},
	title, author, centerItem, horizontalWall, leftUpChar, rightUpChar, verticalWall, leftDownChar, rightDownChar string,
	spaceLength int,
) (string, []string, string) {
	maxLabelLength := 0
	maxValueLength := 0

	for _, lv := range labelsAndValues {
		if utf8.RuneCountInString(lv.label) > maxLabelLength {
			maxLabelLength = utf8.RuneCountInString(lv.label)
		}
		if utf8.RuneCountInString(lv.value) > maxValueLength {
			maxValueLength = utf8.RuneCountInString(lv.value)
		}
	}

	var bodyLines []string

	for _, lv := range labelsAndValues {
		line := fmt.Sprintf("%s %-*s%s%*s %s",
			horizontalWall,
			maxLabelLength, lv.label,
			centerItem,
			maxValueLength, lv.value,
			horizontalWall,
		)
		bodyLines = append(bodyLines, line)
	}

	totalWidth := 0
	for _, line := range bodyLines {
		if utf8.RuneCountInString(line) > totalWidth {
			totalWidth = utf8.RuneCountInString(line)
		}
	}

	header := buildCenteredBox(title, totalWidth, verticalWall, spaceLength, leftUpChar, rightUpChar)
	footer := buildCenteredBox(author, totalWidth, verticalWall, spaceLength, leftDownChar, rightDownChar)

	return header, bodyLines, footer
}

func buildCenteredBox(
	text string,
	width int,
	verticalWall string,
	spaceLength int,
	leftChar, rightChar string,
) string {
	textLen := utf8.RuneCountInString(text)
	leftCharLen := utf8.RuneCountInString(leftChar)
	rightCharLen := utf8.RuneCountInString(rightChar)
	wallLen := utf8.RuneCountInString(verticalWall)
	if wallLen == 0 {
		wallLen = 1
	}

	dashes := (width - textLen - leftCharLen - rightCharLen - (spaceLength * 2)) / wallLen
	leftDash := dashes / 2
	rightDash := dashes - leftDash

	leftDashes := strings.Repeat(verticalWall, leftDash)
	rightDashes := strings.Repeat(verticalWall, rightDash)
	space := strings.Repeat(" ", spaceLength)

	return fmt.Sprintf("%s%s%s%s%s%s%s", leftChar, leftDashes, space, text, space, rightDashes, rightChar)
}
