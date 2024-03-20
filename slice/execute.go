package slice

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	folderChmod  = 0775
	defaultChmod = 0664
)

func (s *Split) processSingleFile(file []byte) error {
	s.log.Printf("Found a new YAML file in buffer, number %d", s.fileCount)

	// If there's no data in the buffer, return without doing anything
	// but count the file
	file = bytes.TrimSpace(file)

	if len(file) == 0 {
		// If it is the first file, it means the original file started
		// with "---", which is valid YAML, but we don't count it
		// as a file.
		if s.fileCount == 1 {
			s.log.Println("Got empty file. Skipping.")
			return nil
		}

		return nil
	}

	// Send it for processing
	meta, err := s.parseYAMLManifest(file)
	if err != nil {
		switch err.(type) {
		case *skipErr:
			s.log.Printf("Skipping file %d: %s", s.fileCount, err.Error())
			return nil

		case *strictModeSkipErr:
			s.log.Printf("Skipping file %d: %s", s.fileCount, err.Error())
			return nil

		default:
			return err
		}
	}

	existentData, position := []byte(nil), -1
	for pos := 0; pos < len(s.filesFound); pos++ {
		if s.filesFound[pos].filename == meta.filename {
			existentData = s.filesFound[pos].data
			position = pos
			break
		}
	}

	if position == -1 {
		s.log.Printf("Got nonexistent file. Adding it to the list: %s", meta.filename)
		s.filesFound = append(s.filesFound, yamlFile{
			filename: meta.filename,
			meta:     meta.meta,
			data:     file,
		})
	} else {
		s.log.Printf("Got existent file. Appending to original buffer: %s", meta.filename)
		existentData = append(existentData, []byte("\n---\n")...)
		existentData = append(existentData, file...)
		s.filesFound[position] = yamlFile{
			filename: meta.filename,
			meta:     meta.meta,
			data:     existentData,
		}
	}

	return nil
}

func (s *Split) scan() error {
	// Since we'll be iterating over files that potentially might end up being
	// duplicated files, we need to store them somewhere to, later, save them
	// to files
	s.fileCount = 0
	s.filesFound = make([]yamlFile, 0)

	// We can totally create a single decoder then decode using that, however,
	// we want to maintain 1:1 exactly the same declaration as the YAML originally
	// fed by the user, so we split and save copies of these resources locally.
	// If we re-marshal the YAML, it might lose the format originally provided
	// by the user.
	scanner := bufio.NewReader(s.data)

	// Create a local buffer to read files line by line
	local := bytes.Buffer{}

	// Parse a single file
	parseFile := func() error {
		contents := local.Bytes()
		local = bytes.Buffer{}
		return s.processSingleFile(contents)
	}

	// Iterate over the entire buffer
	for {
		// Grab a single line
		line, err := scanner.ReadString('\n')

		// Find if there's an error
		if err != nil {
			// If we reached the end of file, handle up to this point
			if err == io.EOF {
				s.log.Println("Reached end of file while parsing. Sending remaining buffer to process.")
				local.WriteString(line)

				if err := parseFile(); err != nil {
					return err
				}

				s.fileCount++
				break
			}

			// Otherwise handle the unexpected error
			return fmt.Errorf("unable to read YAML file number %d: %w", s.fileCount, err)
		}

		// Check if we're at the end of the file
		if line == "---\n" || line == "---\r\n" {
			s.log.Println("Found the end of a file. Sending buffer to process.")
			if err := parseFile(); err != nil {
				return err
			}
			s.fileCount++
			continue
		}

		fmt.Fprint(&local, line)
	}

	s.log.Printf(
		"Finished processing buffer. Generated %d individual files, and processed %d files in the original YAML.",
		len(s.filesFound), s.fileCount,
	)

	return nil
}

