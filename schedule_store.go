package main

import (
	"strconv"
	"sync"
	"time"
)

type ScheduleStore struct {
	sync.RWMutex
	info  *AllScheduleInfo
	cache *FileCache
}

func NewScheduleStore() ScheduleStore {
	return ScheduleStore{
		cache: NewFileCache("./", "api_call_cache"),
	}
}

func (ss *ScheduleStore) MaybeRefresh() {
	ss.Lock()
	defer ss.Unlock()
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
		logger.Sugar().Infof("found => %d; req => %d\n", tsinfo.StartTime.Hour(), hour)
		if tsinfo.StartTime.Hour() == hour && !tsinfo.IsFest {
			return &tsinfo, true
		}
	}
	return nil, false
}

func lookupByRule(tsinfos []TimeSlotInfo, ruleKey string, skipCount int) (matched *TimeSlotInfo, found bool) {
	for _, tsinfo := range tsinfos {
		if tsinfo.Rule.Key == ruleKey && !tsinfo.IsFest {
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
	logger.Sugar().Infof("search request: %#v", *query)
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
		var skipCount int
		sc, err := strconv.Atoi(query.RelativeIndex)
		if query.RelativeIndex != "" && err == nil {
			// valid
			skipCount = sc
		} else {
			skipCount = 0
		}

		if query.Mode == "BANKARA" {
			matched1, found1 := lookupByRule(info.BankaraChallenge, query.Rule, skipCount)
			matched2, found2 := lookupByRule(info.BankaraOpen, query.Rule, skipCount)
			return SearchResult{
				Query:      query,
				Found:      found1 || found2,
				IsTwoSlots: true,
				Slot1:      matched1,
				Slot2:      matched2,
			}
		} else {
			matched, found := lookupByRule(target, query.Rule, skipCount)
			return SearchResult{
				Query:      query,
				Found:      found,
				IsTwoSlots: false,
				Slot1:      matched,
			}
		}
	}

	// search case #2, #3: lookup by time
	// assume Timestamp in API results is JST
	loc, _ := time.LoadLocation("Asia/Tokyo")
	absoluteStartTime := (timeStamp.In(loc).Hour() - ((timeStamp.In(loc).Hour() + 1) % 2)) % 24
	if absoluteStartTime < 0 {
		absoluteStartTime = 23
	}
	if query.RelativeIndex != "" {
		relativeIdx, err := strconv.Atoi(query.RelativeIndex)
		if err == nil {
			absoluteStartTime += relativeIdx * 2
			absoluteStartTime %= 24
		}
	}
	if query.TimeIndex != "" {
		timeIdx, err := strconv.Atoi(query.TimeIndex)
		if err == nil {
			absoluteStartTime = (timeIdx - ((timeIdx + 1) % 2)) % 24
		}
	}
	logger.Sugar().Debugf("absolute start time: %d", absoluteStartTime)

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
