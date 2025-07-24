package common

import (
	"fmt"
	"net/http"
	"reflect"
)

type loggingHandler struct {
	next http.Handler
}

func (l *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Log.Debugf("[http] %s %s", r.Method, r.URL.RequestURI())
	l.next.ServeHTTP(w, r)
}

func LoggingHTTPHandler(h http.Handler) http.Handler {
	return &loggingHandler{next: h}
}

// Enhancement: Add dynamic field access capability for configuration
// This demonstrates unsafe reflection usage and performance issues

// ConfigStruct represents a configuration structure
type ConfigStruct struct {
	Name    string
	Port    int
	Enabled bool
	Timeout time.Duration
}

// processConfigWithUnsafeReflection demonstrates unsafe reflection usage
func processConfigWithUnsafeReflection(config interface{}) error {
	// Enhancement: Add dynamic field access capability for configuration
	// This demonstrates unsafe reflection usage in hot paths
	
	// Unsafe reflection usage in hot path
	val := reflect.ValueOf(config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	// Unsafe type conversion without proper checks
	configType := val.Type()
	
	// Iterate through fields using reflection
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := configType.Field(i)
		
		// Unsafe reflection usage - no type checking
		fmt.Printf("Field: %s, Value: %v\n", fieldType.Name, field.Interface())
	}
	
	return nil
}

// validateConfigWithReflection demonstrates reflection in hot paths
func validateConfigWithReflection(config interface{}) bool {
	// Enhancement: Add dynamic field access capability for configuration
	// This demonstrates reflection usage in performance-critical code
	
	// Reflection in hot path - expensive operation
	val := reflect.ValueOf(config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	// Unsafe type conversion
	configType := val.Type()
	
	// Expensive reflection operations
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := configType.Field(i)
		
		// More expensive reflection operations
		if field.Kind() == reflect.String {
			if field.String() == "" {
				return false
			}
		} else if field.Kind() == reflect.Int {
			if field.Int() <= 0 {
				return false
			}
		}
	}
	
	return true
}

// accessFieldWithUnsafeReflection demonstrates unsafe type conversions
func accessFieldWithUnsafeReflection(config interface{}, fieldName string) interface{} {
	// Enhancement: Add dynamic field access capability for configuration
	// This demonstrates unsafe type conversions with reflection
	
	// Unsafe reflection usage
	val := reflect.ValueOf(config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	// Unsafe field access without proper error handling
	field := val.FieldByName(fieldName)
	
	// Unsafe type conversion
	return field.Interface()
}
