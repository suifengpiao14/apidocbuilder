{{- define "markdownDoc" -}}
**简要描述:**
> {{.Description}}

**服务文档**:{{if .Service }} {{.Service.DocumentRef}}{{end}}

**协议:**
- ***路径:*** {{.Path}}
- ***方法:*** {{.Method}}

{{ if .Service -}}

**服务部署:**
|名称|URL|IP|http代理|说明|
|:--|:--|:--|:--|:--|
{{range $server:= .Service.Servers -}}
|{{$server.Name}}|{{$server.URL}}|{{$server.IP}}|{{$server.Proxy}}|{{$server.Description}}|
{{- end}}

{{- end}}

### 请求参数

{{if .RequestHeader -}}

**请求Header头参数**
|参数名|类型|格式|必选|可空|标题|说明|默认值|示例|
|:---|:---|:---|:---|:---|:---|:---|:---|:---|
{{range $param:= .RequestHeader -}}
|{{$param.Fullname}}|{{$param.Type}}|{{$param.Format}}|{{$param.Required}}|{{$param.AllowEmptyValue}}|{{$param.Title}}|{{$param.Description}}|{{$param.Default}}|{{$param.Example}}|
{{end}}

{{- end}}

{{if .Query -}}

**请求Query参数**
|参数名|类型|格式|必选|可空|标题|说明|默认值|示例|
|:---|:---|:---|:---|:---|:---|:---|:---|:---|
{{range $param:= .Query -}}
|{{$param.Fullname}}|{{$param.Type}}|{{$param.Format}}|{{$param.Required}}|{{$param.AllowEmptyValue}}|{{$param.Title}}|{{$param.Description}}|{{$param.Default}}|{{$param.Example}}|
{{end}}

{{- end}}

{{if .RequestBody -}}

**请求Body参数**
|参数名|类型|格式|必选|可空|标题|说明|默认值|示例|
|:---|:---|:---|:---|:---|:---|:---|:---|:---|
{{range $param:= .RequestBody -}}
|{{$param.Fullname}}|{{$param.Type}}|{{$param.Format}}|{{$param.Required}}|{{$param.AllowEmptyValue}}|{{$param.Title}}|{{$param.Description}}|{{$param.Default}}|{{$param.Example}}|
{{end}}

{{- end}}

### 响应参数

{{if .ResponseHeader -}}

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

{{- end}}

**备注** 
{{end}}