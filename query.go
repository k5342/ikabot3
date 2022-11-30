package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
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
	XMatch           []TimeSlotInfo `json:"x"`
}

type TimeSlotInfo struct {
	StartTime time.Time   `json:"start_time"`
	EndTime   time.Time   `json:"end_time"`
	Rule      RuleInfo    `json:"rule"`
	Stages    []StageInfo `json:"stages"`
	IsFest    bool        `json:"is_fest"`
	// Salmon Run
	Stage   StageInfo    `json:"stage"`
	Weapons []WeaponInfo `json:"weapons"`
}

type RuleInfo struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type StageInfo struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image,omitempty"`
}

type SalmonAPIResult struct {
	Results []TimeSlotInfo `json:"results"`
}

type WeaponInfo struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

func query(url string) ([]byte, error) {
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

	return io.ReadAll(resp.Body)
}

func getSource() string {
	return os.Getenv("IKABOT3_API_SOURCE")
}

func fetchAll() (*AllAPIResult, error) {
	bytes, err := query(getSource())
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

func fetchSalmon() (*SalmonAPIResult, error) {
	url, err := url.JoinPath(getSource(), "..", "coop-grouping-regular/schedule")
	if err != nil {
		return nil, err
	}
	bytes, err := query(url)
	if err != nil {
		return nil, err
	}

	var ar SalmonAPIResult
	err = json.Unmarshal(bytes, &ar)
	if err != nil {
		return nil, err
	}

	return &ar, nil
}

func FetchScheduleInfoSalmon() (*[]TimeSlotInfo, error) {
	result, err := fetchSalmon()
	if err != nil {
		return nil, err
	}
	return &result.Results, nil
}
