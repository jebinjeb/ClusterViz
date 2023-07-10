package utils

func StringPtr(data string) *string {
	return &data
}

func ReturnEmptyOnNil(data *string) string {
	if data == nil {
		return ""
	}
	return *data

}
