package htmlcheck

import (
	"errors"
	"fmt"
)

type ValidationErrorList []error

func (el ValidationErrorList) Join() error {
	var err error
	for _, e := range el {
		err = errors.Join(err, e)
	}
	return err
}

type ValidationError interface {
	Error() string
	Details() ErrorDetails
}

type ErrorDetails struct {
	TagName        string
	AttributeName  string
	AttributeValue string
	Reason         ErrorReason
}

func (d ErrorDetails) Details() ErrorDetails {
	return d
}

type ErrInvAttribute struct{ ErrorDetails }

func (e ErrInvAttribute) Error() string {
	return fmt.Sprintf("invalid attribute '%s' in tag '%s'", e.AttributeName, e.TagName)
}

type ErrInvClosedBeforeOpened struct{ ErrorDetails }

func (e ErrInvClosedBeforeOpened) Error() string {
	return fmt.Sprintf("tag '%s' closed before opened", e.TagName)
}

type ErrInvDuplicatedAttribute struct{ ErrorDetails }

func (e ErrInvDuplicatedAttribute) Error() string {
	return fmt.Sprintf("duplicate attribute '%s' in tag '%s'", e.AttributeName, e.TagName)
}

type ErrInvTag struct{ ErrorDetails }

func (e ErrInvTag) Error() string {
	return fmt.Sprintf("invalid tag '%s'", e.TagName)
}

type ErrInvNotProperlyClosed struct{ ErrorDetails }

func (e ErrInvNotProperlyClosed) Error() string {
	return fmt.Sprintf("tag '%s' is never closed", e.TagName)
}

type ErrInvAttributeValue struct{ ErrorDetails }

func (e ErrInvAttributeValue) Error() string {
	return fmt.Sprintf("invalid attribute value '%s' in attribute '%s' in tag '%s'", e.AttributeValue, e.AttributeName, e.TagName)
}

type ErrInvEOF struct{ ErrorDetails }

func (e ErrInvEOF) Error() string {
	return fmt.Sprintln("error occurred during tokenization")
}

func isEOF(err error) bool {
	return errors.As(err, &ErrInvEOF{})
}
