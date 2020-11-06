func (m *{{- .Func.RecvType }}) {{- .Func.Name }}Handler(w http.ResponseWriter, r *http.Request) {
	var err error

    {{- if eq .Func.Instr.Method "POST"}}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		resp := map[string]string{
			"error": "bad method",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
    {{- else if eq .Func.Instr.Method "GET"}}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		resp := map[string]string{
			"error": "bad method",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
    {{- end }}
    
    {{- if .Func.Instr.Auth }}
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		resp := map[string]string{
			"error": "unauthorized",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
    {{- end }}

	params := new({{- .Func.ParamType }})

    {{- range .Param.Fields }}

		{{ if eq .Type "int"}}
			{{- if ne ($p := GetTag "apivalidator" .Tag "paramname") "" }}
	params.{{- .Name }}, err = strconv.Atoi(r.FormValue("{{- $p }}"))
			{{- else }}
	params.{{- .Name }}, err = strconv.Atoi(r.FormValue("{{- ToLower .Name }}"))
			{{ end }}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "{{- ToLower .Name}} must be int",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
        {{- else }}
			{{- if ne ($p := GetTag "apivalidator" .Tag "paramname") "" }}
	params.{{- .Name }}= r.FormValue("{{- $p }}")
			{{ else }}
	params.{{- .Name }} = r.FormValue("{{- ToLower .Name }}")
			{{ end }}
        {{- end }}

        {{- if eq (GetTag "apivalidator" .Tag "required") "required" }}
            {{- if eq .Type "string" }}
	if params.{{- .Name }} == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "{{- ToLower .Name }} must me not empty",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
            {{- else if eq .Type "int" }}
	if params.{{- .Name }} == 0 {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "login must me not empty",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
            {{- end }}
        {{- end }}

        {{- if ne ($p := GetTag "apivalidator" .Tag "enum") "" }}
	switch params.{{- .Name }} {
            {{- range $i, $el := GetRange $p}}
	case "{{- $el }}":
            {{- end }}
	default:
		if params.{{- .Name }} != "" {
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]string{
                {{- $r := GetRange $p}}
				"error": "{{- ToLower .Name}} must be one of [{{- range $i, $el := $r}}{{- if (Last $i $r) }}{{- $el }}{{- else }}{{- $el }}, {{ end }}{{- end }}]",
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		params.{{- .Name }} = "{{- GetTag "apivalidator" .Tag "default" }}"
	}
        {{- end }}

        {{- if eq .Type "int" }}
			{{- if ne ($p := GetTag "apivalidator" .Tag "max") "" }}
	if params.{{- .Name }} > {{ $p }} {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "{{- ToLower .Name}} must be <= {{ $p }}",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
            {{- end }}

            {{- if ne ($p := GetTag "apivalidator" .Tag "min") "" }}
	if params.{{- .Name }} < {{ $p }} {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "{{- ToLower .Name}} must be >= {{ $p }}",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
            {{- end }}
		{{ else if eq .Type "string" }}
        	{{- if (ne ($p := GetTag "apivalidator" .Tag "min") "")}}
	if len(params.{{- .Name }}) < {{ $p }} {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "{{- ToLower .Name }} len must be >= {{ $p }}",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
			{{ end }}

			{{- if (ne ($p := GetTag "apivalidator" .Tag "max") "")}}
	if len(params.{{- .Name }}) < {{ $p }} {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "{{- ToLower .Name }} len must be <= {{ $p }}",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
			{{ end }}
		{{ end }}
    {{- end }}

	ctx := r.Context()

	res, err := m.{{ .Func.Name }}(ctx, *params)
	if err != nil {
		if e, ok := err.(ApiError); ok {
			w.WriteHeader(e.HTTPStatus)
			resp := make(map[string]interface{})
			resp["error"] = e.Error()
			json.NewEncoder(w).Encode(resp)
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			resp := make(map[string]interface{})
			resp["error"] = err.Error()
			json.NewEncoder(w).Encode(resp)
			return
		}
	}

	mp := map[string]interface{}{
		"error":    "",
		"response": res,
	}
	json.NewEncoder(w).Encode(mp)
}

