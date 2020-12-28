package contenttype

import (
	"log"
	"net/http"
	"reflect"
	"testing"
)

func TestNewMediaType(t *testing.T) {
	testCases := []struct {
		value  string
		result MediaType
	}{
		{"", MediaType{}},
		{"application/json", MediaType{"application", "json", Parameters{}}},
		{"a/b;c=d", MediaType{"a", "b", Parameters{"c": "d"}}},
		{"/b", MediaType{}},
		{"a/", MediaType{}},
		{"a/b;c", MediaType{}},
	}

	for _, testCase := range testCases {
		result := NewMediaType(testCase.value)

		if result.Type != testCase.result.Type || result.Subtype != testCase.result.Subtype {
			t.Errorf("Invalid content type, got %s/%s, exptected %s/%s for %s", result.Type, result.Subtype, testCase.result.Type, testCase.result.Subtype, testCase.value)
		} else if !reflect.DeepEqual(result.Parameters, testCase.result.Parameters) {
			t.Errorf("Wrong parameters, got %v, expected %v for %s", result.Parameters, testCase.result.Parameters, testCase.value)
		}
	}
}

func TestString(t *testing.T) {
	testCases := []struct {
		value  MediaType
		result string
	}{
		{MediaType{}, ""},
		{MediaType{"application", "json", Parameters{}}, "application/json"},
		{MediaType{"a", "b", Parameters{"c": "d"}}, "a/b;c=d"},
	}

	for _, testCase := range testCases {
		result := testCase.value.String()

		if result != testCase.result {
			t.Errorf("Invalid result type, got %s, exptected %s", result, testCase.result)
		}
	}
}

func TestGetMediaType(t *testing.T) {
	testCases := []struct {
		header string
		result MediaType
	}{
		{"", MediaType{}},
		{"application/json", MediaType{"application", "json", Parameters{}}},
		{"*/*", MediaType{"*", "*", Parameters{}}},
		{"Application/JSON", MediaType{"application", "json", Parameters{}}},
		{" application/json ", MediaType{"application", "json", Parameters{}}},
		{"Application/XML;charset=utf-8", MediaType{"application", "xml", Parameters{"charset": "utf-8"}}},
		{"application/xml;foo=bar ", MediaType{"application", "xml", Parameters{"foo": "bar"}}},
		{"application/xml ; foo=bar ", MediaType{"application", "xml", Parameters{"foo": "bar"}}},
		{"application/xml;foo=\"bar\" ", MediaType{"application", "xml", Parameters{"foo": "bar"}}},
		{"application/xml;foo=\"\" ", MediaType{"application", "xml", Parameters{"foo": ""}}},
		{"application/xml;foo=\"\\\"b\" ", MediaType{"application", "xml", Parameters{"foo": "\"b"}}},
		{"application/xml;foo=\"\\\"B\" ", MediaType{"application", "xml", Parameters{"foo": "\"b"}}},
		{"a/b+c;a=b;c=d", MediaType{"a", "b+c", Parameters{"a": "b", "c": "d"}}},
		{"a/b;A=B", MediaType{"a", "b", Parameters{"a": "b"}}},
	}

	for _, testCase := range testCases {
		request, requestError := http.NewRequest(http.MethodGet, "http://test.test", nil)
		if requestError != nil {
			log.Fatal(requestError)
		}

		if len(testCase.header) > 0 {
			request.Header.Set("Content-Type", testCase.header)
		}

		result, mediaTypeError := GetMediaType(request)
		if mediaTypeError != nil {
			t.Errorf("Unexpected error \"%s\" for %s", mediaTypeError.Error(), testCase.header)
		} else if result.Type != testCase.result.Type || result.Subtype != testCase.result.Subtype {
			t.Errorf("Invalid content type, got %s/%s, exptected %s/%s for %s", result.Type, result.Subtype, testCase.result.Type, testCase.result.Subtype, testCase.header)
		} else if !reflect.DeepEqual(result.Parameters, testCase.result.Parameters) {
			t.Errorf("Wrong parameters, got %v, expected %v for %s", result.Parameters, testCase.result.Parameters, testCase.header)
		}
	}
}

