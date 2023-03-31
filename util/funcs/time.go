package funcs

import "time"

// TimeInMs truncates time in ms.
//
// Sometimes higher precision than ms gets trouble.
// For MySQL, if the precision is higher than field settings,
// it will round time while some other databases will truncate it.
func TimeInMs(t time.Time) time.Time {
	return t.Truncate(time.Millisecond)
}

func NowInMs() time.Time {
	return TimeInMs(time.Now())
}
