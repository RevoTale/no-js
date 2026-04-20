package metagen

import "context"

type managedStylesheetsContextKey struct{}

func WithManagedStylesheets(ctx context.Context, stylesheets []string) context.Context {
	normalized := normalizeStylesheets(stylesheets)
	if len(normalized) == 0 {
		return ctx
	}

	existing := ManagedStylesheetsFromContext(ctx)
	merged := normalizeStylesheets(append(existing, normalized...))
	if len(merged) == 0 {
		return ctx
	}

	return context.WithValue(ctx, managedStylesheetsContextKey{}, merged)
}

func ManagedStylesheetsFromContext(ctx context.Context) []string {
	if ctx == nil {
		return nil
	}

	stylesheets, ok := ctx.Value(managedStylesheetsContextKey{}).([]string)
	if !ok || len(stylesheets) == 0 {
		return nil
	}

	out := make([]string, len(stylesheets))
	copy(out, stylesheets)
	return out
}

func MergeManagedStylesheets(ctx context.Context, meta Metadata) Metadata {
	stylesheets := ManagedStylesheetsFromContext(ctx)
	if len(stylesheets) == 0 {
		return meta
	}

	return Merge(Metadata{Stylesheets: stylesheets}, meta)
}
