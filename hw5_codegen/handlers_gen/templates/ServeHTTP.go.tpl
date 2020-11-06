func (a *{{ .RecvType }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
    {{- range .Funcs }}
	case "{{.Instr.URL}}":
		a.{{.Name}}Handler(w, r)
    {{- end }}
	default:
		w.WriteHeader(http.StatusNotFound)
		resp := map[string]string{
			"error": "unknown method",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
}

