{{- define "markdownDoc" -}}
**简要描述:**
> {{.Description}}

**协议:**
- ***路径:*** {{.Path}}
- ***方法:*** {{.Method}}


[在线调试]({{.GetFormPathWithQuery}})

{{ if .Service -}}

**服务部署:**
|名称|URL|说明|IP|http代理|
|:--|:--|:--|:--|:--|
{{range $server:= .Service.Servers -}}
|{{$server.Name}}|{{$server.URL}}|{{$server.Description}}|{{$server.IP}}|{{$server.Proxy}}|
{{end}}
{{- end}}


### 请求

***请求格式：*** {{.RequestContentType}}

{{if .RequestHeader }}

**请求Header头**
|参数名|类型|格式|必选|可空|标题|说明|默认值|示例|
|:---|:---|:---|:---|:---|:---|:---|:---|:---|
{{range $param:= .RequestHeader -}}
|{{$param.Fullname}}|{{$param.Type}}|{{$param.Format}}|{{$param.Required}}|{{$param.AllowEmptyValue}}|{{$param.Title}}|{{$param.Description}}|{{$param.Default}}|{{$param.Example}}|
{{end}}

{{- end}}

{{if .Query -}}

**请求Query**
|参数名|类型|格式|必选|可空|标题|说明|默认值|示例|
|:---|:---|:---|:---|:---|:---|:---|:---|:---|
{{range $param:= .Query -}}
|{{$param.Fullname}}|{{$param.Type}}|{{$param.Format}}|{{$param.Required}}|{{$param.AllowEmptyValue}}|{{$param.Title}}|{{$param.Description}}|{{$param.Default}}|{{$param.Example}}|
{{end}}

{{- end}}

{{if .RequestBody -}}

**请求Body**
|参数名|类型|格式|必选|可空|标题|说明|默认值|示例|
|:---|:---|:---|:---|:---|:---|:---|:---|:---|
{{range $param:= .RequestBody -}}
|{{$param.Fullname}}|{{$param.Type}}|{{$param.Format}}|{{$param.Required}}|{{$param.AllowEmptyValue}}|{{$param.Title}}|{{$param.Description}}|{{$param.Default}}|{{$param.Example}}|
{{end}}

**请求案例**
```json
{{$example:=.GetFirstExample}}
{{$example.RequestBody}}
```

{{- end}}

### 响应

***响应格式：*** {{.ResponseContentType}}

{{if .ResponseHeader }}

**响应Header头参数**
|参数名|类型|格式|必选|可空|标题|说明|默认值|示例|
|:---|:---|:---|:---|:---|:---|:---|:---|:---|
{{range $param:= .ResponseHeader -}}
|{{$param.Fullname}}|{{$param.Type}}|{{$param.Format}}|{{$param.Required}}|{{$param.AllowEmptyValue}}|{{$param.Title}}|{{$param.Description}}|{{$param.Default}}|{{$param.Example}}|
{{end}}

{{- end}}

{{if .ResponseBody -}}

**响应Body参数**
|参数名|类型|格式|标题|说明|示例|
|:---|:---|:---|:---|:---|:---|
{{range $param:= .ResponseBody -}}
|{{$param.Fullname}}|{{$param.Type}}|{{$param.Format}}|{{$param.Title}}|{{$param.Description}}|{{$param.Example}}|
{{end}}

**响应案例**
```json
{{$example:=.GetFirstExample}}
{{$example.Response}}
```

{{- end}}

**备注** 
{{end}}