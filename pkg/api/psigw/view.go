package psigw

import (
	"fmt"
	"html/template"
	"net/url"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

var nodeEdgeListTemplate = template.Must(template.New("node-edge-list").Funcs(template.FuncMap{
	"psiPathEscape": func(v any) string {
		switch v := v.(type) {
		case psi.EdgeReference:
			return url.PathEscape(v.GetKey().String())

		case psi.PathElement:
			return url.PathEscape(v.String())

		case *psi.Path:
			return v.Format(true)

		case psi.Path:
			return v.Format(true)

		case string:
			return url.PathEscape(v)
		}

		panic(fmt.Errorf("invalid type %T", v))
	},
}).Parse(`
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
