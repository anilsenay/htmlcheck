package htmlcheck

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

var v Validator = Validator{}

func TestMain(m *testing.M) {
	v.AddGroup(&TagGroup{
		Name: "example",
		// Attrs: []string{"test1", "test2"},
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
		Attrs: []Attribute{{Name: "src", Value: &AttributeValue{StartsWith: "http"}}},
	})
	os.Exit(m.Run())
}

func Test_SingleTag(t *testing.T) {
	errors := v.ValidateHtmlString("<a></a>")
	assert.Len(t, errors, 0)
}

func Test_SelfClosingTag(t *testing.T) {
	errors := v.ValidateHtmlString("<a>")
	assert.Len(t, errors, 0)
}

func Test_SingleAttr(t *testing.T) {
	errors := v.ValidateHtmlString("<a href='test'>")
	assert.Len(t, errors, 0)
}

func Test_UnknownAttr(t *testing.T) {
	errors := v.ValidateHtmlString("<a hrefff='test'>")
	assert.NotEmpty(t, errors)
}

func Test_DuplicatedAttr(t *testing.T) {
	errors := v.ValidateHtmlString("<a href='test' href='test2'>")
	assert.NotEmpty(t, errors)
}

func Test_SingleUnknownTag(t *testing.T) {
	errors := v.ValidateHtmlString("<art>")
	assert.NotEmpty(t, errors)
}

func Test_AttributeValue(t *testing.T) {
	errors := v.ValidateHtmlString("<img src='http://test.com'></img>")
	assert.Len(t, errors, 0)
}

func Test_InvalidAttributeValue(t *testing.T) {
	errors := v.ValidateHtmlString("<img src='ftp://test.com'></img>")
	assert.NotEmpty(t, errors)
}

func Test_UnclosedTag(t *testing.T) {
	errors := v.ValidateHtmlString("<b>df")
	assert.NotEmpty(t, errors)
}

func Test_NestedTags(t *testing.T) {
	errors := v.ValidateHtmlString("<b><a></a></b>")
	assert.Len(t, errors, 0)
}

func Test_Groups(t *testing.T) {
	errors := v.ValidateHtmlString("<a test1=4 test2=5></a>")
	assert.Len(t, errors, 0)
}

func Test_WronglyNestedTags(t *testing.T) {
	errors := v.ValidateHtmlString("<b><c></b></c>")
	assert.NotEmpty(t, errors)
}

func Test_SwapedStartClosingTags(t *testing.T) {
	errors := v.ValidateHtmlString("</b><b>")
	assert.NotEmpty(t, errors)
	assert.ErrorAs(t, errors[0], &ErrInvClosedBeforeOpened{})
}

func Test_NextedTagsWithSelfClosing(t *testing.T) {
	errors := v.ValidateHtmlString("<b><a></b>")
	assert.Len(t, errors, 0)
}

func Test_AttributesWithoutValue(t *testing.T) {
	errors := v.ValidateHtmlString("<a test1 test2></a>")
	assert.Len(t, errors, 0)
}

func Test_NextedTagsWithUnkonwAttribute1(t *testing.T) {
	errors := v.ValidateHtmlString("<b kkk='kkk'><a></b>")
	if len(errors) != 1 {
		t.Fatal("should raise invalid attribute error")
	}
}

func Test_NextedTagsWithUnkonwAttribute2(t *testing.T) {
	errors := v.ValidateHtmlString("<b><a kkk='kkk'></b>")
	if len(errors) != 1 {
		t.Fatal("should raise invalid attribute error")
	}
}

func Test_AttrStartsWith(t *testing.T) {
	errors := v.ValidateHtmlString("<style data-jiis='cc' id='gstyle'></style>")
	assert.Len(t, errors, 0)
}

func Test_Callback(t *testing.T) {
	triggerd := false
	v.RegisterCallback(func(tagName string, attributeName string, value string, reason ErrorReason) error {
		triggerd = true
		return nil
	})

	errors := v.ValidateHtmlString("<kk>")
	if !triggerd {
		t.Fatal("should trigger callback")
	}

	assert.Len(t, errors, 0)
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
