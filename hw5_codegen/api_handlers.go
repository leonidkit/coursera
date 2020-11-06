package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func (a *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		a.CreateHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		resp := map[string]string{
			"error": "unknown method",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
}

func (a *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		a.ProfileHandler(w, r)
	case "/user/create":
		a.CreateHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		resp := map[string]string{
			"error": "unknown method",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
}

func (m *MyApi)ProfileHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	params := new(ProfileParams)

		
	params.Login = r.FormValue("login")
			
	if params.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "login must me not empty",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
		

	ctx := r.Context()

	res, err := m.Profile(ctx, *params)
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

func (m *MyApi)CreateHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		resp := map[string]string{
			"error": "bad method",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		resp := map[string]string{
			"error": "unauthorized",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	params := new(CreateParams)

		
	params.Login = r.FormValue("login")
			
	if params.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "login must me not empty",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	if len(params.Login) < 10 {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "login len must be >= 10",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
			
		

		
	params.Name= r.FormValue("full_name")
			
		

		
	params.Status = r.FormValue("status")
			
	switch params.Status {
	case "user":
	case "moderator":
	case "admin":
	default:
		if params.Status != "" {
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]string{
				"error": "status must be one of [user, moderator, admin]",
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		params.Status = "user"
	}
		

		
	params.Age, err = strconv.Atoi(r.FormValue("age"))
			
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "age must be int",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	if params.Age > 128 {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "age must be <= 128",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	if params.Age < 0 {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "age must be >= 0",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
		

	ctx := r.Context()

	res, err := m.Create(ctx, *params)
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

func (m *OtherApi)CreateHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		resp := map[string]string{
			"error": "bad method",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		resp := map[string]string{
			"error": "unauthorized",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	params := new(OtherCreateParams)

		
	params.Username = r.FormValue("username")
			
	if params.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "username must me not empty",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	if len(params.Username) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "username len must be >= 3",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
			
		

		
	params.Name= r.FormValue("account_name")
			
		

		
	params.Class = r.FormValue("class")
			
	switch params.Class {
	case "warrior":
	case "sorcerer":
	case "rouge":
	default:
		if params.Class != "" {
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]string{
				"error": "class must be one of [warrior, sorcerer, rouge]",
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		params.Class = "warrior"
	}
		

		
	params.Level, err = strconv.Atoi(r.FormValue("level"))
			
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "level must be int",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	if params.Level > 50 {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "level must be <= 50",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	if params.Level < 1 {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{
			"error": "level must be >= 1",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
		

	ctx := r.Context()

	res, err := m.Create(ctx, *params)
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

