package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

// https://spla3.yuu26.com/api/schedule
type AllAPIResult struct {
	Result AllScheduleInfo `json:"result"`
}

type AllScheduleInfo struct {
	Regular          []TimeSlotInfo `json:"regular"`
	BankaraChallenge []TimeSlotInfo `json:"bankara_challenge"`
	BankaraOpen      []TimeSlotInfo `json:"bankara_open"`
}

type TimeSlotInfo struct {
	StartTime time.Time   `json:"start_time"`
	EndTime   time.Time   `json:"end_time"`
	Rule      RuleInfo    `json:"rule"`
	Stages    []StageInfo `json:"stages"`
}

type RuleInfo struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type StageInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func fetchAll() (*AllAPIResult, error) {
	url := os.Getenv("IKABOT3_API_SOURCE")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", USER_AGENT)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ar AllAPIResult
	err = json.Unmarshal(bytes, &ar)
	if err != nil {
		return nil, err
	}

	return &ar, nil
}

func FetchScheduleInfo() (*AllScheduleInfo, error) {
	result, err := fetchAll()
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}