func (s *Split) store() error {
	// Handle output directory being empty
	if s.opts.OutputDirectory == "" {
		s.opts.OutputDirectory = "."
	}

	// If the user wants to prune the output directory, do it
	if s.opts.PruneOutputDir && !s.opts.OutputToStdout && !s.opts.DryRun {
		// Check if the directory exists and if it does, prune it
		if _, err := os.Stat(s.opts.OutputDirectory); !os.IsNotExist(err) {
			s.log.Printf("Pruning output directory %q", s.opts.OutputDirectory)
			if err := deleteFolderContents(s.opts.OutputDirectory); err != nil {
				return fmt.Errorf("unable to prune output directory %q: %w", s.opts.OutputDirectory, err)
			}
			s.log.Printf("Output directory %q pruned", s.opts.OutputDirectory)
		}
	}

	// Now save those files to disk (or if dry-run is on, print what it would
	// save). Files will be overwritten.
	s.fileCount = 0
	for _, v := range s.filesFound {
		s.fileCount++

		fullpath := filepath.Join(s.opts.OutputDirectory, v.filename)
		fileLength := len(v.data)

		s.log.Printf("Handling file %q: %d bytes long.", fullpath, fileLength)

		switch {
		case s.opts.DryRun:
			s.WriteStderr("Would write %s -- %d bytes.", fullpath, fileLength)
			continue

		case s.opts.OutputToStdout:
			if s.fileCount != 1 {
				s.WriteStdout("---")
			}

			s.WriteStdout("# File: %s (%d bytes)", fullpath, fileLength)
			s.WriteStdout("%s", v.data)
			continue

		default:
			local := make([]byte, 0, len(v.data)+4)

			// If the user wants to include the triple dash, add it
			// at the beginning of the file
			if s.opts.IncludeTripleDash && !bytes.Equal(v.data, []byte("---")) {
				local = append([]byte("---\n"), v.data...)
			} else {
				local = append(local, v.data...)
			}

			// do nothing, handling below
			if err := s.writeToFile(fullpath, local); err != nil {
				return err
			}

			s.WriteStderr("Wrote %s -- %d bytes.", fullpath, len(local))
			continue
		}
	}

	switch {
	case s.opts.DryRun:
		s.WriteStderr("%d %s generated (dry-run)", s.fileCount, pluralize("file", s.fileCount))

	case s.opts.OutputToStdout:
		s.WriteStderr("%d %s parsed to stdout.", s.fileCount, pluralize("file", s.fileCount))

	default:
		s.WriteStderr("%d %s generated.", s.fileCount, pluralize("file", s.fileCount))
	}

	return nil
}

func (s *Split) sort() {
	if s.opts.SortByKind {
		s.filesFound = sortYAMLsByKind(s.filesFound)
	}
}

// Execute runs the process according to the split.Options provided. This will
// generate the files in the given directory.
func (s *Split) Execute() error {
	if err := s.scan(); err != nil {
		return err
	}

	s.sort()

	return s.store()
}

func (s *Split) writeToFile(path string, data []byte) error {
	// Since a single Go Template File Name might render different folder prefixes,
	// we need to ensure they're all created.
	if err := os.MkdirAll(filepath.Dir(path), folderChmod); err != nil {
		return fmt.Errorf("unable to create output folder for file %q: %w", path, err)
	}

	// Open the file as read/write, create the file if it doesn't exist, and if
	// it does, truncate it.
	s.log.Printf("Opening file path %q for writing", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, defaultChmod)
	if err != nil {
		return fmt.Errorf("unable to create/open file %q: %w", path, err)
	}

	defer f.Close()

	// Check if the last character is a newline, and if not, add one
	if !bytes.HasSuffix(data, []byte{'\n'}) {
		s.log.Printf("Adding new line to end of contents (content did not end on a line break)")
		data = append(data, '\n')
	}

	// Write the entire file buffer back to the file in disk
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("unable to write file contents for file %q: %w", path, err)
	}

	// Attempt to close the file cleanly
	if err := f.Close(); err != nil {
		return fmt.Errorf("unable to close file after write for file %q: %w", path, err)
	}

	return nil
}
