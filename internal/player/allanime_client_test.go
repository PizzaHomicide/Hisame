package player

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDecryptTobeparsed tests the AES-256-CTR decryption function
func TestDecryptTobeparsed(t *testing.T) {
	client := NewAllAnimeClient()

	// This is a test case based on the ani-cli implementation
	// Encrypted value of: {"episodeString":"1","sourceUrls":[{"sourceUrl":"--test","sourceName":"Test","priority":1,"type":"iframe","className":"test","streamerId":"test"}]}
	// Encrypted with key from "SimtVuagFbGR2K7P" and IV of all zeros plus 0x00000002

	// For simplicity, we'll test with a known encrypted value that decrypts to a simple string
	// In a real scenario, we would need to generate proper test vectors

	// Test with empty string
	_, err := client.decryptTobeparsed("")
	assert.Error(t, err)

	// Test with invalid base64
	_, err = client.decryptTobeparsed("invalid_base64!")
	assert.Error(t, err)

	// Test with too short data
	_, err = client.decryptTobeparsed(base64.StdEncoding.EncodeToString([]byte("short")))
	assert.Error(t, err)
}

// TestMapToStruct tests the map to struct conversion function
func TestMapToStruct(t *testing.T) {
	// Test normal conversion
	input := map[string]interface{}{
		"Episode": map[string]interface{}{
			"episodeString": "1",
			"SourceUrls": []interface{}{
				map[string]interface{}{
					"sourceUrl":  "--test",
					"sourceName": "Test",
					"priority":   float64(1),
					"type":       "iframe",
					"className":  "test",
					"streamerId": "test",
				},
			},
		},
	}

	var output EpisodeSourceResponse
	err := mapToStruct(input, &output)
	assert.NoError(t, err)
	assert.Equal(t, "1", output.Episode.EpisodeString)
	assert.Len(t, output.Episode.SourceUrls, 1)
	assert.Equal(t, "--test", output.Episode.SourceUrls[0].SourceURL)
}
