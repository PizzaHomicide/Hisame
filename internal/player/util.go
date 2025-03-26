package player

// ParseArgs splits a string of command-line arguments, respecting quotes
func ParseArgs(argsString string) []string {
	var args []string
	inQuotes := false
	current := ""

	for _, r := range argsString {
		switch r {
		case '"', '\'':
			inQuotes = !inQuotes
		case ' ':
			if !inQuotes {
				if current != "" {
					args = append(args, current)
					current = ""
				}
			} else {
				current += string(r)
			}
		default:
			current += string(r)
		}
	}

	if current != "" {
		args = append(args, current)
	}

	return args
}