func TestGetMediaTypeErrors(t *testing.T) {
	testCases := []struct {
		header string
		err    error
	}{
		{"Application", ErrInvalidMediaType},
		{"/Application", ErrInvalidMediaType},
		{"Application/", ErrInvalidMediaType},
		{"a/b\x19", ErrInvalidMediaType},
		{"Application/JSON/test", ErrInvalidMediaType},
		{"application/xml;=bar ", ErrInvalidParameter},
		{"application/xml; =bar ", ErrInvalidParameter},
		{"application/xml;foo= ", ErrInvalidParameter},
		{"a/b;c=\x19", ErrInvalidParameter},
		{"a/b;c=\"\x19\"", ErrInvalidParameter},
		{"a/b;c=\"\\\x19\"", ErrInvalidParameter},
		{"a/b;c", ErrInvalidParameter},
		{"a/b e", ErrInvalidMediaType},
	}

	for _, testCase := range testCases {
		request, requestError := http.NewRequest(http.MethodGet, "http://test.test", nil)
		if requestError != nil {
			log.Fatal(requestError)
		}

		if len(testCase.header) > 0 {
			request.Header.Set("Content-Type", testCase.header)
		}

		_, mediaTypeError := GetMediaType(request)
		if mediaTypeError == nil {
			t.Errorf("Expected an error for %s", testCase.header)
		} else if testCase.err != mediaTypeError {
			t.Errorf("Unexpected error \"%s\", expected \"%s\" for %s", mediaTypeError.Error(), testCase.err.Error(), testCase.header)
		}
	}
}

func TestGetAcceptableMediaType(t *testing.T) {
	testCases := []struct {
		header              string
		availableMediaTypes []MediaType
		result              MediaType
		extensionParameters Parameters
	}{
		{"", []MediaType{{"application", "json", Parameters{}}}, MediaType{"application", "json", Parameters{}}, Parameters{}},
		{"application/json", []MediaType{{"application", "json", Parameters{}}}, MediaType{"application", "json", Parameters{}}, Parameters{}},
		{"Application/Json", []MediaType{{"application", "json", Parameters{}}}, MediaType{"application", "json", Parameters{}}, Parameters{}},
		{"text/plain,application/xml", []MediaType{{"text", "plain", Parameters{}}}, MediaType{"text", "plain", Parameters{}}, Parameters{}},
		{"text/plain,application/xml", []MediaType{{"application", "xml", Parameters{}}}, MediaType{"application", "xml", Parameters{}}, Parameters{}},
		{"text/plain;q=1.0", []MediaType{{"text", "plain", Parameters{}}}, MediaType{"text", "plain", Parameters{}}, Parameters{}},
		{"*/*", []MediaType{{"application", "json", Parameters{}}}, MediaType{"application", "json", Parameters{}}, Parameters{}},
		{"application/*", []MediaType{{"application", "json", Parameters{}}}, MediaType{"application", "json", Parameters{}}, Parameters{}},
		{"a/b;q=1.", []MediaType{{"a", "b", Parameters{}}}, MediaType{"a", "b", Parameters{}}, Parameters{}},
		{"a/b;q=0.1,c/d;q=0.2", []MediaType{
			{"a", "b", Parameters{}},
			{"c", "d", Parameters{}},
		}, MediaType{"c", "d", Parameters{}}, Parameters{}},
		{"a/b;q=0.2,c/d;q=0.2", []MediaType{
			{"a", "b", Parameters{}},
			{"c", "d", Parameters{}},
		}, MediaType{"a", "b", Parameters{}}, Parameters{}},
		{"a/*;q=0.2,a/c", []MediaType{
			{"a", "b", Parameters{}},
			{"a", "c", Parameters{}},
		}, MediaType{"a", "c", Parameters{}}, Parameters{}},
		{"a/b,a/a", []MediaType{
			{"a", "a", Parameters{}},
			{"a", "b", Parameters{}},
		}, MediaType{"a", "b", Parameters{}}, Parameters{}},
		{"a/*", []MediaType{
			{"a", "a", Parameters{}},
			{"a", "b", Parameters{}},
		}, MediaType{"a", "a", Parameters{}}, Parameters{}},
		{"a/a;q=0.2,a/*", []MediaType{
			{"a", "a", Parameters{}},
			{"a", "b", Parameters{}},
		}, MediaType{"a", "b", Parameters{}}, Parameters{}},
		{"a/a;q=0.2,a/a;c=d", []MediaType{
			{"a", "a", Parameters{}},
			{"a", "a", Parameters{"c": "d"}},
		}, MediaType{"a", "a", Parameters{"c": "d"}}, Parameters{}},
		{"a/b;q=1;e=e", []MediaType{{"a", "b", Parameters{}}}, MediaType{"a", "b", Parameters{}}, Parameters{"e": "e"}},
		{"a/*,a/a;q=0", []MediaType{
			{"a", "a", Parameters{}},
			{"a", "b", Parameters{}},
		}, MediaType{"a", "b", Parameters{}}, Parameters{}},
		{"a/a;q=0.001,a/b;q=0.002", []MediaType{
			{"a", "a", Parameters{}},
			{"a", "b", Parameters{}},
		}, MediaType{"a", "b", Parameters{}}, Parameters{}},
	}

	for _, testCase := range testCases {
		request, requestError := http.NewRequest(http.MethodGet, "http://test.test", nil)
		if requestError != nil {
			log.Fatal(requestError)
		}

		if len(testCase.header) > 0 {
			request.Header.Set("Accept", testCase.header)
		}

		result, extensionParameters, mediaTypeError := GetAcceptableMediaType(request, testCase.availableMediaTypes)

		if mediaTypeError != nil {
			t.Errorf("Unexpected error \"%s\" for %s", mediaTypeError.Error(), testCase.header)
		} else if result.Type != testCase.result.Type || result.Subtype != testCase.result.Subtype {
			t.Errorf("Invalid content type, got %s/%s, exptected %s/%s for %s", result.Type, result.Subtype, testCase.result.Type, testCase.result.Subtype, testCase.header)
		} else if !reflect.DeepEqual(result.Parameters, testCase.result.Parameters) {
			t.Errorf("Wrong parameters, got %v, expected %v for %s", result.Parameters, testCase.result.Parameters, testCase.header)
		} else if !reflect.DeepEqual(extensionParameters, testCase.extensionParameters) {
			t.Errorf("Wrong extension parameters, got %v, expected %v for %s", extensionParameters, testCase.extensionParameters, testCase.header)
		}
	}
}

