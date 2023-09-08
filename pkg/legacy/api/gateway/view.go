package gateway

import (
	"html/template"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering/themes"
)

type EdgeDescription struct {
	Ino      int64
	Key      psi.EdgeKey
	NodeType string
	ToPath   psi.Path
	ToLink   *cidlink.Link
}

var nodeEdgeListTemplate = template.Must(template.New("node-edge-list").Funcs(themes.GenericTemplateHelpers).Parse(`
{{ $ctx := . }}
<html>
	<head>
		<title>{{ $ctx.Node }} - PsiDB</title>
	</head>
	<body>
		<h1>{{ $ctx.Node }}</h1>
		<table>
			<thead>
				<tr>	
					<th>Ino</th>	
					<th>Name</th>
					<th>Type</th>
					<th>Path</th>
					<th>Link</th>
				</tr>
			</thead>
			<tbody>	
				{{- range .Edges }}	
				<tr>
					<td>{{ .Ino }}</td>
					<td><a href="/psi/{{ $ctx.CurrentPath }}/{{ .Key | psiPathEscape }}">{{ .Key }}</a></td>	
					<td>{{ .NodeType }}</td>
					<td><a href="/psi/{{ .ToPath | psiPathEscape }}">{{ .ToPath }}</a></td>
					<td><a href="/_objects/{{ .ToLink }}">{{ .ToLink }}</a></td>
				</tr>	
				{{- end }}
			</tbody>	
		</table>	
	</body>
</html>
`))
