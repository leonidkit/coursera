package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type UserRow struct {
	Id        int    `xml:"id"`
	Guid      string `xml:"guid"`
	IsActive  bool   `xml:"isActive"`
	Balance   string `xml:"balance"`
	Picture   string `xml:"picture"`
	Age       int    `xml:"age"`
	EyeColor  string `xml:"eyeColor"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Gender    string `xml:"gender"`
	Company   string `xml:"company"`
	Email     string `xml:"email"`
	Phone     string `xml:"phone"`
	Address   string `xml:"address"`
	About     string `xml:"about"`
}

type UsersRow struct {
	Users []UserRow `xml:"row"`
}

func TestNegative(t *testing.T) {
	expectedErr := "limit must be > 0"

	searchClient := SearchClient{
		URL: "",
	}

	_, err := searchClient.FindUsers(SearchRequest{
		Limit: -1,
	})
	if err.Error() != expectedErr {
		t.Fatalf("Test case error, expected %#v, got %#v", expectedErr, err.Error())
	}

	expectedErr = "offset must be > 0"

	searchClient = SearchClient{
		URL: "",
	}

	_, err = searchClient.FindUsers(SearchRequest{
		Offset: -1,
	})
	if err.Error() != expectedErr {
		t.Fatalf("Test case error, expected %#v, got %#v", expectedErr, err.Error())
	}
}

func testTemplate(expectedErr string, handleFunc func(w http.ResponseWriter, r *http.Request), isContain bool) error {
	ts := httptest.NewServer(http.HandlerFunc(handleFunc))
	defer ts.Close()

	searchClient := SearchClient{
		URL: ts.URL,
	}
	_, err := searchClient.FindUsers(SearchRequest{})

	if isContain {
		if !strings.Contains(err.Error(), expectedErr) {
			return fmt.Errorf("Test case error, expected %#v, got %#v", expectedErr, err.Error())
		}
	} else {
		if err.Error() != expectedErr {
			return fmt.Errorf("Test case error, expected %#v, got %#v", expectedErr, err.Error())
		}
	}
	return nil
}

func TimeoutServer(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 2)
}

func TestTimeoutErrorServer(t *testing.T) {
	expectedErr := "timeout for"
	err := testTemplate(expectedErr, TimeoutServer, true)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestUnknowErrorServer(t *testing.T) {
	expectedErr := "unknown error"

	searchClient := SearchClient{
		URL: "",
	}
	_, err := searchClient.FindUsers(SearchRequest{})

	if !strings.Contains(err.Error(), expectedErr) {
		t.Fatalf("Test case error, expected %#v, got %#v", expectedErr, err.Error())
	}
}

func AccessControlServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
}
func TestAccessControlServer(t *testing.T) {
	expectedErr := "Bad AccessToken"

	err := testTemplate(expectedErr, AccessControlServer, false)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func InternalErrorServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}
func TestInternalErrorServer(t *testing.T) {
	expectedErr := "SearchServer fatal error"

	err := testTemplate(expectedErr, InternalErrorServer, false)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func UnpackBadRequestServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Unpack response"))
}

func TestUnpackBadRequestServer(t *testing.T) {
	expectedErr := "cant unpack error json"

	err := testTemplate(expectedErr, UnpackBadRequestServer, true)
	if err != nil {
		t.Fatal(err.Error())
	}

}

func OrderFieldErrorServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	searchResp := &SearchErrorResponse{
		Error: "ErrorBadOrderField",
	}
	resp, err := json.Marshal(searchResp)
	if err != nil {
		panic("cant pack error response")
	}
	w.Write(resp)
}

func TestOrderFieldErrorServer(t *testing.T) {
	expectedErr := "OrderFeld"

	err := testTemplate(expectedErr, OrderFieldErrorServer, true)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func UnknowBadRequestErrorServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	searchResp := &SearchErrorResponse{
		Error: "Another",
	}
	resp, err := json.Marshal(searchResp)
	if err != nil {
		panic("cant pack error response")
	}
	w.Write(resp)
}
func TestUnknowBadRequestErrorServer(t *testing.T) {
	expectedErr := "unknown bad request error"

	err := testTemplate(expectedErr, UnknowBadRequestErrorServer, true)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func UnpackJsonErrorServer(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Not json"))
}
func TestUnpackJsonErrorServer(t *testing.T) {
	expectedErr := "cant unpack result json"

	err := testTemplate(expectedErr, UnpackJsonErrorServer, true)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestFindUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	var sc = &SearchClient{
		URL: ts.URL,
	}

	testCases := []SearchRequest{
		{Limit: 0, Query: "Boyd"},
		{Limit: 30, Query: ""},
		{Limit: 30, Offset: 1},
		{Limit: 30, OrderField: ""},
	}

	for tnum, tcase := range testCases {
		_, err := sc.FindUsers(tcase)
		if err != nil {
			t.Fatalf("[%d] expected successful completion, but received an error: %s", tnum, err.Error())
		}
	}
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("dataset.xml")
	if err != nil {
		panic(err)
	}

	var users = &UsersRow{}

	err = xml.Unmarshal(data, &users)
	if err != nil {
		panic(err)
	}

	res := []User{}

	if r.FormValue("query") != "" {
		for _, val := range users.Users {
			if q := r.FormValue("query"); val.FirstName == q || val.LastName == q || strings.Contains(val.About, q) {
				res = append(res, User{
					Id:     val.Id,
					Name:   val.FirstName + " " + val.LastName,
					Age:    val.Age,
					About:  val.About,
					Gender: val.Gender,
				})
			}
		}
	} else {
		sort.Slice(users.Users, func(i, j int) bool {
			return users.Users[i].Id < users.Users[j].Id
		})
		for _, val := range users.Users {
			res = append(res, User{
				Id:     val.Id,
				Name:   val.FirstName + " " + val.LastName,
				Age:    val.Age,
				About:  val.About,
				Gender: val.Gender,
			})
		}
	}

	lim, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		panic(err)
	}

	off, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(res[off:lim])
}
