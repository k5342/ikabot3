package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type SearchQuery struct {
	RelativeIndex      int
	RelativeIndexValid bool
	TimeIndex          int
	Mode               string
	Rule               string
}

// <command> := [前の|次の]+<type> | <type><time>
// <type> := <rule> | <mode>
// <mode> := ナワバリ[バトル]? | [ガチ|オープン|チャレンジ][マッチ]?
// <time> := 0, 1, ..., 24

// valid is to destinguish zero-value vs calcuration result (just for nit case)
func countRelativeIdentifier(input string) (result int, valid bool) {
	next := strings.Count(input, IDENTIFIER_NEXT)
	prev := strings.Count(input, IDENTIFIER_PREV)
	return next - prev, (next > 0 || prev > 0)
}

func searchModeIdentifier(input string) string {
	if strings.HasPrefix(input, "ガチマ") || strings.HasPrefix(input, "チャレンジ") {
		return "CHALLENGE"
	}
	if strings.HasPrefix(input, "リグマ") || strings.HasPrefix(input, "オープン") {
		return "OPEN"
	}
	if strings.HasPrefix(input, "ナワバリ") {
		return "REGULAR"
	}
	if strings.HasPrefix(input, "バンカラ") {
		// show both CHALLENGE and OPEN
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
	regex := regexp.MustCompile(`(((次の|前の)*)((\d{0,2}) ?時の)?(ナワバリ(バトル)?|(ガチマッチ|ガチマ|ガチ|リグマ|(リーグ|バンカラ|オープン|チャレンジ)(マッチ)?)?(ガチ)?(エリア|ホコ|ホコバトル|ヤグラ|アサリ)?) ?(\d{0,2}))`)
	fss := regex.FindStringSubmatch(input)
	fmt.Printf("%#v\n", fss)
	var timeIndex int
	if fss[5] != "" {
		timeIndex, _ = strconv.Atoi(fss[5])
	} else if fss[13] != "" {
		timeIndex, _ = strconv.Atoi(fss[13])
	} else {
		timeIndex = 0
	}
	rindex, rvalid := countRelativeIdentifier(fss[2])
	return &SearchQuery{
		RelativeIndex:      rindex,
		RelativeIndexValid: rvalid,
		TimeIndex:          timeIndex,
		Mode:               searchModeIdentifier(fss[6]),
		Rule:               searchRuleIdentifier(fss[12]),
	}
}