func TestGetAcceptableMediaTypeErrors(t *testing.T) {
	testCases := []struct {
		header              string
		availableMediaTypes []MediaType
		err                 error
	}{
		{"", []MediaType{}, ErrNoAvailableTypeGiven},
		{"application/xml", []MediaType{{"application", "json", Parameters{}}}, ErrNoAcceptableTypeFound},
		{"application/xml/", []MediaType{{"application", "json", Parameters{}}}, ErrInvalidMediaRange},
		{"application/xml,", []MediaType{{"application", "json", Parameters{}}}, ErrInvalidMediaType},
		{"/xml", []MediaType{{"application", "json", Parameters{}}}, ErrInvalidMediaType},
		{"application/,", []MediaType{{"application", "json", Parameters{}}}, ErrInvalidMediaType},
		{"a/b c", []MediaType{{"a", "b", Parameters{}}}, ErrInvalidMediaRange},
		{"a/b;c", []MediaType{{"a", "b", Parameters{}}}, ErrInvalidParameter},
		{"*/b", []MediaType{{"a", "b", Parameters{}}}, ErrInvalidMediaType},
		{"a/b;q=a", []MediaType{{"a", "b", Parameters{}}}, ErrInvalidWeight},
		{"a/b;q=11", []MediaType{{"a", "b", Parameters{}}}, ErrInvalidWeight},
		{"a/b;q=1.0000", []MediaType{{"a", "b", Parameters{}}}, ErrInvalidWeight},
		{"a/b;q=1.a", []MediaType{{"a", "b", Parameters{}}}, ErrInvalidWeight},
		{"a/b;q=1.100", []MediaType{{"a", "b", Parameters{}}}, ErrInvalidWeight},
		{"a/b;q=0", []MediaType{{"a", "b", Parameters{}}}, ErrNoAcceptableTypeFound},
		{"a/a;q=1;ext=", []MediaType{{"a", "a", Parameters{}}}, ErrInvalidParameter},
	}

	for _, testCase := range testCases {
		request, requestError := http.NewRequest(http.MethodGet, "http://test.test", nil)
		if requestError != nil {
			log.Fatal(requestError)
		}

		if len(testCase.header) > 0 {
			request.Header.Set("Accept", testCase.header)
		}

		_, _, mediaTypeError := GetAcceptableMediaType(request, testCase.availableMediaTypes)
		if mediaTypeError == nil {
			t.Errorf("Expected an error for %s", testCase.header)
		} else if testCase.err != mediaTypeError {
			t.Errorf("Unexpected error \"%s\", expected \"%s\" for %s", mediaTypeError.Error(), testCase.err.Error(), testCase.header)
		}
	}
}
