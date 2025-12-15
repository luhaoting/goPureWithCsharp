package main

import (
	"sync/atomic"
)

var globalBattleID int64 = 1000

func GenerateBattleID() int64 {
	battleId := atomic.AddInt64(&globalBattleID, 1)
	return battleId
}
