{{define "json"}}
{{template "header" .}}

{{template "back" .}}

<textarea style="width: 100%; height: 500px; padding: 5px;">
{{.Data}}
</textarea>

{{template "back" .}}

{{template "footer" .}}
{{end}}
