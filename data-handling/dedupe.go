package main

import "log"

// dedupe removes repeated products, keeping the first occurrence.
//
// Identity is source-scoped: the key is source + "|" + id. The same id from two
// different sources is treated as two distinct products (we do not attempt to
// merge across heterogeneous upstreams). Duplicates within a single source
// (e.g. pagination overlap or a looping cursor) are collapsed.
func dedupe(in []Product) []Product {
	seen := make(map[string]struct{}, len(in))
	out := make([]Product, 0, len(in))
	dropped := 0

	for _, p := range in {
		key := p.Source + "|" + p.ID
		if _, ok := seen[key]; ok {
			dropped++
			log.Printf("dedupe: dropping duplicate source=%s id=%s", p.Source, p.ID)
			continue
		}
		seen[key] = struct{}{}
		out = append(out, p)
	}

	if dropped > 0 {
		log.Printf("dedupe: removed %d duplicate(s), %d unique products", dropped, len(out))
	}
	return out
}
