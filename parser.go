package main

import (
	"regexp"
	"strconv"
	"strings"
)

type SearchQuery struct {
	OriginalText  string
	RelativeIndex string
	TimeIndex     string
	Mode          string
	Rule          string
}

// <command> := [前の|次の]+<type> | <type><time>
// <type> := <rule> | <mode>
// <mode> := ナワバリ[バトル]? | [ガチ|オープン|チャレンジ][マッチ]?
// <time> := 0, 1, ..., 24

func countRelativeIdentifier(input string) (result int) {
	next := strings.Count(input, IDENTIFIER_NEXT)
	prev := strings.Count(input, IDENTIFIER_PREV)
	return next - prev
}

func isRuleName(input string) bool {
	regex := regexp.MustCompile("(ガチ)?(エリア|ホコ|ホコバトル|ヤグラ|アサリ)")
	return regex.MatchString(input)
}

func searchModeIdentifier(input string) string {
	if strings.HasPrefix(input, "ガチマ") || strings.HasPrefix(input, "チャレンジ") {
		return "CHALLENGE"
	}
	if strings.HasPrefix(input, "リグマ") || strings.HasPrefix(input, "オープン") || strings.HasPrefix(input, "リーグ") {
		return "OPEN"
	}
	if strings.HasPrefix(input, "エックス") || strings.HasPrefix(input, "X") || strings.HasPrefix(input, "x") {
		return "X"
	}
	if strings.HasPrefix(input, "レギュラー") {
		return "REGULAR"
	}
	if strings.HasPrefix(input, "バカマ") || strings.HasPrefix(input, "バンカラ") {
		// show both CHALLENGE and OPEN
		return "BANKARA"
	}
	if strings.HasPrefix(input, "サーモン") || strings.HasPrefix(input, "シャケ") || strings.HasPrefix(input, "鮭") {
		return "SALMON"
	}
	if strings.HasPrefix(input, "ナワバリ") {
		// possible both regular or Splatfest
		return ""
	}
	if isRuleName(input) {
		return "BANKARA"
	}
	return ""
}

func searchRuleIdentifier(input string) string {
	if strings.HasPrefix(input, "エリア") {
		return "AREA"
	}
	if strings.HasPrefix(input, "ホコ") {
		return "GOAL"
	}
	if strings.HasPrefix(input, "ヤグラ") {
		return "LOFT"
	}
	if strings.HasPrefix(input, "アサリ") {
		return "CLAM"
	}
	if strings.HasPrefix(input, "ナワバリ") {
		return "TURF_WAR"
	}
	return ""
}

func Parse(input string) *SearchQuery {
	/*
	   次の次の前の次の次のガチマッチ
	   ガチマ
	   次のガチマ
	   次のオープンマッチ
	   ガチマアサリ
	   次のリグマヤグラ
	   次のナワバリバトル
	   エリア20
	   19 時のガチマッチ
	   ガチマ 20
	   次のエリア
	*/
	regex := regexp.MustCompile(`(((次の|前の)*)((\d{0,2}) ?時の)?((ガチマッチ|ガチマ|ガチ|リグマ|バカマ|(レギュラー|リーグ|バンカラ|オープン|チャレンジ|エックス|[Xx] ?)(マッチ)?)?(ガチ)?(ナワバリ|ナワバリバトル|エリア|ホコ|ホコバトル|ヤグラ|アサリ)?|シャケ|サーモン|サーモンラン|鮭) ?(\d{0,2}))$`)
	fss := regex.FindStringSubmatch(input)
	logger.Sugar().Infof("keyword input: %#v", fss)
	var timeIndex string
	if fss[5] != "" {
		timeIndex = fss[5]
	} else if fss[12] != "" {
		timeIndex = fss[12]
	} else {
		timeIndex = ""
	}
	var rindex string
	if fss[2] == "" {
		rindex = ""
	} else {
		rindex = strconv.Itoa(countRelativeIdentifier(fss[2]))
	}
	return &SearchQuery{
		OriginalText:  fss[0],
		RelativeIndex: rindex,
		TimeIndex:     timeIndex,
		Mode:          searchModeIdentifier(fss[6]),
		Rule:          searchRuleIdentifier(fss[11]),
	}
}
