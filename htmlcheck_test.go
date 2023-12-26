package htmlcheck

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

var v Validator = Validator{}

func TestMain(m *testing.M) {
	v.AddGroup(&TagGroup{
		Name:  "example",
		Attrs: []Attribute{{Name: "test1"}, {Name: "test2"}},
	})
	v.AddValidTag(ValidTag{
		Name:           "", //global tag
		Attrs:          []Attribute{{Name: "id"}},
		AttrStartsWith: "data-",
		IsSelfClosing:  true,
	})
	v.AddValidTag(ValidTag{
		Name:          "a",
		Attrs:         []Attribute{{Name: "href"}},
		Groups:        []string{"example"},
		IsSelfClosing: true,
	})
	v.AddValidTag(ValidTag{
		Name:          "b",
		Attrs:         []Attribute{{Name: "id"}},
		IsSelfClosing: false,
	})
	v.AddValidTag(ValidTag{
		Name:          "c",
		Attrs:         []Attribute{{Name: "id"}},
		IsSelfClosing: false,
	})
	v.AddValidTag(ValidTag{
		Name:          "style",
		Attrs:         []Attribute{{Name: "id"}},
		IsSelfClosing: false,
	})
	v.AddValidTag(ValidTag{
		Name:  "img",
		Attrs: []Attribute{{Name: "src", Value: &AttributeValue{Regex: "^(http(s|))://.*"}}},
	})
	v.AddValidTag(ValidTag{
		Name:      "testregex",
		AttrRegex: "^(test).*",
	})
	v.AddValidTag(ValidTag{
		Name:           "startswith",
		AttrStartsWith: "test-",
	})
	v.AddValidTag(ValidTag{
		Name:  "q",
		Attrs: []Attribute{{Name: "cite", Value: &AttributeValue{StartsWith: "http"}}},
	})
	v.AddValidTag(ValidTag{
		Name:  "span",
		Attrs: []Attribute{{Name: "class", Value: &AttributeValue{List: []string{"text-sm", "text-md", "text-lg"}}}},
	})
	os.Exit(m.Run())
}

func Test_ValidateHTMLString(t *testing.T) {
	testCases := []struct {
		desc           string
		rawHTML        string
		isValid        bool
		expectedErrors []interface{}
	}{
		{"Single Tag", "<a></a>", true, nil},
		{"Self Closing Tag", "<a>", true, nil},
		{"Nested Tags", "<b><a></a></b>", true, nil},
		{"Nested Tags + Self Closing", "<b><a></b>", true, nil},
		{"Single Attribute", "<a href='test'>", true, nil},
		{"Global Attribute", "<a data-test='test'>", true, nil}, // `data-` is added as global attribute
		{"Single Attribute - Starts With (From Global)", "<style data-jiis='cc' id='gstyle'></style>", true, nil},
		{"Single Attribute - Starts With", "<startswith test-data='test'></startswith>", true, nil},
		{"Single Attribute - Regex", "<testregex test-attribute='test'></testregex>", true, nil},
		{"Attributes Without Value", "<a test1 test2></a>", true, nil},
		{"Attribute Value - List", "<span class='text-sm'></span>", true, nil},
		{"Attribute Value - StartsWith", "<q cite='http://test.com'></q>", true, nil},
		{"Attribute Value - Regex", "<img src='http://test.com'></img>", true, nil},
		{"Attribute Value - Regex", "<img src='https://test.com'></img>", true, nil},
		{"Groups", "<a test1=4 test2=5></a>", true, nil},

		{"Unclosed Tag", "<b>text", false, []interface{}{&ErrInvNotProperlyClosed{}}},
		{"Closed Before Opened", "</b><b>", false, []interface{}{&ErrInvClosedBeforeOpened{}}},
		{"Unknown Tag", "<asd></asd>", false, []interface{}{&ErrInvTag{}}},
		{"Unknown Tag - Single", "<asd>", false, []interface{}{&ErrInvTag{}}},
		{"Wrongly Nested Tags", "<b><c></b></c>", false, nil},
		{"Unknown Attribute", "<a hrefff='test'>", false, []interface{}{&ErrInvAttribute{}}},
		{"Unknown Attribute - Regex", "<testregex attr-test='test'></testregex>", false, []interface{}{&ErrInvAttribute{}}},
		{"Unknown Attribute - StartsWith", "<startswith attr-test='test'></startswith>", false, []interface{}{&ErrInvAttribute{}}},
		{"Unknown Attribute + Nested", "<b kkk='kkk'><a></a></b>", false, []interface{}{&ErrInvAttribute{}}},
		{"Unknown Attribute Value - List", "<span class='someclass'></span>", false, []interface{}{&ErrInvAttributeValue{}}},
		{"Invalid Attribute Value - Regex", "<img src='ftp://test.com'></img>", false, []interface{}{&ErrInvAttributeValue{}}},
		{"Duplicate Attribute", "<a href='test' href='test2'>", false, []interface{}{&ErrInvDuplicatedAttribute{}}},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			_errors := v.ValidateHtmlString(tC.rawHTML)
			if !tC.isValid {
				assert.NotEmpty(t, _errors)
			} else {
				assert.Len(t, _errors, 0)
			}

			if len(tC.expectedErrors) > 0 {
				for _, expectedError := range tC.expectedErrors {
					found := false
					for _, err := range _errors {
						if errors.As(err, expectedError) {
							found = true
						}
					}
					assert.True(t, found)
				}
			}
		})
	}
}

func Test_Errors_Join(t *testing.T) {
	err := v.ValidateHtmlString("<kkk></a>").Join()
	assert.Error(t, err)
	assert.ErrorAs(t, err, &ErrInvTag{})
	assert.ErrorAs(t, err, &ErrInvClosedBeforeOpened{})
}

func Test_Callback(t *testing.T) {
	generalError := errors.New("validation error")
	triggerd := false
	v.RegisterCallback(func(tagName string, attributeName string, value string, reason ErrorReason) error {
		triggerd = true
		return generalError
	})

	errors := v.ValidateHtmlString("<kkk>")
	assert.True(t, triggerd)
	assert.ErrorIs(t, errors[0], generalError)
}

func Test_Callback_DisableErrors(t *testing.T) {
	v.RegisterCallback(func(tagName string, attributeName string, value string, reason ErrorReason) error {
		if reason == InvTag || reason == InvAttribute {
			return fmt.Errorf("validation error: tag '%s', attr: %s", tagName, attributeName)
		}
		return nil
	})

	errors := v.ValidateHtmlString("<kkk>")
	assert.NotEmpty(t, errors)
	errors = v.ValidateHtmlString("<a>")
	assert.Empty(t, errors)
}

func BenchmarkValidateHtmlString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v.ValidateHtmlString("<b></b>\n<b></b>\n<b kkk='kkk'></b>")
	}
}

func BenchmarkPlainTokenizerString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		str := "<b></b>\n<b></b>\n<b kkk='kkk'></b>"
		d := html.NewTokenizer(strings.NewReader(str))
		for {
			d.Token()
			t := d.Next()
			if t == html.ErrorToken {
				break
			}

		}
	}
}
