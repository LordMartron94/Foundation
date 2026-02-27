// Package domain defines ordered, discrete value spaces for use in pattern and lexer compilation.
//
// A discrete domain has a minimum and maximum value, a total ordering, and successor/previous
// functions so that algorithms can enumerate ranges and detect adjacency. DiscreteDomain is
// the main type: it holds Min, Max, OrderingCmp, NextFn, and PreviousFn. Callers typically
// use a single DiscreteDomain instance instead of passing separate cmp and successor
// functions (e.g. in autarch/pattern and lexarch).
//
// Use DiscreteDomainRuneCreate, DiscreteDomainByteCreate, or DiscreteDomainIntCreate for
// built-in element types; for custom types, construct a DiscreteDomain with the appropriate
// fields set.
package domain
