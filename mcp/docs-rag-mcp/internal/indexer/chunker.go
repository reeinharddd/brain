package indexer

import (
	"strings"
	"unicode"
)

// Chunk represents a piece of a document with metadata.
type Chunk struct {
	DocumentID string
	Index      int    // Chunk sequence number
	Content    string // The actual text content
	StartLine  int    // Approximate line number in original document
}

// ChunkDocument splits a document into chunks based on its chunk strategy.
func ChunkDocument(doc *Document) []Chunk {
	if doc == nil || doc.Body == "" {
		return []Chunk{}
	}

	strategy := doc.ChunkStrategy
	if strategy == "" {
		strategy = "section"
	}

	switch strategy {
	case "section":
		return chunkBySection(doc)
	case "sentence":
		return chunkBySentence(doc)
	default:
		// Default to section chunking
		return chunkBySection(doc)
	}
}

// chunkBySection splits document by level-2 headers (## heading).
func chunkBySection(doc *Document) []Chunk {
	var chunks []Chunk
	lines := strings.Split(doc.Body, "\n")

	var currentChunk strings.Builder
	var currentLine int
	var chunkCount int

	for i, line := range lines {
		// Check if this is a section header (## heading)
		if strings.HasPrefix(line, "## ") {
			// Save previous chunk if not empty
			content := strings.TrimSpace(currentChunk.String())
			if content != "" {
				chunks = append(chunks, Chunk{
					DocumentID: doc.ID,
					Index:      chunkCount,
					Content:    content,
					StartLine:  currentLine,
				})
				chunkCount++
			}

			// Start new chunk with this header
			currentChunk.Reset()
			currentChunk.WriteString(line)
			currentLine = i
		} else {
			currentChunk.WriteString(line)
			currentChunk.WriteString("\n")
		}
	}

	// Add final chunk
	content := strings.TrimSpace(currentChunk.String())
	if content != "" {
		chunks = append(chunks, Chunk{
			DocumentID: doc.ID,
			Index:      chunkCount,
			Content:    content,
			StartLine:  currentLine,
		})
	}

	return chunks
}

// chunkBySentence splits document into sentences and groups them.
func chunkBySentence(doc *Document) []Chunk {
	var chunks []Chunk
	sentences := splitBySentence(doc.Body)

	const sentencesPerChunk = 3
	var currentChunk strings.Builder
	var chunkCount int
	var currentLine int

	for i, sentence := range sentences {
		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(sentence)

		// Create chunk every N sentences or at end
		if (i+1)%sentencesPerChunk == 0 || i == len(sentences)-1 {
			content := strings.TrimSpace(currentChunk.String())
			if content != "" {
				chunks = append(chunks, Chunk{
					DocumentID: doc.ID,
					Index:      chunkCount,
					Content:    content,
					StartLine:  currentLine,
				})
				chunkCount++
			}
			currentChunk.Reset()
		}
	}

	return chunks
}

// splitBySentence splits text into sentences.
func splitBySentence(text string) []string {
	var sentences []string
	var currentSentence strings.Builder

	runeText := []rune(text)

	for i := 0; i < len(runeText); i++ {
		r := runeText[i]
		currentSentence.WriteRune(r)

		// Check for sentence endings
		if (r == '.' || r == '!' || r == '?') && i+1 < len(runeText) {
			// Look ahead to see if next char is whitespace or end
			nextRune := runeText[i+1]
			if unicode.IsSpace(nextRune) || i+1 == len(runeText)-1 {
				sentence := strings.TrimSpace(currentSentence.String())
				if sentence != "" {
					sentences = append(sentences, sentence)
				}
				currentSentence.Reset()
				// Skip the whitespace
				for i+1 < len(runeText) && unicode.IsSpace(runeText[i+1]) {
					i++
				}
			}
		}
	}

	// Add remaining text as final sentence
	if currentSentence.Len() > 0 {
		sentence := strings.TrimSpace(currentSentence.String())
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
	}

	return sentences
}
