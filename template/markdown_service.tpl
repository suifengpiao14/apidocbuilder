{{- define "markdownService" -}}
# 项目接口文档
**描述**:
{{.Description}}

## 服务部署

|名称|URL|说明|IP|http代理|
|:--|:--|:--|:--|:--|
{{- if .Servers -}}
{{range $server:= .Servers }}
|{{$server.Name}}|{{$server.URL}}|{{$server.Description}}|{{$server.IP}}|{{$server.Proxy}}|
{{- end}}

{{- end}}


## 接口列表
{{$apis:=.Apis}}
{{- if $apis -}}
{{range $group:= .Apis.GetGroups -}}

{{$subApis:= $apis.GetByGroups $group}}
**{{$group}}**
{{range $api:= $subApis -}}
- [{{$api.TitleOrDescription}}]({{$api.DocumentRef}})
{{ end}}

------




{{end}}

{{- end}}
{{- end}}
