package utils

import (
	"database/sql"
	"fmt"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	// ErrorMustSpecifyFields is returned when attempting to update zero fields.
	ErrorMustSpecifyFields = "must specify fields"
	// ErrorFieldNotAllowed is returned when attempting to update an immutable field.
	ErrorFieldNotAllowed = "field '%v' not allowed"
	// ErrorInvalidFieldValue is returned when a field validation fails.
	ErrorInvalidFieldValue = "invalid field value '%v' for field '%v'"
	// ErrorInvalidUTF8Encoding is returned when a string is not valid UTF-8.
	ErrorInvalidUTF8Encoding = "invalid field utf8 encoding for field '%v'"
	// ErrorInvalidRegexpTag is returned when a "regexp" tag cannot be compiled.
	ErrorInvalidRegexpTag = "invalid regexp tag for field '%v'"
	// ErrorInvalidFieldTypeForRegexpTag is returned when the regexp tag is applied to a field whose type is not string.
	ErrorInvalidFieldTypeForRegexpTag = "invalid field type '%t' for regexp tag for field '%v'"
	// ErrorInvalidFieldTypeForURLTag is returned when the regexp tag is applied to a field whose type is not string.
	ErrorInvalidFieldTypeForURLTag = "invalid field type '%t' for url tag for field '%v'"
	// ErrorInvalidFieldTypeForEnumTag is returned when the regexp tag is applied to a field whose type is not string.
	ErrorInvalidFieldTypeForEnumTag = "invalid field type '%t' for enum tag for field '%v'"
	// ErrorDB is returned when there is a db error
	ErrorDB = "db"
)

const (
	// TagRegexp is the regexp validation tag.
	TagRegexp = "regexp"
	// TagURL is the url validation tag.
	TagURL = "url"
	// TagEnum is the enum validation tag.
	TagEnum = "enum"
)

// MakeUpdate generates the query update string for the given fields.
func MakeUpdate(fields ...string) string {
	entries := make([]string, len(fields))
	for i := 0; i < len(fields); i++ {
		entries[i] = fmt.Sprintf("`%v` = :%v", fields[i], fields[i])
	}
	return strings.Join(entries, ", ")
}

func flattenStruct(t reflect.Type) []reflect.StructField {
	var fields = []reflect.StructField{}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			fields = append(fields, flattenStruct(f.Type)...)
		} else {
			fields = append(fields, f)
		}
	}
	return fields
}

// CheckMutableFields verifies that the given field names have the mutable:"true" tag on the given Type.
func CheckMutableFields(t reflect.Type, fields ...string) error {
	if fields == nil || len(fields) == 0 {
		return xerror.New(ErrorMustSpecifyFields)
	}
	fieldsCopy := make([]string, len(fields))
	copy(fieldsCopy, fields)
	fStruct := flattenStruct(t)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for _, f := range fStruct {
		for j, field := range fieldsCopy {
			if f.Tag.Get("db") == field {
				if f.Tag.Get("mutable") != "true" {
					return xerror.New(ErrorFieldNotAllowed, field)
				}
				fieldsCopy = append(fieldsCopy[:j], fieldsCopy[j+1:]...)
				break
			}
		}
	}
	if len(fieldsCopy) > 0 {
		return xerror.New(ErrorFieldNotAllowed, fieldsCopy)
	}
	return nil
}

// ValidateFields validates fields values on the given Value.
func ValidateFields(v reflect.Value) error {
	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		s, ok := maybeGetStringValue(v.Field(i).Interface())
		regexpTag := f.Tag.Get(TagRegexp)
		urlTag := f.Tag.Get(TagURL)
		enumTag := f.Tag.Get(TagEnum)

		if ok {
			if s != nil {
				if !utf8.ValidString(*s) {
					return xerror.New(ErrorInvalidUTF8Encoding, f.Name, *s)
				}
				if regexpTag != "" {
					r, err := regexp.Compile(regexpTag)
					if err != nil {
						return xerror.Wrap(err, ErrorInvalidRegexpTag, f.Name, regexpTag)
					}
					if s != nil && !r.MatchString(*s) {
						return xerror.New(ErrorInvalidFieldValue, *s, f.Name)
					}
				}
				if urlTag != "" {
					if *s == "" {
						return xerror.New(ErrorInvalidFieldValue, *s, f.Name)
					}
					if _, err := url.Parse(*s); err != nil {
						return xerror.Wrap(err, ErrorInvalidFieldValue, *s, f.Name)
					}
				}
				if enumTag != "" {
					if !checkEnum(*s, strings.Split(enumTag, ",")) {
						return xerror.New(ErrorInvalidFieldValue, *s, f.Name)
					}
				}
			}
		} else {
			if regexpTag != "" {
				return xerror.New(ErrorInvalidFieldTypeForRegexpTag, f.Type, f.Name)
			}
			if urlTag != "" {
				return xerror.New(ErrorInvalidFieldTypeForURLTag, f.Type, f.Name)
			}
			if enumTag != "" {
				return xerror.New(ErrorInvalidFieldTypeForEnumTag, f.Type, f.Name)
			}
		}
	}
	return nil
}

func maybeGetStringValue(v interface{}) (*string, bool) {
	s, ok := v.(string)
	if ok {
		return &s, true
	}

	ns, ok := v.(null.String)
	if !ok {
		return nil, false
	}
	if !ns.Valid {
		return nil, true
	}
	return &ns.String, true
}

func checkEnum(v string, values []string) bool {
	for _, value := range values {
		if v == value {
			return true
		}
	}
	return false
}

// WrapTx is a helper that allows to wrap a closure call inside a DB transaction.
func WrapTx(db *sqlx.DB, f func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return xerror.Wrap(err, ErrorDB)
	}
	defer tx.Rollback()
	err = f(tx)
	if err == nil {
		err = tx.Commit()
	}
	return err
}

//IsNotFoundError checks if the err is sql.ErrNoRows
func IsNotFoundError(err error) bool {
	return err != nil && xerror.Contains(err, sql.ErrNoRows.Error())
}
