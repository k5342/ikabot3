package main

import (
	"reflect"
	"testing"

	"go.uber.org/zap"
)

func init() {
	logger, _ = zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()
}

func Test_countRelativeIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		args       string
		wantResult int
	}{
		{
			name:       "次の must be proceed as 1",
			args:       "次の",
			wantResult: 1,
		},
		{
			name:       "前の must be proceed as -1",
			args:       "前の",
			wantResult: -1,
		},
		{
			name:       "次の次の must be proceed as 2",
			args:       "次の次の",
			wantResult: 2,
		},
		{
			name:       "次の前の次の must be proceed as 1",
			args:       "次の前の次の",
			wantResult: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := countRelativeIdentifier(tt.args); gotResult != tt.wantResult {
				t.Errorf("countRelativeIdentifier() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_isRuleName(t *testing.T) {
	tests := []struct {
		name string
		args string
		want bool
	}{
		{
			name: "ガチエリア must be valid",
			args: "ガチエリア",
			want: true,
		},
		{
			name: "ガチホコ must be valid",
			args: "ガチホコ",
			want: true,
		},
		{
			name: "ホコ must be valid",
			args: "ホコ",
			want: true,
		},
		{
			name: "ガチホコバトル must be valid",
			args: "ガチホコバトル",
			want: true,
		},
		{
			name: "ガチヤグラ must be valid",
			args: "ガチヤグラ",
			want: true,
		},
		{
			name: "ヤグラ must be valid",
			args: "ヤグラ",
			want: true,
		},
		{
			name: "ガチ must not be valid",
			args: "ガチ",
			want: false,
		},
		{
			name: "バトル must not be valid",
			args: "バトル",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRuleName(tt.args); got != tt.want {
				t.Errorf("isRuleName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_searchModeIdentifier(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "ガチマ must be proceed as CHALLENGE",
			args: "ガチマ",
			want: "CHALLENGE",
		},
		{
			name: "ガチマッチ must be proceed as CHALLENGE",
			args: "ガチマ",
			want: "CHALLENGE",
		},
		{
			name: "チャレンジ must be proceed as CHALLENGE",
			args: "ガチマ",
			want: "CHALLENGE",
		},
		{
			name: "チャレンジマッチ must be proceed as CHALLENGE",
			args: "ガチマ",
			want: "CHALLENGE",
		},
		{
			name: "リグマ must be proceed as OPEN",
			args: "リグマ",
			want: "OPEN",
		},
		{
			name: "リーグマッチ must be proceed as OPEN",
			args: "リーグマッチ",
			want: "OPEN",
		},
		{
			name: "リーグ must be proceed as OPEN",
			args: "リーグ",
			want: "OPEN",
		},
		{
			name: "オープン must be proceed as OPEN",
			args: "オープン",
			want: "OPEN",
		},
		{
			name: "オープンマッチ must be proceed as OPEN",
			args: "オープンマッチ",
			want: "OPEN",
		},
		{
			name: "ナワバリ must be proceed as ''",
			args: "ナワバリ",
			want: "",
		},
		{
			name: "ナワバリバトル must be proceed as ''",
			args: "ナワバリバトル",
			want: "",
		},
		{
			name: "レギュラーマッチ must be proceed as REGULAR",
			args: "レギュラーマッチ",
			want: "REGULAR",
		},
		{
			name: "レギュラー must be proceed as REGULAR",
			args: "レギュラー",
			want: "REGULAR",
		},
		{
			name: "バンカラマッチ must be proceed as BANKARA",
			args: "バンカラマッチ",
			want: "BANKARA",
		},
		{
			name: "バカマ must be proceed as BANKARA",
			args: "バカマ",
			want: "BANKARA",
		},
		{
			name: "バンカラ must be proceed as BANKARA",
			args: "バンカラ",
			want: "BANKARA",
		},
		{
			name: "サーモンラン must be proceed as SALMON",
			args: "サーモンラン",
			want: "SALMON",
		},
		{
			name: "サーモン must be proceed as SALMON",
			args: "サーモン",
			want: "SALMON",
		},
		{
			name: "シャケ must be proceed as SALMON",
			args: "シャケ",
			want: "SALMON",
		},
		{
			name: "鮭 must be proceed as SALMON",
			args: "鮭",
			want: "SALMON",
		},
		{
			name: "エックスマッチ must be proceed as X",
			args: "エックスマッチ",
			want: "X",
		},
		{
			name: "エックス must be proceed as X",
			args: "エックス",
			want: "X",
		},
		{
			name: "Xマッチ must be proceed as X",
			args: "Xマッチ",
			want: "X",
		},
		{
			name: "xマッチ must be proceed as X",
			args: "xマッチ",
			want: "X",
		},
		{
			name: "X マッチ must be proceed as X",
			args: "X マッチ",
			want: "X",
		},
		{
			name: "x マッチ must be proceed as X",
			args: "x マッチ",
			want: "X",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := searchModeIdentifier(tt.args); got != tt.want {
				t.Errorf("searchModeIdentifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_searchRuleIdentifier(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "エリア must be proceed as AREA",
			args: "エリア",
			want: "AREA",
		},
		{
			name: "ホコ must be proceed as GOAL",
			args: "ホコ",
			want: "GOAL",
		},
		{
			name: "ホコバトル must be proceed as GOAL",
			args: "ホコバトル",
			want: "GOAL",
		},
		{
			name: "ヤグラ must be proceed as LOFT",
			args: "ヤグラ",
			want: "LOFT",
		},
		{
			name: "アサリ must be proceed as CLAM",
			args: "アサリ",
			want: "CLAM",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := searchRuleIdentifier(tt.args); got != tt.want {
				t.Errorf("searchRuleIdentifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		args string
		want *SearchQuery
	}{
		{
			name: "ガチマ",
			args: "ガチマ",
			want: &SearchQuery{
				OriginalText:  "ガチマ",
				RelativeIndex: "",
				TimeIndex:     "",
				Mode:          getMode("CHALLENGE"),
				Rule:          "",
			},
		},
		{
			name: "次の次の前の次の次のガチマッチ",
			args: "次の次の前の次の次のガチマッチ",
			want: &SearchQuery{
				OriginalText:  "次の次の前の次の次のガチマッチ",
				RelativeIndex: "3",
				TimeIndex:     "",
				Mode:          getMode("CHALLENGE"),
				Rule:          "",
			},
		},
		{
			name: "次のガチマ",
			args: "次のガチマ",
			want: &SearchQuery{
				OriginalText:  "次のガチマ",
				RelativeIndex: "1",
				TimeIndex:     "",
				Mode:          getMode("CHALLENGE"),
				Rule:          "",
			},
		},
		{
			name: "次のオープンマッチ",
			args: "次のオープンマッチ",
			want: &SearchQuery{
				OriginalText:  "次のオープンマッチ",
				RelativeIndex: "1",
				TimeIndex:     "",
				Mode:          getMode("OPEN"),
				Rule:          "",
			},
		},
		{
			name: "ガチマアサリ",
			args: "ガチマアサリ",
			want: &SearchQuery{
				OriginalText:  "ガチマアサリ",
				RelativeIndex: "",
				TimeIndex:     "",
				Mode:          getMode("CHALLENGE"),
				Rule:          "CLAM",
			},
		},
		{
			name: "次のリグマヤグラ",
			args: "次のリグマヤグラ",
			want: &SearchQuery{
				OriginalText:  "次のリグマヤグラ",
				RelativeIndex: "1",
				TimeIndex:     "",
				Mode:          getMode("OPEN"),
				Rule:          "LOFT",
			},
		},
		{
			name: "次のナワバリバトル",
			args: "次のナワバリバトル",
			want: &SearchQuery{
				OriginalText:  "次のナワバリバトル",
				RelativeIndex: "1",
				TimeIndex:     "",
				Mode:          getMode(""),
				Rule:          "TURF_WAR",
			},
		},
		{
			name: "エリア20",
			args: "エリア20",
			want: &SearchQuery{
				OriginalText:  "エリア20",
				RelativeIndex: "",
				TimeIndex:     "20", // XXX: parser returns both info even if conflict search mode with rule and timeIndex
				Mode:          getMode("BANKARA"),
				Rule:          "AREA",
			},
		},
		{
			name: "19 時のガチマッチ",
			args: "19 時のガチマッチ",
			want: &SearchQuery{
				OriginalText:  "19 時のガチマッチ",
				RelativeIndex: "",
				TimeIndex:     "19",
				Mode:          getMode("CHALLENGE"),
				Rule:          "",
			},
		},
		{
			name: "ガチマ 20",
			args: "ガチマ 20",
			want: &SearchQuery{
				OriginalText:  "ガチマ 20",
				RelativeIndex: "",
				TimeIndex:     "20",
				Mode:          getMode("CHALLENGE"),
				Rule:          "",
			},
		},
		{
			name: "次のエリア",
			args: "次のエリア",
			want: &SearchQuery{
				OriginalText:  "次のエリア",
				RelativeIndex: "1",
				TimeIndex:     "",
				Mode:          getMode("BANKARA"),
				Rule:          "AREA",
			},
		},
		{
			name: "次のガチヤグラ",
			args: "次のガチヤグラ",
			want: &SearchQuery{
				OriginalText:  "次のガチヤグラ",
				RelativeIndex: "1",
				TimeIndex:     "",
				Mode:          getMode("BANKARA"), // not ガチマヤグラ
				Rule:          "LOFT",
			},
		},
		{
			name: "次のガチマヤグラ",
			args: "次のガチマヤグラ",
			want: &SearchQuery{
				OriginalText:  "次のガチマヤグラ",
				RelativeIndex: "1",
				TimeIndex:     "",
				Mode:          getMode("CHALLENGE"),
				Rule:          "LOFT",
			},
		},
		{
			name: "シャケ",
			args: "シャケ",
			want: &SearchQuery{
				OriginalText:  "シャケ",
				RelativeIndex: "",
				TimeIndex:     "",
				Mode:          getMode("SALMON"),
				Rule:          "",
			},
		},
		{
			name: "次のサーモンラン",
			args: "次のサーモンラン",
			want: &SearchQuery{
				OriginalText:  "次のサーモンラン",
				RelativeIndex: "1",
				TimeIndex:     "",
				Mode:          getMode("SALMON"),
				Rule:          "",
			},
		},
		{
			name: "ナワバリバトル",
			args: "ナワバリバトル",
			want: &SearchQuery{
				OriginalText:  "ナワバリバトル",
				RelativeIndex: "",
				TimeIndex:     "",
				Mode:          getMode(""),
				Rule:          "TURF_WAR",
			},
		},
		{
			name: "次のレギュラーマッチ",
			args: "次のレギュラーマッチ",
			want: &SearchQuery{
				OriginalText:  "次のレギュラーマッチ",
				RelativeIndex: "1",
				TimeIndex:     "",
				Mode:          getMode("REGULAR"),
				Rule:          "",
			},
		},
		{
			name: "次のエックスマッチ",
			args: "次のエックスマッチ",
			want: &SearchQuery{
				OriginalText:  "次のエックスマッチ",
				RelativeIndex: "1",
				TimeIndex:     "",
				Mode:          getMode("X"),
				Rule:          "",
			},
		},
		{
			name: "次の次のエックスマッチガチホコバトル",
			args: "次の次のエックスマッチガチホコバトル",
			want: &SearchQuery{
				OriginalText:  "次の次のエックスマッチガチホコバトル",
				RelativeIndex: "2",
				TimeIndex:     "",
				Mode:          getMode("X"),
				Rule:          "GOAL",
			},
		},
		{
			name: "Xマッチアサリ",
			args: "Xマッチアサリ",
			want: &SearchQuery{
				OriginalText:  "Xマッチアサリ",
				RelativeIndex: "",
				TimeIndex:     "",
				Mode:          getMode("X"),
				Rule:          "CLAM",
			},
		},
		{
			name: "x マッチガチエリア",
			args: "x マッチガチエリア",
			want: &SearchQuery{
				OriginalText:  "x マッチガチエリア",
				RelativeIndex: "",
				TimeIndex:     "",
				Mode:          getMode("X"),
				Rule:          "AREA",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
