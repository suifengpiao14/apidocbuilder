{{- define "markdownList" -}}
# 项目接口文档

## 服务部署

|名称|URL|IP|http代理|说明|
|:--|:--|:--|:--|:--|
{{- if .Servers -}}
{{range $server:= .Servers -}}
|{{$server.Name}}|{{$server.URL}}|{{$server.IP}}|{{$server.Proxy}}|{{$server.Description}}|
{{- end}}

{{- end}}


## 接口列表

{{- if .Apis -}}
{{range $group:= .Apis.GetGroups -}}
{{$apis:= .Apis.GetByGroup $group}}
{{range $api:= $apis -}}
- [{{$api.DocumentRef}}]({{$api.Title}})
{{- end}}

{{end}}
{{- end}}
