package balancer

import (
	"fmt"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

type Iterator struct {
	candidates      []model.GroupItem
	baseCandidates  []model.GroupItem
	index           int
	stickyChannelID int
	modelName       string
	groupID         int

	attemptMu sync.Mutex
	attempts  []model.ChannelAttempt
	count     int
}

func NewIterator(group model.Group, apiKeyID int, requestModel string) *Iterator {
	b := GetBalancer(group.Mode)
	base := b.Candidates(group.Items)
	return NewIteratorWithCandidates(group, apiKeyID, requestModel, base)
}

func NewIteratorWithCandidates(group model.Group, apiKeyID int, requestModel string, base []model.GroupItem) *Iterator {
	candidates := make([]model.GroupItem, len(base))
	copy(candidates, base)

	baseCandidates := make([]model.GroupItem, len(base))
	copy(baseCandidates, base)

	stickyChannelID := 0
	if group.SessionKeepTime > 0 {
		stickyTTL := time.Duration(group.SessionKeepTime) * time.Second
		if sticky := GetSticky(apiKeyID, requestModel, stickyTTL); sticky != nil {
			for i, item := range candidates {
				if item.ChannelID != sticky.ChannelID {
					continue
				}
				if i > 0 {
					stickyItem := candidates[i]
					copy(candidates[1:i+1], candidates[0:i])
					candidates[0] = stickyItem
				}
				stickyChannelID = sticky.ChannelID
				break
			}
		}
	}

	return &Iterator{
		candidates:      candidates,
		baseCandidates:  baseCandidates,
		index:           -1,
		stickyChannelID: stickyChannelID,
		modelName:       requestModel,
		groupID:         group.ID,
	}
}

func (it *Iterator) Reset() {
	it.index = -1
	it.candidates = make([]model.GroupItem, len(it.baseCandidates))
	copy(it.candidates, it.baseCandidates)
	if it.stickyChannelID > 0 {
		for i, item := range it.candidates {
			if item.ChannelID != it.stickyChannelID {
				continue
			}
			if i > 0 {
				stickyItem := it.candidates[i]
				copy(it.candidates[1:i+1], it.candidates[0:i])
				it.candidates[0] = stickyItem
			}
			return
		}
	}
}

func (it *Iterator) Next() bool {
	it.index++
	return it.index < len(it.candidates)
}

func (it *Iterator) Item() model.GroupItem {
	return it.candidates[it.index]
}

func (it *Iterator) CandidateAt(index int) (model.GroupItem, bool) {
	if index < 0 || index >= len(it.candidates) {
		return model.GroupItem{}, false
	}
	return it.candidates[index], true
}

func (it *Iterator) IsSticky() bool {
	return it.stickyChannelID > 0 && it.index >= 0 && it.index < len(it.candidates) && it.candidates[it.index].ChannelID == it.stickyChannelID
}

func (it *Iterator) Len() int {
	return len(it.candidates)
}

func (it *Iterator) Index() int {
	return it.index
}

func (it *Iterator) GroupID() int {
	return it.groupID
}

func (it *Iterator) Skip(channelID, channelKeyID int, channelName, msg string) {
	it.recordAttempt(model.ChannelAttempt{
		ChannelID:    channelID,
		ChannelKeyID: channelKeyID,
		ChannelName:  channelName,
		ModelName:    it.currentModelName(),
		Status:       model.AttemptSkipped,
		Sticky:       it.isStickyChannel(channelID),
		Msg:          msg,
	})
}

func (it *Iterator) Record(channelID, channelKeyID int, channelName, modelName string, status model.AttemptStatus, statusCode int, duration time.Duration, msg string) {
	durationMs := 0
	if duration > 0 {
		durationMs = int(duration.Milliseconds())
	}
	if modelName == "" {
		modelName = it.currentModelName()
	}
	it.recordAttempt(model.ChannelAttempt{
		ChannelID:    channelID,
		ChannelKeyID: channelKeyID,
		ChannelName:  channelName,
		ModelName:    modelName,
		Status:       status,
		StatusCode:   statusCode,
		Duration:     durationMs,
		Sticky:       it.isStickyChannel(channelID),
		Msg:          msg,
	})
}

func (it *Iterator) SkipCircuitBreak(channelID, channelKeyID int, channelName string) bool {
	modelName := it.currentModelName()
	tripped, remaining := IsTripped(channelID, channelKeyID, modelName)
	if !tripped {
		return false
	}
	msg := "circuit breaker tripped"
	if remaining > 0 {
		msg = fmt.Sprintf("circuit breaker tripped, remaining cooldown: %ds", int(remaining.Seconds()))
	}
	it.recordAttempt(model.ChannelAttempt{
		ChannelID:    channelID,
		ChannelKeyID: channelKeyID,
		ChannelName:  channelName,
		ModelName:    modelName,
		Status:       model.AttemptCircuitBreak,
		Sticky:       it.isStickyChannel(channelID),
		Msg:          msg,
	})
	return true
}

func (it *Iterator) StartAttempt(channelID, channelKeyID int, channelName string) *AttemptSpan {
	return &AttemptSpan{
		attempt: model.ChannelAttempt{
			ChannelID:    channelID,
			ChannelKeyID: channelKeyID,
			ChannelName:  channelName,
			ModelName:    it.currentModelName(),
			Sticky:       it.isStickyChannel(channelID),
		},
		startTime: time.Now(),
		iter:      it,
	}
}

func (it *Iterator) Attempts() []model.ChannelAttempt {
	it.attemptMu.Lock()
	defer it.attemptMu.Unlock()
	out := make([]model.ChannelAttempt, len(it.attempts))
	copy(out, it.attempts)
	return out
}

type AttemptSpan struct {
	attempt   model.ChannelAttempt
	startTime time.Time
	iter      *Iterator
	ended     bool
}

func (s *AttemptSpan) End(status model.AttemptStatus, statusCode int, msg string) {
	if s.ended {
		return
	}
	s.ended = true
	s.attempt.Status = status
	s.attempt.StatusCode = statusCode
	s.attempt.Duration = int(time.Since(s.startTime).Milliseconds())
	s.attempt.Msg = msg
	s.iter.recordAttempt(s.attempt)
}

func (s *AttemptSpan) Duration() time.Duration {
	return time.Since(s.startTime)
}

func (it *Iterator) recordAttempt(attempt model.ChannelAttempt) {
	it.attemptMu.Lock()
	defer it.attemptMu.Unlock()
	it.count++
	attempt.AttemptNum = it.count
	it.attempts = append(it.attempts, attempt)
}

func (it *Iterator) currentModelName() string {
	if it.index >= 0 && it.index < len(it.candidates) {
		if modelName := it.candidates[it.index].ModelName; modelName != "" {
			return modelName
		}
	}
	return it.modelName
}

func (it *Iterator) isStickyChannel(channelID int) bool {
	if it.stickyChannelID <= 0 {
		return false
	}
	return it.stickyChannelID == channelID
}
