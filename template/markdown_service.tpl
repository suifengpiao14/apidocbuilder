{{- define "markdownService" -}}
# 项目接口文档
**描述**:
{{.Description}}

## 服务部署

|名称|URL|IP|http代理|说明|
|:--|:--|:--|:--|:--|
{{- if .Servers -}}
{{range $server:= .Servers }}
|{{$server.Name}}|{{$server.URL}}|{{$server.IP}}|{{$server.Proxy}}|{{$server.Description}}|
{{- end}}

{{- end}}


## 接口列表
{{$apis:=.Apis}}
{{- if $apis -}}
{{range $group:= .Apis.GetGroups -}}

{{$subApis:= $apis.GetByGroups $group}}

{{range $api:= $subApis -}}
- [{{$api.TitleOrDescription}}]({{$api.DocumentRef}})
{{ end}}





{{end}}

{{- end}}
{{- end}}
