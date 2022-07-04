package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/egorovdmi/financify/app/financify-api/handlers"
	"github.com/egorovdmi/financify/business/auth"
	"github.com/egorovdmi/financify/business/data/user"
	"github.com/egorovdmi/financify/business/tests"
	"github.com/google/go-cmp/cmp"
)

type UserTests struct {
	app        http.Handler
	kid        string
	userToken  string
	adminToken string
}

func TestUsers(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	tests := UserTests{
		app:        handlers.API("develop", shutdown, test.Log, test.Auth, test.DB),
		kid:        test.KID,
		userToken:  test.Token(test.KID, "user@example.com", "gophers"),
		adminToken: test.Token(test.KID, "admin@example.com", "gophers"),
	}

	t.Run("crudUsers", tests.crudUsers)
}

func (ut *UserTests) crudUsers(t *testing.T) {
	nu := ut.postUser201(t)
	defer ut.deleteUser204(t, nu.ID)

	ut.getUser200(t, nu.ID)
	ut.putUser204(t, nu.ID)
	ut.putUser403(t, nu.ID)
}

func (ut *UserTests) postUser201(t *testing.T) user.User {
	nu := user.NewUser{
		Name:            "John Smith",
		Email:           "smith@example.com",
		Roles:           []string{auth.RoleAdmin},
		Password:        "gophers",
		PasswordConfirm: "gophers",
	}

	body, err := json.Marshal(&nu)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Add("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	var got user.User

	t.Log("Given the need to create a new user with the users endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the declared user value.", testID)
		{
			if w.Code != http.StatusCreated {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 201 for the response.", tests.Failed, testID)
			}

			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould able to unmarshal the response : %v.", tests.Failed, testID, err)
			}

			exp := got
			exp.Name = "John Smith"
			exp.Email = "smith@example.com"
			exp.Roles = []string{auth.RoleAdmin}

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s.", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}

	}

	return got
}

func (ut *UserTests) deleteUser204(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodDelete, "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Add("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deliting a new user that doesn't exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new user %s.", testID, id)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)
		}
	}
}

func (ut *UserTests) getUser200(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodGet, "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Add("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	var got user.User

	t.Log("Given the need to retreive a user.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the provided user ID.", testID)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response. Received code: %v.", tests.Failed, testID, w.Code)
			}

			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould able to unmarshal the response : %v.", tests.Failed, testID, err)
			}

			exp := got
			exp.Name = "John Smith"
			exp.Email = "smith@example.com"
			exp.Roles = []string{auth.RoleAdmin}

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s.", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}

	}
}

func (ut *UserTests) putUser204(t *testing.T, id string) {
	body := `{ "name": "Anna Smith" }`

	r := httptest.NewRequest(http.MethodPut, "/v1/users/"+id, strings.NewReader(body))
	w := httptest.NewRecorder()

	r.Header.Add("Authorization", "Bearer "+ut.adminToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to update a new user with the users endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using modified user value.", testID)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)

			r = httptest.NewRequest(http.MethodGet, "/v1/users/"+id, nil)
			w = httptest.NewRecorder()

			r.Header.Add("Authorization", "Bearer "+ut.adminToken)
			ut.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", tests.Success, testID)

			var got user.User

			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould able to unmarshal the response : %v.", tests.Failed, testID, err)
			}

			if got.Name != "Anna Smith" {
				t.Fatalf("\t%s\tTest %d:\tShould see an updated Name : got %q want %q.", tests.Failed, testID, got.Name, "Anna Smith")
			}
			t.Logf("\t%s\tTest %d:\tShould see an updated Name.", tests.Success, testID)

			if got.Email != "smith@example.com" {
				t.Fatalf("\t%s\tTest %d:\tShould not affect other fields like Email : got %q want %q.", tests.Failed, testID, got.Email, "smith@example.com")
			}
			t.Logf("\t%s\tTest %d:\tShould not affect other fields like Email.", tests.Success, testID)
		}
	}
}

func (ut *UserTests) putUser403(t *testing.T, id string) {
	body := `{ "name": "Anna Smith" }`

	r := httptest.NewRequest(http.MethodPut, "/v1/users/"+id, strings.NewReader(body))
	w := httptest.NewRecorder()

	r.Header.Add("Authorization", "Bearer "+ut.userToken)
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to update a new user with the users endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen a non-admin user makes a request..", testID)
		{
			if w.Code != http.StatusForbidden {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 403 for the response.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 403 for the response.", tests.Success, testID)
		}
	}
}
