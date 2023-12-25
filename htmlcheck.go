package htmlcheck

import (
	"encoding/json"
	errorsPkg "errors"
	"io"
	"io/ioutil"
	"log"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

type ErrorReason int

const (
	InvTag                 ErrorReason = 0
	InvAttribute           ErrorReason = 1
	InvClosedBeforeOpened  ErrorReason = 2
	InvNotProperlyClosed   ErrorReason = 3
	InvDuplicatedAttribute ErrorReason = 4
	InvEOF                 ErrorReason = 5
)

type ErrorCallback func(tagName string, attributeName string, value string, reason ErrorReason) error

type TagGroup struct {
	Name  string
	Attrs []Attribute
}

type AttributeValue struct {
	List       []string
	Regex      string
	StartsWith string
}

type Attribute struct {
	Name  string
	Value *AttributeValue
}

type ValidTag struct {
	Name           string
	Attrs          []Attribute
	AttrRegex      string
	AttrStartsWith string
	Groups         []string
	IsSelfClosing  bool
}

type TagsFile struct {
	Groups []*TagGroup
	Tags   []*ValidTag
}

type Validator struct {
	validTagMap          map[string]map[string]Attribute
	validSelfClosingTags map[string]bool
	errorCallback        ErrorCallback
	StopAfterFirstError  bool
	validTags            map[string]*ValidTag
	validGroups          map[string]*TagGroup
}

func (v *Validator) AddValidTags(validTags []*ValidTag) {
	if v.validSelfClosingTags == nil {
		v.validSelfClosingTags = make(map[string]bool)
	}
	if v.validTagMap == nil {
		v.validTagMap = make(map[string]map[string]Attribute)
	}
	if v.validTags == nil {
		v.validTags = map[string]*ValidTag{}
	}

	for _, tag := range validTags {
		if tag.IsSelfClosing {
			v.validSelfClosingTags[tag.Name] = true
		}
		v.validTagMap[tag.Name] = make(map[string]Attribute, 0)
		for _, a := range tag.Attrs {
			v.validTagMap[tag.Name][a.Name] = a
		}
		if tag.Name == "" {
			_, hasGlobalTag := v.validTags[""]
			if hasGlobalTag {
				log.Println("second global tag")
			}
		}
		v.validTags[tag.Name] = tag

		for _, groupName := range tag.Groups {
			group := v.validGroups[groupName]
			for _, attr := range group.Attrs {
				v.validTagMap[tag.Name][attr.Name] = attr
			}
		}
	}
}

func (v *Validator) AddValidTag(validTag ValidTag) {
	v.AddValidTags([]*ValidTag{&validTag})
}

func (v *Validator) AddGroup(group *TagGroup) {
	v.AddGroups([]*TagGroup{group})
}

func (v *Validator) AddGroups(groups []*TagGroup) {
	if v.validGroups == nil {
		v.validGroups = map[string]*TagGroup{}
	}
	for _, g := range groups {
		v.validGroups[g.Name] = g

		for _, t := range v.validTags {
			if t.HasGroup(g.Name) {
				for _, attr := range g.Attrs {
					v.validTagMap[t.Name][attr.Name] = attr
				}
			}
		}
	}
}

func (tag *ValidTag) HasGroup(groupName string) bool {
	for _, g := range tag.Groups {
		if g == groupName {
			return true
		}
	}
	return false
}

func (v *Validator) RegisterCallback(f ErrorCallback) {
	v.errorCallback = f
}

func (v *Validator) IsValidTag(tagName string) bool {
	_, ok := v.validTagMap[tagName]
	return ok
}

func (v *Validator) IsValidSelfClosingTag(tagName string) bool {
	_, ok := v.validSelfClosingTags[tagName]
	if !ok {
		return false
	}
	return ok
}

func (v *Validator) LoadTagsFromFile(path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	tagFile := TagsFile{}
	err = json.Unmarshal(content, &tagFile)

	if err != nil {
		return err
	}

	v.AddGroups(tagFile.Groups)
	v.AddValidTags(tagFile.Tags)

	return nil
}

func (v *Validator) validateAttribute(tagName string, attrName string, attrValue string) ValidationError {
	if attrs, hasTag := v.validTagMap[tagName]; hasTag {
		attr, hasAttr := attrs[attrName]
		if hasAttr {
			return v.validateAttributeValue(tagName, attr, attrValue)
		}

		//test regex
		if ok := v.checkAttributeRegex(tagName, attrName); ok {
			return v.validateAttributeValue(tagName, attr, attrValue)
		}
	}

	//check global attributes
	if gAttrs, hasGlobals := v.validTagMap[""]; hasGlobals {
		globalAttr, hasGlobalAttr := gAttrs[attrName]
		if hasGlobalAttr {
			return v.validateAttributeValue(tagName, globalAttr, attrValue)
		}

		//test regex
		if ok := v.checkAttributeRegex("", attrName); ok {
			return v.validateAttributeValue(tagName, globalAttr, attrValue)
		}
	}

	return ErrInvAttribute{ErrorDetails{TagName: tagName, AttributeName: attrName, AttributeValue: attrValue}}
}

func (v *Validator) validateAttributeValue(tagName string, attr Attribute, attrValue string) ValidationError {
	tagAttrValue := attr.Value
	if tagAttrValue == nil {
		return nil
	}

	if slices.Contains(tagAttrValue.List, attrValue) {
		return nil
	}
	if attr.Value.StartsWith != "" {
		if strings.HasPrefix(attrValue, tagAttrValue.StartsWith) {
			return nil
		}
	}
	if attr.Value.Regex != "" {
		matches, err := regexp.MatchString(attr.Value.Regex, attrValue)
		if err == nil && matches {
			return nil
		}
	}

	return ErrInvAttributeValue{ErrorDetails{TagName: tagName, AttributeName: attr.Name, AttributeValue: attrValue}}
}

func (v *Validator) checkAttributeRegex(tagName string, attrName string) bool {
	tag := v.validTags[tagName]
	if tag.AttrStartsWith != "" {
		return strings.HasPrefix(attrName, tag.AttrStartsWith)
	}
	if tag.AttrRegex != "" {
		matches, err := regexp.MatchString(tag.AttrRegex, attrName)
		if err == nil && matches {
			return true
		}
	}
	return false
}

func (v *Validator) ValidateHtmlString(str string) []error {
	buffer := strings.NewReader(str)
	errors := v.ValidateHtml(buffer)
	return errors
}

func (v *Validator) checkErrorCallback(err ValidationError) error {
	if err == nil {
		return nil
	}

	if v.errorCallback != nil {
		details := err.Details()
		return v.errorCallback(details.TagName, details.AttributeName, details.AttributeValue, details.Reason)
	}
	return err
}

func (v *Validator) ValidateHtml(r io.Reader) []error {
	d := html.NewTokenizer(r)

	errors := []error{}
	parents := []string{}
	var err error
	for {
		parents, err = v.checkToken(d, parents)
		if err != nil {
			if errorsPkg.As(err, &ErrInvEOF{}) {
				break
			}
			errors = append(errors, err)
			if v.StopAfterFirstError {
				return errors
			}
		}
	}

	err = v.checkParents(d, parents)
	if err != nil {
		errors = append(errors, err)
	}
	return errors
}

func indexOf(arr []string, val string) int {
	for i, k := range arr {
		if k == val {
			return i
		}
	}
	return -1
}

func (v *Validator) checkParents(d *html.Tokenizer, parents []string) error {
	for _, tagName := range parents {
		if v.IsValidSelfClosingTag(tagName) {
			continue
		}

		cError := v.checkErrorCallback(ErrInvNotProperlyClosed{ErrorDetails{TagName: tagName}})
		if cError != nil {
			return cError
		}
	}
	return nil
}

func popLast(list []string) []string {
	if len(list) == 0 {
		return list
	}
	return list[0 : len(list)-1]
}

func (v *Validator) checkToken(d *html.Tokenizer, parents []string) ([]string, error) {

	tokenType := d.Next()

	if tokenType == html.ErrorToken {
		return parents, ErrInvEOF{}
	}

	token := d.Token()

	if tokenType == html.EndTagToken ||
		tokenType == html.StartTagToken ||
		tokenType == html.SelfClosingTagToken {

		tagName := token.Data

		if !v.IsValidTag(tagName) {
			cError := v.checkErrorCallback(ErrInvTag{ErrorDetails{TagName: tagName}})
			if cError != nil {
				return parents, cError
			}
		}

		if token.Type == html.StartTagToken ||
			token.Type == html.SelfClosingTagToken {
			parents = append(parents, tagName)
		}

		attrs := map[string]bool{}

		for _, attr := range token.Attr {
			err := v.validateAttribute(tagName, attr.Key, attr.Val)
			if err != nil {
				cError := v.checkErrorCallback(err)
				if cError != nil {
					return parents, cError
				}
			}

			_, ok := attrs[attr.Key]
			if !ok {
				attrs[attr.Key] = true
			} else {
				cError := v.checkErrorCallback(ErrInvDuplicatedAttribute{ErrorDetails{TagName: tagName, AttributeName: attr.Key, AttributeValue: attr.Val}})
				if cError != nil {
					return parents, cError
				}
			}
		}

		if token.Type == html.EndTagToken {
			if len(parents) > 0 && parents[len(parents)-1] == tagName {
				parents = popLast(parents)
			} else if len(parents) == 0 ||
				parents[len(parents)-1] != tagName {
				index := indexOf(parents, tagName)
				if index > -1 {
					missingTagName := parents[len(parents)-1]
					parents = parents[0:index]
					if !v.IsValidSelfClosingTag(missingTagName) {
						cError := v.checkErrorCallback(ErrInvNotProperlyClosed{ErrorDetails{TagName: tagName}})
						if cError != nil {
							return parents, cError
						}
					}
				} else {
					cError := v.checkErrorCallback(ErrInvClosedBeforeOpened{ErrorDetails{TagName: tagName}})
					if cError != nil {
						return parents, cError
					}
				}
			}
		}
	}

	return parents, nil
}
