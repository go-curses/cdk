package ptypes

var (
	cdkQuarkMap = make(map[QuarkID]*CQuark)
)

// Quarks are associations between strings and integer identifiers. Given either
// the string or the QuarkID identifier it is possible to retrieve the other.
//
// Quarks are used for both datasets and keyed data lists.
//
// To create a new quark from a string, use QuarkFromString().
//
// To find the string corresponding to a given QuarkID, use QuarkID.ToString().
//
// To find the QuarkID corresponding to a given string, use QuarkID.TryString().
type CQuark struct {
	id QuarkID
	v  string
}

func (q CQuark) ID() QuarkID {
	return q.id
}

func (q CQuark) String() string {
	return q.v
}

// Gets the QuarkID identifying the given string. If the string does not
// currently have an associated QuarkID, a new QuarkID is created, using a copy
// of the string.
//
// This function must not be used before library constructors have finished
// running.
//
// Parameters
// string	a string.
//
// Returns
//     the QuarkID identifying the string, or 0 if string is nil
func QuarkFromString(text string) (qid QuarkID) {
	for id, cq := range cdkQuarkMap {
		if cq.v == text {
			return id
		}
	}
	next := QuarkID(0)
	for id, _ := range cdkQuarkMap {
		if next <= id {
			next = id + 1
		}
	}
	cdkQuarkMap[next] = &CQuark{
		id: next,
		v:  text,
	}
	return next
}

// Gets the string associated with the given GQuark.
//
// Parameters:
//     quark    a GQuark.
//
// Returns:
//     the string associated with the GQuark
func QuarkToString(id QuarkID) string {
	if q, ok := cdkQuarkMap[id]; ok {
		return q.v
	}
	return ""
}

// Gets the GQuark associated with the given string, or 0 if string is nil or it
// has no associated QuarkID.
//
// If you want the GQuark to be created if it doesn't already exist, use
// QuarkFromString().
//
// This function must not be used before library constructors have finished
// running.
//
// Parameters
//     string     a string.
//
// Returns
//     the GQuark associated with the string, or 0 if string is nil or there is
//     no GQuark associated with it
func QuarkTryString(text string) QuarkID {
	for id, cq := range cdkQuarkMap {
		if cq.v == text {
			return id
		}
	}
	return QuarkID(0)
}

// A QuarkID is a non-zero integer which uniquely identifies a particular string.
// A QuarkID value of zero is associated to nil.
type QuarkID uint64
