package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const allowedAccessToken = "c2VjcmV0"

// some errors for forwarding to client.
var (
	errSorting  = errors.New("ErrorBadOrderField")
	errAuth     = errors.New("Bad AccessToken")
	errSource   = errors.New("source error - data source unavailable")
	errPaginate = errors.New("limit error - params must be positive integer")
)

// UserSet its struct to parsing xml dataset.
type UserSet struct {
	Users []UserData `xml:"row"`
}

// UserData its struct to parsing xml data.
type UserData struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

// Require search any contains in user struct.
func (s *UserSet) Require(query string) ([]User, bool) {
	reqSet := make([]User, 0)
	for _, u := range s.Users {
		if strings.Contains(u.FirstName, query) ||
			strings.Contains(u.LastName, query) ||
			strings.Contains(u.About, query) {

			reqSet = append(reqSet, User{
				Id:     u.Id,
				Name:   u.FirstName + " " + u.LastName,
				Age:    u.Age,
				About:  u.About,
				Gender: u.Gender,
			})
		}
	}
	isMatch := len(reqSet) > 0
	return reqSet, isMatch
}

// Sort required users by field (asc|desc) if needed
// TO-DO : make more clear through the sort.interface.
func Sort(data []User, order, field string) ([]User, error) {
	if order != "0" {
		var Less func(i, j User) bool
		switch field {
		case "Id":
			Less = func(i, j User) bool { return i.Id < j.Id }
		case "Age":
			Less = func(i, j User) bool { return i.Age < j.Age }
		case "Name", "":
			Less = func(i, j User) bool { return i.Name < j.Name }
		default:
			return data, errors.New("ErrorBadOrderField")
		}

		sort.Slice(data, func(i, j int) bool {
			if order == "1" {
				return Less(data[i], data[j])
			}
			return !Less(data[i], data[j])
		})
		return data, nil
	}
	return data, nil
}

// Limit - paginate reqired users to response.
func Limit(data []User, limit, offset string) ([]User, error) {
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 0 {
		return []User{}, err
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 0 {
		return []User{}, err
	}

	switch {
	case offsetInt >= len(data):
		return []User{}, nil
	case offsetInt+limitInt > len(data):
		return data[offsetInt:], nil
	default:
		return data[offsetInt : offsetInt+limitInt], nil
	}
}

// pushError helper func to sending errors in pesponse body.
func pushError(w http.ResponseWriter, _err error, code int) {
	errResp, err := json.Marshal(SearchErrorResponse{Error: _err.Error()})
	if err != nil {
		log.Printf("problem with decoding in push error, %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(errResp)
	if err != nil {
		log.Printf("problem with writing result , %v\n", err)
		return
	}
}

// pushResponse helper func to sending pesponse body.
func pushResponse(w http.ResponseWriter, users []User) error {
	respJSON, err := json.Marshal(users)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respJSON)
	if err != nil {
		log.Printf("problem with writing result , %v\n", err)
		return err
	}
	return nil
}

// SearchServer implementation server to test our client.
func SearchServer(w http.ResponseWriter, r *http.Request) {
	req := r.URL.Query()
	var (
		query      = req.Get("query")
		orderBy    = req.Get("order_by")
		orderField = req.Get("order_field")
		limit      = req.Get("limit")
		offset     = req.Get("offset")
	)

	// Check client credentials.
	if r.Header.Get("AccessToken") != allowedAccessToken {
		log.Printf("inconsist access attempt: %v\n", r.RemoteAddr)
		pushError(w, errAuth, http.StatusUnauthorized)
		return
	}

	// Read data xml-file and close it.
	file, err := os.Open("dataset.xml")
	if err != nil {
		log.Printf("opening file was failed: %v\n", err)
		pushError(w, errSource, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Read opened file in byte slice
	// TO-DO : Buffered parse and search "query" inplace!
	byteSlice, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("reading file was failed: %v\n", err)
		pushError(w, errSource, http.StatusInternalServerError)
		return
	}

	// Unmarshaling in user struct.
	users := UserSet{}
	err = xml.Unmarshal(byteSlice, &users)
	if err != nil {
		log.Printf("decoding file was failed: %v\n", err)
		pushError(w, errSource, http.StatusInternalServerError)
		return
	}

	required, ok := users.Require(query)
	if !ok {
		log.Println("quary not matched, empty result")
	}

	// Sort required users.
	sorted, err := Sort(required, orderBy, orderField)
	if err != nil {
		log.Printf("can't sorting reqired set: %v\n", err)
		pushError(w, errSorting, http.StatusBadRequest)
		return
	}

	// Paginate resopnse.
	limited, err := Limit(sorted, limit, offset)
	if err != nil {
		log.Printf("can't paginate reqired set: %v\n", err)
		pushError(w, errPaginate, http.StatusBadRequest)
		return
	}

	err = pushResponse(w, limited)
	if err != nil {
		pushError(w, err, http.StatusInternalServerError)
	}
}

type TestServer struct {
	Server *httptest.Server
	Client SearchClient
}

func NewTestServer(token string) TestServer {
	Server := httptest.NewServer(http.HandlerFunc(SearchServer))
	Client := SearchClient{token, Server.URL}
	return TestServer{Server, Client}
}

func (ts *TestServer) Close() {
	ts.Server.Close()
}

// TestInvalidToken testing authtorization.
func TestInvalidToken(t *testing.T) {
	ts := NewTestServer("fake_token")
	defer ts.Close()

	_, err := ts.Client.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("empty error, expected error: %v\n", errAuth.Error())
	} else if err.Error() != errAuth.Error() {
		t.Errorf("invalid error: %v, expected error: %v\n", err.Error(), errAuth.Error())
	}
}

// TestTimeout testing client timeout case.
func TestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	client := SearchClient{allowedAccessToken, server.URL}
	defer server.Close()

	_, err := client.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("empty error, must be timeout error")
	} else if !strings.Contains(err.Error(), "timeout for") {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

// TestMainFuture testing search.
func TestMainFuture(t *testing.T) {
	ts := NewTestServer(allowedAccessToken)
	defer ts.Close()

	resp, err := ts.Client.FindUsers(SearchRequest{
		Query:   "Sims",
		Limit:   3,
		OrderBy: OrderByDesc,
	})

	if len(resp.Users) != 1 {
		t.Errorf("invalid num of users: %d expected num : 1", len(resp.Users))
	}

	if resp.Users[0].Name != "Sims Cotton" {
		t.Errorf("invalid user found: %v, expected user: Sims Cotton", resp.Users[0])
	}

	if err != nil {
		t.Errorf("something gonna wrong, error: %v", err)
	}
}

// TestBadRequest testing handling 404 status.
func TestBadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pushError(w, errPaginate, http.StatusBadRequest)
	}))
	client := SearchClient{allowedAccessToken, server.URL}
	defer server.Close()

	_, err := client.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("empty error")
	} else if !strings.Contains(err.Error(), "unknown bad request error") {
		t.Errorf("invalid error: %v", err.Error())
	}
}

