package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInArray(t *testing.T) {
	tests := []struct {
		name     string
		needle   string
		haystack []string
		expected bool
	}{
		{
			name:     "found in array",
			needle:   ".yaml",
			haystack: []string{".yml", ".yaml", ".json"},
			expected: true,
		},
		{
			name:     "not found in array",
			needle:   ".txt",
			haystack: []string{".yml", ".yaml", ".json"},
			expected: false,
		},
		{
			name:     "empty array",
			needle:   ".yaml",
			haystack: []string{},
			expected: false,
		},
		{
			name:     "empty needle",
			needle:   "",
			haystack: []string{".yml", ".yaml", ".json"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inArray(tt.needle, tt.haystack)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOpenFile(t *testing.T) {
	t.Run("open stdin when filename matches stdin name", func(t *testing.T) {
		f, err := OpenFile(os.Stdin.Name())
		require.NoError(t, err)
		assert.Equal(t, os.Stdin, f)
	})

	t.Run("open stdin when filename is dash", func(t *testing.T) {
		f, err := OpenFile("-")
		require.NoError(t, err)
		assert.Equal(t, os.Stdin, f)
	})

	t.Run("open existing file", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "test-open-file-*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		// Open the file using our function
		f, err := OpenFile(tmpFile.Name())
		require.NoError(t, err)
		defer f.Close()
		
		assert.NotNil(t, f)
	})

	t.Run("error opening non-existent file", func(t *testing.T) {
		_, err := OpenFile("/non/existent/file.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to open file")
	})
}

func TestLoadFile(t *testing.T) {
	t.Run("load existing file", func(t *testing.T) {
		// Create a temporary file with content
		content := []byte("test content")
		tmpFile, err := os.CreateTemp("", "test-load-file-*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		
		_, err = tmpFile.Write(content)
		require.NoError(t, err)
		tmpFile.Close()

		// Load the file
		buf, err := LoadFile(tmpFile.Name())
		require.NoError(t, err)
		assert.Equal(t, "test content", buf.String())
	})

	t.Run("error loading non-existent file", func(t *testing.T) {
		_, err := LoadFile("/non/existent/file.txt")
		assert.Error(t, err)
	})
}

func TestLoadFolder(t *testing.T) {
	t.Run("load files from folder with matching extensions", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "test-load-folder-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create files with different extensions
		files := map[string]string{
			"file1.yaml": "content1",
			"file2.yml":  "content2",
			"file3.json": "content3",
			"file4.txt":  "content4",
		}

		for name, content := range files {
			err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
			require.NoError(t, err)
		}

		// Test loading only yaml/yml files
		extensions := []string{".yaml", ".yml"}
		buf, count, err := LoadFolder(extensions, tmpDir, false)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		
		// The buffer should contain content1 and content2 with separator
		// Order is not guaranteed, so we check for both possibilities
		expected1 := "content1\n---\ncontent2"
		expected2 := "content2\n---\ncontent1"
		bufStr := buf.String()
		assert.True(t, bufStr == expected1 || bufStr == expected2, "Expected buffer to contain concatenated yaml/yml file contents")
	})

	t.Run("load files recursively", func(t *testing.T) {
		// Create a temporary directory structure
		tmpDir, err := os.MkdirTemp("", "test-load-folder-recursive-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create subdirectory
		subDir := filepath.Join(tmpDir, "subdir")
		require.NoError(t, os.Mkdir(subDir, 0755))

		// Create files in main and sub directory
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.yaml"), []byte("main-content"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(subDir, "sub.yaml"), []byte("sub-content"), 0644))

		// Test with recursion enabled
		extensions := []string{".yaml"}
		buf, count, err := LoadFolder(extensions, tmpDir, true)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Contains(t, buf.String(), "main-content")
		assert.Contains(t, buf.String(), "sub-content")

		// Test with recursion disabled
		buf, count, err = LoadFolder(extensions, tmpDir, false)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Contains(t, buf.String(), "main-content")
		assert.NotContains(t, buf.String(), "sub-content")
	})

	t.Run("error when no matching files found", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "test-load-folder-empty-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Test with extensions that don't match any files
		extensions := []string{".xyz"}
		_, _, err = LoadFolder(extensions, tmpDir, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no files found")
	})

	t.Run("error with non-existent folder", func(t *testing.T) {
		extensions := []string{".yaml"}
		_, _, err := LoadFolder(extensions, "/non/existent/folder", false)
		assert.Error(t, err)
	})
}

func TestDeleteFolderContents(t *testing.T) {
	t.Run("delete folder contents", func(t *testing.T) {
		// Create a temporary directory with contents
		tmpDir, err := os.MkdirTemp("", "test-delete-folder-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create files and subdirectory
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644))
		subDir := filepath.Join(tmpDir, "subdir")
		require.NoError(t, os.Mkdir(subDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("content"), 0644))

		// Delete contents
		err = DeleteFolderContents(tmpDir)
		require.NoError(t, err)

		// Verify contents are deleted but directory still exists
		entries, err := os.ReadDir(tmpDir)
		require.NoError(t, err)
		assert.Empty(t, entries)
		_, err = os.Stat(tmpDir)
		assert.NoError(t, err)
	})

	t.Run("error with non-existent folder", func(t *testing.T) {
		err := DeleteFolderContents("/non/existent/folder")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to open folder")
	})
}
