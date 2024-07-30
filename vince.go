package vice

import (
	"github.com/blevesearch/vellum"
	"github.com/blevesearch/vellum/levenshtein"
	"github.com/gernest/vice/pkg/fst"
)

type Vice[T any] struct {
	get func(uint64) T
	fst *vellum.FST
	dfa *levenshtein.LevenshteinAutomatonBuilder
}

func New[T any](retrieval func(uint64) T) (*Vice[T], error) {
	dfa, err := levenshtein.NewLevenshteinAutomatonBuilder(2, false)
	if err != nil {
		return nil, err
	}
	return &Vice[T]{
		get: retrieval,
		fst: fst.Get(),
		dfa: dfa,
	}, nil
}

func (v *Vice[T]) Get(ua string) (result T, err error) {
	var (
		dfa *levenshtein.DFA
		it  *vellum.FSTIterator
	)
	dfa, err = v.dfa.BuildDfa(ua, 2)
	if err != nil {
		return
	}
	it, err = v.fst.Search(dfa, nil, nil)
	for err == nil {
		_, val := it.Current()
		result = v.get(val)
		return
	}
	return
}