// TestPaginateParams testing paginate params.
func TestPaginateParams(t *testing.T) {
	type TestCase struct {
		req SearchRequest
		res *SearchResponse
		err error
	}

	cases := []TestCase{
		{
			req: SearchRequest{Limit: -5, Offset: 3},
			err: fmt.Errorf("limit must be > 0"),
		},
		{
			req: SearchRequest{Offset: -3, Limit: 2},
			err: fmt.Errorf("offset must be > 0"),
		},
	}

	ts := NewTestServer(allowedAccessToken)
	defer ts.Close()

	for caseNum, item := range cases {
		resp, err := ts.Client.FindUsers(item.req)
		if err.Error() != item.err.Error() {
			t.Errorf("[%d] invalid error, expected %, got %v", caseNum, item.err, err)
		}
		if !reflect.DeepEqual(item.res, resp) {
			t.Errorf("[%d] invalid result, expected %v, got %v", caseNum, item.res, resp)
		}
	}
}

func TestHugeLimit(t *testing.T) {
	ts := NewTestServer(allowedAccessToken)
	defer ts.Close()

	resp, _ := ts.Client.FindUsers(SearchRequest{
		Limit: 1000,
	})

	if len(resp.Users) != 25 {
		t.Errorf("invalid number of users: %d", len(resp.Users))
	}
}

// TestPagination testing pagitanion correctness.
func TestPagination(t *testing.T) {
	ts := NewTestServer(allowedAccessToken)
	defer ts.Close()

	resp, _ := ts.Client.FindUsers(SearchRequest{
		Limit:  4,
		Offset: 0,
	})

	if len(resp.Users) != 4 {
		t.Errorf("invalid nums of users: %d", len(resp.Users))
		return
	}

	if resp.Users[3].Id != 3 {
		t.Errorf("invalid user at row 3: %v", resp.Users[3])
		return
	}

	resp, _ = ts.Client.FindUsers(SearchRequest{
		Limit:  6,
		Offset: 3,
	})

	if len(resp.Users) != 6 {
		t.Errorf("invalid number of users: %d", len(resp.Users))
		return
	}

	if resp.Users[0].Name != "Everett Dillard" {
		t.Errorf("invalid user at row 3: %v", resp.Users[0])
		return
	}
}

// TestWrongOrderField testing non-existent field in sorting.
func TestWrongOrderField(t *testing.T) {
	ts := NewTestServer(allowedAccessToken)
	defer ts.Close()

	_, err := ts.Client.FindUsers(SearchRequest{
		OrderBy:    OrderByAsc,
		OrderField: "fake",
	})

	if err == nil {
		t.Errorf("empty error, must be %v", ErrorBadOrderField)
	} else if err.Error() != "OrderFeld fake invalid" {
		t.Errorf("invalid error: %v", err.Error())
	}
}

func TestUnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "ooops it's error!", http.StatusBadRequest)
	}))
	client := SearchClient{allowedAccessToken, server.URL}
	defer server.Close()

	_, err := client.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("empty error")
	} else if !strings.Contains(err.Error(), "cant unpack error json") {
		t.Errorf("invalid error: %v", err.Error())
	}
}

func TestUnmarshalResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ooops strange result!")
	}))
	client := SearchClient{allowedAccessToken, server.URL}
	defer server.Close()

	_, err := client.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("empty error")
	} else if !strings.Contains(err.Error(), "cant unpack result json") {
		t.Errorf("invalid error: %v", err.Error())
	}
}

func TestUnknownError(t *testing.T) {
	client := SearchClient{allowedAccessToken, "ooops unknown endpoint!"}

	_, err := client.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("empty error")
	} else if !strings.Contains(err.Error(), "unknown error") {
		t.Errorf("invalid error: %v", err.Error())
	}
}

func TestFatalCrash(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pushError(w, errSource, http.StatusInternalServerError)
	}))
	client := SearchClient{allowedAccessToken, server.URL}
	defer server.Close()

	_, err := client.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("empty error")
	} else if err.Error() != "SearchServer fatal error" {
		t.Errorf("invalid error: %v", err.Error())
	}
}
