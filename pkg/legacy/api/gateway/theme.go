package gateway

import (
	"html/template"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphindex"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering/themes"
)

var ApiTheme = rendering.BuildTheme(
	rendering.InheritTheme(themes.GlobalTheme),

	rendering.WithSkinFunc(
		SearchQueryType,
		"text/markdown",
		"",
		func(ctx rendering.SkinRendererContext, node *SearchQuery) error {
			_, err := ctx.Buffer.WriteString(node.Query)
			return err
		},
	),

	rendering.WithSkinFunc(
		SearchResultListType,
		"text/html",
		"",
		func(ctx rendering.SkinRendererContext, node *SearchResultList) error {
			return searchResultListTemplate.Execute(ctx.Buffer, struct {
				Hits []graphindex.NodeSearchHit
			}{
				Hits: node.Hits,
			})
		},
	),
)

var searchResultListTemplate = template.Must(template.New("search-result-list").Funcs(themes.GenericTemplateHelpers).Parse(`
	<table>
		<thead>
			<tr>	
				<th>Ino</th>	
				<th>Score</th>	
				<th>Path</th>
				<th>Link</th>
				<th>Embeddings</th>
			</tr>
		</thead>
		<tbody>	
			{{- range .Hits }}	
			<tr>
				<td>{{ .Index }}</td>
				<td>{{ .Score }}</td>
				<td><a href="/psi/{{ .Path | psiPathEscape }}">{{ .Path }}</a></td>
				<td><a href="/_objects/{{ .Link }}">{{ .Link }}</a></td>
			</tr>	
			{{- end }}
		</tbody>	
	</table>	
`))
