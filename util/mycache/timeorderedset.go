package mycache

import (
	"context"
	"time"
)

// TimeOrdersedSetCache allows users add/fetch items into/from
// a Time-Ordered Set (TOS) with a name to identify it.
//
// A value can only be associated with a single time.
// If a to-add value has alreadly associated with a time,
// then it's time will be updated to the new one.
//
// This interface is originally designed to fast paginate by time
// in a cursor-based "load more" way.

type TOSItem struct {
	T     time.Time
	Value string
}

type TimeOrderedSetCache interface {
	// Add (t, key, value) to the TOS with given name.
	// Remove the item with the largest (time, key) pair
	// if the size of this TOS exceeds maxSize.
	TOSAdd(ctx context.Context, name string, item TOSItem, maxSize int) error
	// Fetch item
	TOSFetch(ctx context.Context, name string, tCursor time.Time, valueCursor string, n int) ([]TOSItem, error)
}
