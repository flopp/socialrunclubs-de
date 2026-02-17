package utils

import (
	"math/rand"
	"time"
)

func RandomSleep(minSeconds, maxSeconds int) {
	if minSeconds <= 0 {
		minSeconds = 1
	}
	if maxSeconds < minSeconds {
		maxSeconds = minSeconds
	}

	sleepDuration := time.Duration(rand.Intn(maxSeconds-minSeconds+1)+minSeconds) * time.Second
	time.Sleep(sleepDuration)
}
