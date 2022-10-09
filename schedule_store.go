package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type ScheduleStore struct {
	sync.RWMutex
	info        *AllScheduleInfo
	salmonInfo  *[]TimeSlotInfo
	cache       *FileCache
	salmonCache *FileCache
}

func NewScheduleStore() ScheduleStore {
	return ScheduleStore{
		cache:       NewFileCache("./", "api_call_cache"),
		salmonCache: NewFileCache("./", "api_call_cache_salmon"),
	}
}

func (ss *ScheduleStore) maybeLoadInfo() {
	ss.Lock()
	defer ss.Unlock()
	cached := ss.cache.MaybeGet(time.Minute * 30)
	if cached == nil {
		// outdated. refresh schedule info
		logger.Sugar().Infof("Cache %s is outdated. fetching...", ss.cache.CacheFileName)
		info, err := FetchScheduleInfo()
		if err == nil {
			logger.Sugar().Infof("Fetch %s completed", ss.cache.CacheFileName)
		} else {
			logger.Sugar().Errorf("Fetch %s failed: %#v", ss.cache.CacheFileName, err)
		}
		ss.cache.Put(info)
		ss.info = info
	} else {
		logger.Sugar().Infof("Cache %s is valid", ss.cache.CacheFileName)
		c, ok := cached.(*AllScheduleInfo)
		if ok {
			ss.info = c
		} else {
			// XXX: convert map[string]interface{} to struct by re-decoding json
			var asi AllScheduleInfo
			bytes, err := json.Marshal(cached)
			if err != nil {
				logger.Sugar().Errorw("Error while preparing loading from cache", err)
			}
			err = json.Unmarshal(bytes, &asi)
			if err != nil {
				logger.Sugar().Errorw("Error while loading from cache", err)
			}
			ss.info = &asi
		}
	}
}

func (ss *ScheduleStore) maybeLoadInfoSalmon() {
	ss.Lock()
	defer ss.Unlock()
	cached := ss.salmonCache.MaybeGet(time.Minute * 30)
	if cached == nil {
		// outdated. refresh schedule info
		logger.Sugar().Infof("Cache %s is outdated. fetching...", ss.salmonCache.CacheFileName)
		info, err := FetchScheduleInfoSalmon()
		if err == nil {
			logger.Sugar().Infof("Fetch %s completed", ss.salmonCache.CacheFileName)
		} else {
			logger.Sugar().Errorf("Fetch %s failed: %#v", ss.salmonCache.CacheFileName, err)
		}
		ss.salmonCache.Put(info)
		ss.salmonInfo = info
	} else {
		logger.Sugar().Infof("Cache %s is valid", ss.salmonCache.CacheFileName)
		c, ok := cached.(*[]TimeSlotInfo)
		if ok {
			ss.salmonInfo = c
		} else {
			// XXX: convert map[string]interface{} to struct by re-decoding json
			var tsi []TimeSlotInfo
			bytes, err := json.Marshal(cached)
			if err != nil {
				logger.Sugar().Errorw("Error while preparing loading from cache", err)
			}
			err = json.Unmarshal(bytes, &tsi)
			if err != nil {
				logger.Sugar().Errorw("Error while loading from cache", err)
			}
			ss.salmonInfo = &tsi
		}
	}
}

func (ss *ScheduleStore) MaybeRefresh() {
	ss.maybeLoadInfo()
	ss.maybeLoadInfoSalmon()
}

type SearchResult struct {
	Query      *SearchQuery
	Found      bool
	IsTwoSlots bool
	Slot1      *TimeSlotInfo
	Slot2      *TimeSlotInfo
}

func lookupByAbsoluteTime(tsinfos []TimeSlotInfo, hour int) (matched *TimeSlotInfo, found bool) {
	fmt.Printf("%#v", tsinfos)
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
	if query.Mode == "SALMON" {
		return searchSalmon(query, ss.salmonInfo, time.Now())
	} else {
		return search(query, ss.info, time.Now())
	}
}

func searchSalmon(query *SearchQuery, salmonInfo *[]TimeSlotInfo, timeStamp time.Time) SearchResult {
	relativeIdx, err := strconv.Atoi(query.RelativeIndex)
	if err != nil {
		relativeIdx = 0
	}
	return SearchResult{
		Query:      query,
		Found:      relativeIdx < 5,
		IsTwoSlots: false,
		Slot1:      &(*salmonInfo)[relativeIdx],
	}
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
