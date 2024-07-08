package utilities


// This is a fast approach for concatenating strings in Go
// strings.Join is optimized for this purpose and generally faster than manual concatenation
func ConcatWords(joiner []byte, words ...[]byte) string {
    // Calculate the total length of the resulting byte slice
    totalLen := len(joiner) * (len(words) - 1) // Length contributed by joiners
    for _, word := range words {
        totalLen += len(word)
    }
    
    if totalLen <= 0 {
        return ""
    }

    // Create a byte slice with the required length
    b := make([]byte, totalLen)

    // Populate the byte slice
    currentIndex := 0
    for i, word := range words {
        copy(b[currentIndex:], word)
        currentIndex += len(word)
        if i < len(words) - 1 {
            copy(b[currentIndex:], joiner)
            currentIndex += len(joiner)
        }
    }

    return string(b)
}
