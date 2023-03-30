package storage

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

var pageCursorTimeFormat = "2006-01-02-15-04-05.000" // pricision: ms

func DecomposePageCursor(cursor string) (ID int64, t time.Time, err error) {
	if cursor == "" {
		return math.MaxInt64, time.Now(), nil
	}
	segs := strings.SplitN(cursor, "_", 2)
	if len(segs) != 2 {
		return ID, t, fmt.Errorf("Wrong page cursor format: %v", cursor)
	}

	t, err = time.ParseInLocation(pageCursorTimeFormat, segs[0], time.UTC)
	if err != nil {
		return ID, t, fmt.Errorf(
			"wrong page cursor format, expect first part time formatted, %v, got %v",
			pageCursorTimeFormat,
			segs[1],
		)
	}

	ID, err = strconv.ParseInt(segs[1], 10, 64)
	if err != nil {
		return ID, t, fmt.Errorf("wrong page cursor format, expect second part a int, got %v", segs[0])
	}

	return ID, t, nil
}

// compose cursor, the dict order of cursor is same as the order of tuple (t, ID)
func ComposePageCursor(ID int64, t time.Time) (cursor string) {
	cursor = fmt.Sprintf("%v_%v", t.UTC().Format(pageCursorTimeFormat), ID) // UTC
	return cursor
}
