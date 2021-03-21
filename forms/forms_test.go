package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestValid(t *testing.T) {
	req := httptest.NewRequest("POST", "/dummy-path", nil)
	testForm := New(req.PostForm)
	isValid := testForm.Valid()
	if !isValid {
		t.Error("Shows invalid when form is valid")
	}
}

func TestNew(t *testing.T) {

}

func TestHas_NoFormData(t *testing.T) {
	req := httptest.NewRequest("POST", "/dummy-path", nil)
	testForm := New(req.PostForm)
	has := testForm.Has("field1")
	if has {
		t.Error("Shows has data when it should not have")
	}
}

func TestHas_WithFormData(t *testing.T) {
	req := httptest.NewRequest("POST", "/dummy-path", nil)
	req.ParseForm()
	formData := req.PostForm
	formData.Add("field1", "value1")

	testForm := New(formData)
	has := testForm.Has("field1")
	if !has {
		t.Error("Shows valid when form is missing field1")
	}

	es := testForm.Errors.Get("field1")
	if len(es) > 0 {
		t.Error("Shows error when there should not be one")
	}
}

func TestRequired_NoFormData(t *testing.T) {
	req := httptest.NewRequest("POST", "/dummy-path", nil)
	req.ParseForm()
	testForm := New(req.PostForm)
	testForm.Required("field1", "field2", "field3")
	isValid := testForm.Valid()
	if isValid {
		t.Error("Shows valid when required fields are missing")
	}

	es := testForm.Errors.Get("field1")
	if es == "" {
		t.Error("Shows form has no error when there should be")
	}
}

func TestRequired_WithFormData(t *testing.T) {
	formData := url.Values{}
	formData.Add("field1", "value1")
	formData.Add("field1", "value2")

	testForm := New(formData)
	testForm.Required("field1")
	isValid := testForm.Valid()
	if !isValid {
		t.Error("Shows invalid when form has the required field")
	}
}

func TestMinLength_NoFormData(t *testing.T) {
	req := httptest.NewRequest("POST", "/dummy-path", nil)
	req.ParseForm()

	formData := req.PostForm
	testForm := New(formData)
	testForm.MinLength("name", 1)
	isValid := testForm.Valid()
	if isValid {
		t.Error("Shows valid field when form does not have form data")
	}
}

func TestMinLength_WithFormData(t *testing.T) {
	req := httptest.NewRequest("POST", "/dummy-path", nil)
	req.ParseForm()

	// Satisfies min length requirement
	formData := req.PostForm
	formData.Add("name", "Jack")

	testForm := New(formData)
	testForm.MinLength("name", 4)

	isValid := testForm.Valid()
	if !isValid {
		t.Error("Form field does not satisfy min length")
	}

	// Does not satisfy min length requirement
	testForm.MinLength("name", 5)
	isValid = testForm.Valid()
	if isValid {
		t.Error("Shows min length of 5 met when data is shorter")
	}
}

func TestIsEmail(t *testing.T) {
	req := httptest.NewRequest("POST", "/dummy-path", nil)
	req.ParseForm()

	// Valid email
	formData := req.PostForm
	formData.Add("email", "jackwong3101@yahoo.com")

	testForm := New(formData)
	testForm.IsEmail("email")

	isValid := testForm.Valid()
	if !isValid {
		t.Error("Sows invalid email when the email string is valid")
	}

	// Invalid email
	formData = url.Values{}
	formData.Add("email", "invalidemail")

	testForm = New(formData)
	testForm.IsEmail("email")

	isValid = testForm.Valid()
	if isValid {
		t.Error("Sows valid email for invalid email string")
	}
}
