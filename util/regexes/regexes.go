package regexes

import "regexp"

var Password = regexp.MustCompile(`^[0-9A-Za-z\W_]*([0-9][A-Za-z\W_]*[a-zA-Z]|[a-zA-Z][0-9\W_]*[0-9]|[0-9\W_]*[a-zA-Z][A-Za-z\W_]*|[a-zA-Z\W_]*[0-9][0-9A-Za-z\W_]*){6,20}$`)

var Email = regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)

var Phone = regexp.MustCompile(`^1[2-9]\d{9}$`)
