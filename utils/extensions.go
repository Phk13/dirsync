package utils

/* ExceptList stores all file extensions to be skipped when synchronizing */
type ExceptList []string

func (i *ExceptList) String() string {
	return "exceptions"
}

func (i *ExceptList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func HasExtension(s []string, str string) bool {
	for _, v := range s {
		if len(str) > len(v) {
			if v == str[len(str)-len(v):] {
				return true
			}
		}
	}
	return false
}