/*
 * Flow CLI
 *
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package super

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getTargetDirectory_NonExistentDirectory(t *testing.T) {
	targetDir, err := getTargetDirectory("non-existent-project")
	assert.NoError(t, err)
	assert.Contains(t, targetDir, "non-existent-project")
}

func Test_getTargetDirectory_ExistingEmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "empty-project")
	err := os.Mkdir(testDir, 0755)
	require.NoError(t, err)
	
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	
	targetDir, err := getTargetDirectory("empty-project")
	assert.NoError(t, err)
	assert.Contains(t, targetDir, "empty-project")
}

func Test_getTargetDirectory_ExistingNonEmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "non-empty-project")
	err := os.Mkdir(testDir, 0755)
	require.NoError(t, err)
	
	testFile := filepath.Join(testDir, "existing-file.txt")
	err = os.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)
	
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	
	_, err = getTargetDirectory("non-empty-project")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directory is not empty")
}

func Test_getTargetDirectory_ExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "project-file")
	err := os.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)
	
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	
	_, err = getTargetDirectory("project-file")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is a file")
}

func Test_copyFile(t *testing.T) {
	tempDir := t.TempDir()
	
	sourceContent := []byte("test file content\nwith multiple lines")
	sourcePath := filepath.Join(tempDir, "source.txt")
	err := os.WriteFile(sourcePath, sourceContent, 0644)
	require.NoError(t, err)
	
	destPath := filepath.Join(tempDir, "dest.txt")
	err = copyFile(sourcePath, destPath)
	assert.NoError(t, err)
	
	destContent, err := os.ReadFile(destPath)
	assert.NoError(t, err, "Destination file should exist")
	assert.Equal(t, sourceContent, destContent)
}

func Test_copyFile_NonExistentSource(t *testing.T) {
	err := copyFile("non-existent.txt", "dest.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func Test_copyDirContents_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")
	
	err := os.Mkdir(srcDir, 0755)
	require.NoError(t, err)
	err = os.Mkdir(dstDir, 0755)
	require.NoError(t, err)
	
	err = copyDirContents(srcDir, dstDir)
	assert.NoError(t, err)
	
	entries, err := os.ReadDir(dstDir)
	assert.NoError(t, err)
	assert.Len(t, entries, 0)
}

func Test_copyDirContents_WithFiles(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")
	
	err := os.Mkdir(srcDir, 0755)
	require.NoError(t, err)
	err = os.Mkdir(dstDir, 0755)
	require.NoError(t, err)
	
	testFiles := map[string]string{
		"file1.txt":    "content of file 1",
		"file2.txt":    "content of file 2",
		"config.json":  `{"test": true}`,
	}
	
	for filename, content := range testFiles {
		err = os.WriteFile(filepath.Join(srcDir, filename), []byte(content), 0644)
		require.NoError(t, err)
	}
	
	err = copyDirContents(srcDir, dstDir)
	assert.NoError(t, err)
	
	for filename, expectedContent := range testFiles {
		destFilePath := filepath.Join(dstDir, filename)
		
		_, err := os.Stat(destFilePath)
		assert.NoError(t, err, "File %s should exist", filename)
		
		actualContent, err := os.ReadFile(destFilePath)
		assert.NoError(t, err)
		assert.Equal(t, expectedContent, string(actualContent), "Content mismatch for %s", filename)
	}
}

func Test_copyDirContents_WithSubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")
	
	err := os.Mkdir(srcDir, 0755)
	require.NoError(t, err)
	err = os.Mkdir(dstDir, 0755)
	require.NoError(t, err)
	
	subDir := filepath.Join(srcDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)
	
	nestedSubDir := filepath.Join(subDir, "nested")
	err = os.Mkdir(nestedSubDir, 0755)
	require.NoError(t, err)
	
	err = os.WriteFile(filepath.Join(srcDir, "root.txt"), []byte("root content"), 0644)
	require.NoError(t, err)
	
	err = os.WriteFile(filepath.Join(subDir, "sub.txt"), []byte("sub content"), 0644)
	require.NoError(t, err)
	
	err = os.WriteFile(filepath.Join(nestedSubDir, "nested.txt"), []byte("nested content"), 0644)
	require.NoError(t, err)
	
	err = copyDirContents(srcDir, dstDir)
	assert.NoError(t, err)
	
	tests := []struct {
		path            string
		expectedContent string
	}{
		{"root.txt", "root content"},
		{"subdir/sub.txt", "sub content"},
		{"subdir/nested/nested.txt", "nested content"},
	}
	
	for _, test := range tests {
		destPath := filepath.Join(dstDir, test.path)
		
		_, err := os.Stat(destPath)
		assert.NoError(t, err, "File %s should exist", test.path)
		
		content, err := os.ReadFile(destPath)
		assert.NoError(t, err)
		assert.Equal(t, test.expectedContent, string(content), "Content mismatch for %s", test.path)
	}
}

func Test_copyDirContents_NonExistentSource(t *testing.T) {
	tempDir := t.TempDir()
	dstDir := filepath.Join(tempDir, "dst")
	err := os.Mkdir(dstDir, 0755)
	require.NoError(t, err)
	
	err = copyDirContents("non-existent-src", dstDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func Test_copyDirContents_SourceIsFile(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "src.txt")
	dstDir := filepath.Join(tempDir, "dst")
	
	err := os.WriteFile(srcFile, []byte("content"), 0644)
	require.NoError(t, err)
	err = os.Mkdir(dstDir, 0755)
	require.NoError(t, err)
	
	err = copyDirContents(srcFile, dstDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source is not a directory")
}