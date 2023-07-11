package psigw

import (
	"html/template"

	"github.com/greenboxal/agibootstrap/pkg/psi/rendering/themes"
)

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
					<th>Path</th>
					<th>Link</th>
				</tr>
			</thead>
			<tbody>	
				{{- range .Edges }}	
				<tr>
					<td>{{ .ToIndex }}</td>
					<td><a href="/psi/{{ $ctx.CurrentPath }}/{{ .Key | psiPathEscape }}">{{ .Key }}</a></td>	
					<td><a href="/psi/{{ .ToPath | psiPathEscape }}">{{ .ToPath }}</a></td>
					<td><a href="/_objects/{{ .ToLink }}">{{ .ToLink }}</a></td>
				</tr>	
				{{- end }}
			</tbody>	
		</table>	
	</body>
</html>
`))
