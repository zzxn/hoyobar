package storage

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

var pageCursorTimeFormat = time.RFC3339Nano

func decomposePageCursor(cursor string) (ID int64, t time.Time, err error) {
	if cursor == "" {
		return math.MaxInt64, time.Now(), nil
	}
	segs := strings.SplitN(strings.ReplaceAll(cursor, "P", "+"), "_", 2)
	if len(segs) != 2 {
		return ID, t, fmt.Errorf("Wrong page cursor format: %v", cursor)
	}

	ID, err = strconv.ParseInt(segs[0], 10, 64)
	if err != nil {
		return ID, t, fmt.Errorf("wrong page cursor format, expect first part a int, got %v", segs[0])
	}

	t, err = time.Parse(pageCursorTimeFormat, segs[1])
	if err != nil {
		return ID, t, fmt.Errorf(
			"wrong page cursor format, expect second part time formatted, %v, got %v",
			pageCursorTimeFormat,
			segs[1],
		)
	}

	return ID, t, nil
}

func composePageCursor(ID int64, t time.Time) (cursor string) {
	cursor = fmt.Sprintf("%d_%v", ID, t.Format(pageCursorTimeFormat))
	cursor = strings.ReplaceAll(cursor, "+", "P") // make it url safe
	return cursor
}
