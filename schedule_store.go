package main

import (
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

type ScheduleStore struct {
	sync.RWMutex
	info        *AllScheduleInfo
	salmonInfo  *[]TimeSlotInfo
	cache       *FileCache
	salmonCache *FileCache
}

type SearchResultSlot struct {
	mode Mode
	tsi  *TimeSlotInfo
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
	Query *SearchQuery
	Found bool
	Slots []SearchResultSlot
}

func lookupByAbsoluteTime(asi *AllScheduleInfo, mode Mode, hour int) (matched SearchResultSlot, found bool) {
	tsinfos := asi.getTimeSlotInfoByMode(mode)
	logger.Debug("tsinfos", zap.Any("tsinfos", tsinfos))
	for _, tsinfo := range tsinfos {
		logger.Sugar().Infof("found => %d; req => %d\n", tsinfo.StartTime.Hour(), hour)
		if tsinfo.StartTime.Hour() == hour && !tsinfo.IsFest {
			logger.Debug("found", zap.Any("tsinfo", tsinfo))
			return SearchResultSlot{mode, &tsinfo}, true
		}
	}
	return SearchResultSlot{mode, nil}, false
}

func lookupByRule(asi *AllScheduleInfo, mode Mode, ruleKey string, skipCount int) (matched SearchResultSlot, found bool) {
	tsinfos := asi.getTimeSlotInfoByMode(mode)
	logger.Debug("tsinfos", zap.Any("tsinfos", tsinfos))
	for _, tsinfo := range tsinfos {
		if tsinfo.Rule.Key == ruleKey && !tsinfo.IsFest {
			if skipCount <= 0 {
				return SearchResultSlot{mode, &tsinfo}, true
			}
			skipCount -= 1
		}
	}
	return SearchResultSlot{mode, nil}, false
}

func (ss *ScheduleStore) Search(query *SearchQuery) SearchResult {
	ss.RLock()
	defer ss.RUnlock()
	if query.Mode.getIdentifier() == "SALMON" {
		return searchSalmon(query, ss.salmonInfo, time.Now())
	} else {
		sr := search(query, ss.info, time.Now())
		logger.Debug("search result", zap.Any("result", sr))
		return sr
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
		Query: query,
		Found: found,
		Slots: []SearchResultSlot{
			{getMode("SALMON"), result},
		},
	}
}

func search(query *SearchQuery, info *AllScheduleInfo, timeStamp time.Time) SearchResult {
	logger.Sugar().Infof("search request: %#v", *query)

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

		// XXX: special case using pseudo mode
		if query.Mode.getIdentifier() == "BYRULE" {
			matched1, found1 := lookupByRule(info, getMode("CHALLENGE"), query.Rule, skipCount)
			matched2, found2 := lookupByRule(info, getMode("OPEN"), query.Rule, skipCount)
			matched3, found3 := lookupByRule(info, getMode("X"), query.Rule, skipCount)
			logger.Debug("search result", zap.Any("matched1", matched1), zap.Any("matched2", matched2), zap.Any("matched3", matched3))
			return SearchResult{
				Query: query,
				Found: found1 || found2 || found3,
				Slots: []SearchResultSlot{matched1, matched2, matched3},
			}
		} else if query.Mode.getIdentifier() == "BANKARA" {
			matched1, found1 := lookupByRule(info, getMode("CHALLENGE"), query.Rule, skipCount)
			matched2, found2 := lookupByRule(info, getMode("OPEN"), query.Rule, skipCount)
			logger.Debug("search result", zap.Any("matched1", matched1), zap.Any("matched2", matched2))
			return SearchResult{
				Query: query,
				Found: found1 || found2,
				Slots: []SearchResultSlot{matched1, matched2},
			}
		} else if query.Rule == "TURF_WAR" {
			// TODO: support Splatfest schedule search
			matched, found := lookupByRule(info, getMode("REGULAR"), query.Rule, skipCount)
			logger.Debug("search result", zap.Any("matched", matched))
			return SearchResult{
				Query: query,
				Found: found,
				Slots: []SearchResultSlot{matched},
			}
		} else {
			matched, found := lookupByRule(info, query.Mode, query.Rule, skipCount)
			logger.Debug("search result", zap.Any("matched", matched))
			return SearchResult{
				Query: query,
				Found: found,
				Slots: []SearchResultSlot{matched},
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
		matched1, found1 := lookupByAbsoluteTime(info, getMode("CHALLENGE"), absoluteStartTime)
		matched2, found2 := lookupByAbsoluteTime(info, getMode("OPEN"), absoluteStartTime)
		logger.Debug("search result", zap.Any("matched1", matched1), zap.Any("matched2", matched2))
		return SearchResult{
			Query: query,
			Found: found1 || found2,
			Slots: []SearchResultSlot{matched1, matched2},
		}
	} else {
		// search case #3: lookup single by time
		matched, found := lookupByAbsoluteTime(info, query.Mode, absoluteStartTime)
		logger.Debug("search result", zap.Any("matched", matched))
		return SearchResult{
			Query: query,
			Found: found,
			Slots: []SearchResultSlot{matched},
		}
	}
}
