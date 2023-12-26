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

Documentation will be added soon

## Contributing

Anyone can contribute by opening issue or pull-request

## License

Distributed under the GPL License. See `LICENSE` for more information.
