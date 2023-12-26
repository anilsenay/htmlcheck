# htmlcheck

simple, fast and easy HTML validator in Go

---

## About The Project

`htmlcheck` is a lightweight and efficient Golang package designed to simplify the validation of HTML content in your Go applications. Whether you're working with complete HTML documents or HTML snippets, this package provides a straightforward interface to check the validity of your markup.

You can specify valid HTML tags, their attributes, and permissible attribute values, providing a comprehensive solution for HTML validation.

This package is a clone of [htmlcheck](https://github.com/mpfund/htmlcheck) package which has not been maintained for a long time

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Examples](#examples)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## Installation

```bash
go get github.com/anilsenay/htmlcheck
```

## Usage

Explain how users can import and use your package in their Go projects. Include code snippets to demonstrate the basic usage.

```go
package main

import (
	"fmt"
	"github.com/anilsenay/htmlcheck"
)

func main() {
	validator := htmlcheck.Validator{}

	validator.AddValidTag(htmlcheck.ValidTag{
		Name:  "a",
		Attrs: []htmlcheck.Attribute{
			{Name: "id"},
			{Name: "href", Value: &htmlcheck.AttributeValue{
				// valid regex for href attribute value
				Regex: "^(http(s|))://.*"
			}},
			{Name: "target", Value: &htmlcheck.AttributeValue{
				// valid values for target attribute value
				List: []string{"_target", "_blank"}
			}},
		},
		IsSelfClosing: false,
	})

	html := "<a href='http://hello.world'>Hello, World!</a>"
	errors := htmlcheck.Validate(html)
	if len(errors) == 0 {
		fmt.Println("HTML is not valid.")
	} else {
		fmt.Println("HTML is valid!")
	}
}
```

## Examples

```go
package main

import (
	"fmt"
	"github.com/anilsenay/htmlcheck"
)

func main() {
	validator := htmlcheck.Validator{}

	validLink := htmlcheck.ValidTag{
		Name:  "a",
		Attrs: []htmlcheck.Attribute{
			{Name: "id"},
			{Name: "href", Value: &htmlcheck.AttributeValue{Regex: "^(http(s|))://.*"}}, // valid regex for href attribute value
			{Name: "target", Value: &htmlcheck.AttributeValue{List: []string{"_target", "_blank"}}}, // valid values for target attribute value
		},
		IsSelfClosing: false,
	}

	validator.AddValidTag(validLink)

	// first check
	err := validator.ValidateHtmlString("<a href='http://google.com'>m</a>").Join()
	if err == nil {
		fmt.Println("ok")
	} else {
		fmt.Println(err)
	}

	// second check
	// notice the missing / in the second <a>:
	errors := validator.ValidateHtmlString("<a href='http://google.com'>m<a>")
	if len(errors) == 0 {
		fmt.Println("ok")
	} else {
		fmt.Println(errors)
	}
}
```

output:

```
ok
tag 'a' is not properly closed
```

## Documentation

### Error Handling

Validation function returns a slice of errors (`[]error`) as type of `ValidationErrors`.

```go
errors := validator.ValidateHtmlString("<a href='http://google.com'>m<a>")
if len(errors) > 0 {
	fmt.Println("NOT valid")
}
```

You can join all errors as one by using `Join()` function:

```go
err := validator.ValidateHtmlString("<a href='http://google.com'>m<a>").Join()
if err != nil {
	fmt.Println("NOT valid")
}
```

##### Stop after first error

It will return after an error occurs

```go
validator := htmlcheck.Validator{
	StopAfterFirstError: true,
}
errors := validator.ValidateHtmlString("<a href='http://google.com'>m<a>")
if len(errors) > 0 {
	err := errors[0] // its the only error
}
```

##### Check error types

You can check type of errors:

<details>
	<summary>Example</summary>
	
```go
var err error
for _, e := range validationErrors {
  switch v := e.(type) {
  case htmlcheck.ErrInvAttribute:
    err = errors.Join(err, fmt.Errorf("inv attr: '%s'", v.AttributeName))
  case htmlcheck.ErrInvAttributeValue:
    err = errors.Join(err, fmt.Errorf("inv attr val: '%s'", v.AttributeValue))
  case htmlcheck.ErrInvClosedBeforeOpened:
    err = errors.Join(err, fmt.Errorf("closed before opened: '%s'", v.TagName))
  case htmlcheck.ErrInvDuplicatedAttribute:
    err = errors.Join(err, fmt.Errorf("dup attr: '%s'", v.AttributeName))
  case htmlcheck.ErrInvTag:
    err = errors.Join(err, fmt.Errorf("inv tag: '%s'", v.TagName))
  case htmlcheck.ErrInvNotProperlyClosed:
    err = errors.Join(err, fmt.Errorf("not properly closed: '%s'", v.TagName))
  case htmlcheck.ErrInvEOF:
    err = errors.Join(err, fmt.Errorf("EOF"))
  default:
    err = errors.Join(err, fmt.Errorf("Validation error: '%s'", e.Error()))
  }
}
```
</details>

#### Register Callback

```go
v.RegisterCallback(func(tagName string, attributeName string, value string, reason ErrorReason) error {
	if reason == InvTag || reason == InvAttribute {
		return fmt.Errorf("validation error: tag '%s', attr: %s", tagName, attributeName)
	}
	return nil
})
```

### Validator Functions

##### AddValidTag

```go
validator := htmlcheck.Validator{}
validator.AddValidTag(ValidTag{
	Name:          "b",
	IsSelfClosing: false,
})
```

##### AddValidTags

```go
validator := htmlcheck.Validator{}
validator.AddValidTags([]*htmlcheck.ValidTag{
	{ Name: "div" },
	{ Name: "p" },
})
```

##### AddGroup / AddGroups

You can group attributes to use in valid tags by group name

```go
validator := htmlcheck.Validator{}
// consider it should only accept http/https urls in some attributes in this example
httpRegex := &htmlcheck.AttributeValue{Regex: "^(http(s|))://.*"}
validator.AddGroup(&htmlcheck.TagGroup{
	Name:  "valid-links",
	Attrs: []htmlcheck.Attribute{
		{Name: "href", Value: httpRegex},
		{Name: "src", Value: httpRegex},
	},
})
validator.AddValidTag(htmlcheck.ValidTag{ Name: "a", Groups: []string{"valid-links"} })
validator.AddValidTag(htmlcheck.ValidTag{ Name: "img", Groups: []string{"valid-links"} })
```

### Types

#### ValidTag

| Field            | Type          | Description                                                      |
| ---------------- | ------------- | ---------------------------------------------------------------- |
| `Name`           | `string`      | Name of tag such as `div`, `a`, `p`, `span`, etc.                |
| `Attrs`          | `[]Attribute` | Valid Attribute list for the tag                                 |
| `AttrRegex`      | `string`      | Attributes that match the regex are valid                        |
| `AttrStartsWith` | `string`      | Attributes that starts with the given input are valid            |
| `Groups`         | `[]string`    | Group list                                                       |
| `IsSelfClosing`  | `bool`        | If true, tag will be valid without closing tag, default: `false` |

<details>
	<summary>Example</summary>
	
```go
validator.AddValidTags([]*htmlcheck.ValidTag{
  { Name: "div", Attrs: []htmlcheck.Attribute{ {Name: "id"} } },
  { Name: "p", AttrStartsWith: "data-" },
  { Name: "a", AttrRegex: "^(data-).+" },
})
```
</details>

#### Attribute

| Field   | Type              | Description                                              |
| ------- | ----------------- | -------------------------------------------------------- |
| `Name`  | `string`          | Name of attribute such as `href`, `class`, `style`, etc. |
| `Value` | `*AttributeValue` | Valid values for the attribute                           |

<details>
	<summary>Example</summary>
	
```go
validLink := htmlcheck.ValidTag{
  Name:  "a",
  Attrs: []htmlcheck.Attribute{
    {Name: "id"},
    {Name: "href", Value: &htmlcheck.AttributeValue{Regex: "^(http(s|))://.*"}}, // valid regex for href attribute value
    {Name: "target", Value: &htmlcheck.AttributeValue{List: []string{"_target", "_blank"}}}, // valid values for target attribute value
  },
}
```
</details>

#### AttributeValue

| Field        | Type       | Description                                                    |
| ------------ | ---------- | -------------------------------------------------------------- |
| `List`       | `[]string` | List of valid attribute values (for example valid class names) |
| `Regex`      | `string`   | Attribute values that match the regex are valid                |
| `StartsWith` | `string`   | Attributes that starts with the given input are valid          |

<details>
	<summary>Example</summary>
	
```go
validLink := htmlcheck.ValidTag{
  Name:  "a",
  Attrs: []htmlcheck.Attribute{
    {Name: "id"},
    {Name: "href", Value: &htmlcheck.AttributeValue{Regex: "^(http(s|))://.*"}}, // valid regex for href attribute value
    {Name: "target", Value: &htmlcheck.AttributeValue{List: []string{"_target", "_blank"}}}, // valid values for target attribute value
  },
}
```
</details>

#### TagGroup

| Field   | Type          | Description                        |
| ------- | ------------- | ---------------------------------- |
| `Name`  | `string`      | Name of group                      |
| `Attrs` | `[]Attribute` | Valid Attribute list for the group |

<details>
	<summary>Example</summary>
	
```go
// consider it should only accept http/https urls in some attributes in this example
httpRegex := &htmlcheck.AttributeValue{Regex: "^(http(s|))://.*"}
validator.AddGroup(&htmlcheck.TagGroup{
  Name:  "valid-links",
  Attrs: []htmlcheck.Attribute{
    {Name: "href", Value: httpRegex}, 
    {Name: "src", Value: httpRegex},
  },
})
```
</details>

### Error Types

| Type                        | Description                                                                           |
| --------------------------- | ------------------------------------------------------------------------------------- |
| `ErrInvTag`                 | Tag is not valid                                                                      |
| `ErrInvClosedBeforeOpened`  | Tag closed before opened e.g: `<div></p></div>`                                       |
| `ErrInvNotProperlyClosed`   | Tag is opened but not closed e.g: `<div><p></div>`                                    |
| `ErrInvAttribute`           | An attribute in tag is not valid                                                      |
| `ErrInvAttributeValue`      | Value of the attribute is not valid                                                   |
| `ErrInvDuplicatedAttribute` | Duplicate attribute e.g: `<a href='..' href='..'></a>`                                |
| `ErrInvEOF`                 | This error occurs when parsing is done. It will not be added in the output error list |

## Contributing

Anyone can contribute by opening issue or pull-request

## License

Distributed under the GPL License. See `LICENSE` for more information.
