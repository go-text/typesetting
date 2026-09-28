package fontscan

import "testing"

func TestRuneLRUReplacement(t *testing.T) {
	l := runeLRU{maxSize: 1}
	first, second := Query{Families: []string{"ab", "c"}}, Query{Families: []string{"a", "bc"}}
	key := l.KeyFor(first, 0, 'A')
	if key == l.KeyFor(second, 0, 'A') {
		t.Fatal("family boundaries omitted from hash")
	}
	// Even a genuine hash collision must replace the old list node.
	for i := 0; i < 100; i++ {
		q := first
		if i%2 != 0 {
			q = second
		}
		l.Put(key, q, nil)
		if _, ok := l.Get(key, q); !ok {
			t.Fatal("replacement was not cached")
		}
		if len(l.m) != 1 || l.tail.next.next != l.head {
			t.Fatal("replaced cache entry retained in list")
		}
	}
	if _, ok := l.Get(key, first); ok {
		t.Fatal("hash collision returned the wrong families")
	}
	l.Put(l.KeyFor(Query{}, 0, 'B'), Query{}, nil)
	if _, ok := l.Get(key, second); ok {
		t.Fatal("replaced entry was not evicted")
	}
}
