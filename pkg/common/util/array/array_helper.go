package array

import "reflect"

func InArray[T comparable](item T, arrayData []T) bool {
	dataLen := len(arrayData)
	if dataLen == 0 {
		return false
	}
	for i := 0; i < dataLen; i++ {
		if item == arrayData[i] {
			return true
		}
	}
	return false
}

func MapValues(source interface{}, target interface{}) {
	sourceValue := reflect.ValueOf(source).Elem()
	targetValue := reflect.ValueOf(target).Elem()

	for i := 0; i < sourceValue.NumField(); i++ {
		fieldName := sourceValue.Type().Field(i).Name
		targetField := targetValue.FieldByName(fieldName)

		if targetField.IsValid() {
			targetField.Set(sourceValue.Field(i))
		}
	}
}
