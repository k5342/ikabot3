package main

import (
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
	cached := MaybeGetFromFileCache[AllScheduleInfo](ss.cache, time.Minute*30)
	if cached == nil {
		// outdated. refresh schedule info
		logger.Sugar().Infof("Cache %s is outdated. fetching...", ss.cache.CacheFileName)
		info, err := FetchScheduleInfo()
		if err == nil {
			logger.Sugar().Infof("Fetch %s completed", ss.cache.CacheFileName)
		} else {
			logger.Sugar().Errorf("Fetch %s failed: %#v", ss.cache.CacheFileName, err)
			return
		}
		_, err = ss.cache.Put(info)
		if err != nil {
			return
		}
		ss.info = info
	} else {
		logger.Sugar().Infof("Cache %s is valid", ss.cache.CacheFileName)
		ss.info = cached
	}
}

func (ss *ScheduleStore) maybeLoadInfoSalmon() {
	ss.Lock()
	defer ss.Unlock()
	cached := MaybeGetFromFileCache[[]TimeSlotInfo](ss.salmonCache, time.Minute*30)
	if cached == nil {
		// outdated. refresh schedule info
		logger.Sugar().Infof("Cache %s is outdated. fetching...", ss.salmonCache.CacheFileName)
		info, err := FetchScheduleInfoSalmon()
		if err == nil {
			logger.Sugar().Infof("Fetch %s completed", ss.salmonCache.CacheFileName)
		} else {
			logger.Sugar().Errorf("Fetch %s failed: %#v", ss.salmonCache.CacheFileName, err)
		}
		_, err = ss.salmonCache.Put(info)
		if err != nil {
			return
		}
		ss.salmonInfo = info
	} else {
		logger.Sugar().Infof("Cache %s is valid", ss.salmonCache.CacheFileName)
		ss.salmonInfo = cached
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
	if query.Mode.getIdentifier() == "SALMON" {
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
	found := relativeIdx < len(*salmonInfo)
	var result *TimeSlotInfo
	if found {
		result = &(*salmonInfo)[relativeIdx]
	} else {
		result = nil
	}
	return SearchResult{
		Query:      query,
		Found:      found,
		IsTwoSlots: false,
		Slot1:      result,
	}
}

func search(query *SearchQuery, info *AllScheduleInfo, timeStamp time.Time) SearchResult {
	logger.Sugar().Infof("search request: %#v", *query)
	var target []TimeSlotInfo
	if query.Mode.getIdentifier() == "REGULAR" {
		target = info.Regular
	}
	if query.Mode.getIdentifier() == "CHALLENGE" {
		target = info.BankaraChallenge
	}
	if query.Mode.getIdentifier() == "OPEN" {
		target = info.BankaraOpen
	}
	if query.Mode.getIdentifier() == "X" {
		target = info.XMatch
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

		if query.Mode.getIdentifier() == "BANKARA" {
			matched1, found1 := lookupByRule(info.BankaraChallenge, query.Rule, skipCount)
			matched2, found2 := lookupByRule(info.BankaraOpen, query.Rule, skipCount)
			return SearchResult{
				Query:      query,
				Found:      found1 || found2,
				IsTwoSlots: true,
				Slot1:      matched1,
				Slot2:      matched2,
			}
		} else if query.Rule == "TURF_WAR" {
			// TODO: support Splatfest schedule search
			matched, found := lookupByRule(info.Regular, query.Rule, skipCount)
			return SearchResult{
				Query:      query,
				Found:      found,
				IsTwoSlots: false,
				Slot1:      matched,
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

	if query.Mode.getIdentifier() == "BANKARA" {
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
