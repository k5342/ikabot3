package main

import (
	"fmt"
	"sync"
	"time"
)

type ScheduleStore struct {
	sync.RWMutex
	info  *AllScheduleInfo
	cache *FileCache
}

func NewScheduleStore() ScheduleStore {
	return ScheduleStore{}
}

func (ss *ScheduleStore) MaybeRefresh() {
	ss.Lock()
	defer ss.Unlock()
	ss.cache = NewFileCache("./", "api_call_cache")
	cached := ss.cache.MaybeGet(time.Minute * 30)
	if cached == nil {
		// outdated. refresh schedule info
		logger.Sugar().Info("Cache is outdated. fetching...")
		info, err := FetchScheduleInfo()
		if err == nil {
			logger.Sugar().Info("fetch completed")
		} else {
			logger.Sugar().Errorw("Fetch failed", err)
		}
		ss.cache.Put(info)
		ss.info = info
	} else {
		logger.Sugar().Info("Cache is valid")
		ss.info = cached
	}
}

type SearchResult struct {
	Query      *SearchQuery
	Found      bool
	IsTwoSlots bool
	Slot1      *TimeSlotInfo
	Slot2      *TimeSlotInfo
}

func lookupByAbsoluteTime(tsinfos []TimeSlotInfo, hour int) (matched *TimeSlotInfo, found bool) {
	for _, tsinfo := range tsinfos {
		fmt.Printf("found => %d; req => %d\n", tsinfo.StartTime.Hour(), hour)
		if tsinfo.StartTime.Hour() == hour {
			return &tsinfo, true
		}
	}
	return nil, false
}

func lookupByRule(tsinfos []TimeSlotInfo, ruleKey string, skipCount int) (matched *TimeSlotInfo, found bool) {
	for _, tsinfo := range tsinfos {
		if tsinfo.Rule.Key == ruleKey {
			if skipCount <= 0 {
				return &tsinfo, true
			}
			skipCount -= 1
		}
	}
	return nil, false
}

func (ss *ScheduleStore) Search(query *SearchQuery) SearchResult {
	ss.RLock()
	defer ss.RUnlock()
	return search(query, ss.info, time.Now())
}

func search(query *SearchQuery, info *AllScheduleInfo, timeStamp time.Time) SearchResult {
	fmt.Printf("%#v\n", *query)
	var target []TimeSlotInfo
	if query.Mode == "REGULAR" {
		target = info.Regular
	}
	if query.Mode == "CHALLENGE" {
		target = info.BankaraChallenge
	}
	if query.Mode == "OPEN" {
		target = info.BankaraOpen
	}

	// search case #1: filter by rule
	if query.Rule != "" {
		matched, found := lookupByRule(target, query.Rule, query.RelativeIndex)
		return SearchResult{
			Query:      query,
			Found:      found,
			IsTwoSlots: false,
			Slot1:      matched,
		}
	}

	// search case #2, #3: lookup by time
	// assume Timestamp in API results is JST
	loc, _ := time.LoadLocation("Asia/Tokyo")
	absoluteStartTime := (int(timeStamp.In(loc).Hour()/2)*2 + 1) % 24
	if query.RelativeIndex != 0 {
		absoluteStartTime += query.RelativeIndex * 2
		absoluteStartTime %= 24
	}
	if query.TimeIndex != 0 {
		absoluteStartTime = (int(query.TimeIndex/2)*2 + 1) % 24
	}
	fmt.Println(absoluteStartTime)

	if query.Mode == "BANKARA" {
		// search case #2: lookup both by time
		matched1, found1 := lookupByAbsoluteTime(info.BankaraChallenge, absoluteStartTime)
		matched2, found2 := lookupByAbsoluteTime(info.BankaraOpen, absoluteStartTime)
		return SearchResult{
			Query:      query,
			Found:      found1 || found2,
			IsTwoSlots: true,
			Slot1:      matched1,
			Slot2:      matched2,
		}
	} else {
		// search case #3: lookup single by time
		matched, found := lookupByAbsoluteTime(target, absoluteStartTime)
		return SearchResult{
			Query:      query,
			Found:      found,
			IsTwoSlots: false,
			Slot1:      matched,
		}
	}
}
