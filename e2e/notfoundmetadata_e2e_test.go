package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNotFoundMetadataFixtureApp(t *testing.T) {
	report := loadNotFoundMetadataFixture(t)

	require.Equal(t, 200, report.Home.Status)
	require.Contains(t, report.Home.Body, `<title data-metagen-managed="true">Not Found Metadata Home</title>`)

	require.Equal(t, 404, report.RootMissing.Status)
	require.Contains(t, report.RootMissing.Body, `id="root-not-found"`)
	require.Contains(t, report.RootMissing.Body, `Root missing /missing`)
	require.Contains(t, report.RootMissing.Body, `<title data-metagen-managed="true">Root 404 Metadata Title</title>`)
	require.Contains(
		t,
		report.RootMissing.Body,
		`<meta data-metagen-managed="true" name="description" content="Root 404 metadata description">`,
	)
	require.Contains(
		t,
		report.RootMissing.Body,
		`<meta data-metagen-managed="true" name="robots" content="noindex, nofollow">`,
	)
	require.Contains(
		t,
		report.RootMissing.Body,
		`<meta data-metagen-managed="true" property="og:title" content="Root 404 Metadata Title">`,
	)

	require.Equal(t, 404, report.DocsMissing.Status)
	require.Contains(t, report.DocsMissing.Body, `data-layout="docs"`)
	require.Contains(t, report.DocsMissing.Body, `id="docs-not-found"`)
	require.Contains(t, report.DocsMissing.Body, `Docs missing /docs/fail`)
	require.Contains(t, report.DocsMissing.Body, `<title data-metagen-managed="true">Docs 404 Metadata Title</title>`)
	require.Contains(
		t,
		report.DocsMissing.Body,
		`<meta data-metagen-managed="true" name="description" content="Docs layout metadata">`,
	)
	require.Equal(t, 1, strings.Count(report.DocsMissing.Body, `<title data-metagen-managed="true">`))
}
