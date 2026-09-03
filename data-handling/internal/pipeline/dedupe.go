package pipeline

// Dedupe removes repeated products, keeping the first occurrence.
//
// Identity is source-scoped: the key is source + "|" + id. The same id from two
// different sources is treated as two distinct products (we do not merge across
// heterogeneous upstreams). Duplicates within a single source (e.g. pagination
// overlap or a looping cursor) are collapsed.
func Dedupe(in []Product) []Product {
	seen := make(map[string]struct{}, len(in))
	out := make([]Product, 0, len(in))
	removed := 0

	for _, p := range in {
		key := p.Source + "|" + p.ID
		if _, ok := seen[key]; ok {
			removed++
			logger.Warn("dropping duplicate", "source", p.Source, "id", p.ID)
			continue
		}
		seen[key] = struct{}{}
		out = append(out, p)
	}

	if removed > 0 {
		logger.Info("dedupe complete", "duplicates_removed", removed, "unique_products", len(out))
	}
	return out
}
