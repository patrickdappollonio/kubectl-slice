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

		s.fileCount++
		return nil
	}

	// Add an empty line at the end
	file = append(file, []byte("\n\n")...)

	// Send it for processing
	name, kind, err := s.parseYAMLManifest(file, s.fileCount, s.template)
	if err != nil {
		switch err.(type) {
		case *kindSkipErr:
			s.log.Printf("Skipping file %d: %s", s.fileCount, err.Error())
			s.fileCount++
			return nil

		case *strictModeErr:
			s.log.Printf("Skipping file %d: %s", s.fileCount, err.Error())
			s.fileCount++
			return nil

		default:
			return err
		}
	}

	existentData, position := []byte(nil), -1
	for pos := 0; pos < len(s.filesFound); pos++ {
		if s.filesFound[pos].name == name {
			existentData = s.filesFound[pos].data
			position = pos
			break
		}
	}

	if position == -1 {
		s.log.Printf("Got nonexistent file. Adding it to the list: %s", name)
		s.filesFound = append(s.filesFound, yamlFile{
			name: name,
			kind: kind,
			data: file,
		})
	} else {
		s.log.Printf("Got existent file. Appending to original buffer: %s", name)
		existentData = append(existentData, []byte("---\n\n")...)
		existentData = append(existentData, file...)
		s.filesFound[position] = yamlFile{
			name: name,
			kind: kind,
			data: existentData,
		}
	}

	s.fileCount++
	return nil
}

func (s *Split) scan() error {
	s.fileCount = 0

	// Since we'll be iterating over files that potentially might end up being
	// duplicated files, we need to store them somewhere to, later, save them
	// to files
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
				if err := parseFile(); err != nil {
					return err
				}
				break
			}

			// Otherwise handle the unexpected error
			return fmt.Errorf("unable to read YAML file number %d: %w", s.fileCount, err)
		}

		// Check if we're at the end of the file
		if line == "---\n" {
			s.log.Println("Found the end of a file. Sending buffer to process.")
			if err := parseFile(); err != nil {
				return err
			}
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

	// Now save those files to disk (or if dry-run is on, print what it would
	// save). Files will be overwritten.
	s.fileCount = 0
	for _, v := range s.filesFound {
		s.fileCount++

		fullpath := filepath.Join(s.opts.OutputDirectory, v.name)
		fileLength := len(v.data)

		s.log.Printf("Handling file %q: %d bytes long.", fullpath, fileLength)

		switch {
		case s.opts.DryRun:
			fmt.Fprintf(os.Stderr, "Would write %s -- %d bytes.\n", fullpath, fileLength)
			continue

		case s.opts.OutputToStdout:
			if s.fileCount != 1 {
				fmt.Fprintf(os.Stdout, "---\n\n")
			}

			fmt.Fprintf(os.Stdout, "# File: %s (%d bytes)\n", fullpath, fileLength)
			fmt.Fprintf(os.Stdout, "%s\n", v.data)
			continue

		default:
			// do nothing, handling below
			if err := writeToFile(fullpath, v.data); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Wrote %s -- %d bytes.\n", fullpath, fileLength)
			continue
		}
	}

	switch {
	case s.opts.DryRun:
		fmt.Fprintf(os.Stderr, "%d %s generated (dry-run)\n", s.fileCount, pluralize("file", s.fileCount))

	case s.opts.OutputToStdout:
		fmt.Fprintf(os.Stderr, "%d %s parsed to stdout.\n", s.fileCount, pluralize("file", s.fileCount))

	default:
		fmt.Fprintf(os.Stderr, "%d %s generated.\n", s.fileCount, pluralize("file", s.fileCount))
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

func writeToFile(path string, data []byte) error {
	// Since a single Go Template File Name might render different folder prefixes,
	// we need to ensure they're all created.
	if err := os.MkdirAll(filepath.Dir(path), folderChmod); err != nil {
		return fmt.Errorf("unable to create output folder for file %q: %w", path, err)
	}

	// Open the file as read/write, create the file if it doesn't exist, and if
	// it does, truncate it.
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, defaultChmod)
	if err != nil {
		return fmt.Errorf("unable to create/open file %q: %w", path, err)
	}

	// Write the contents from the buffer
	if _, err := f.Write(data); err != nil {
		f.Close()
		return fmt.Errorf("unable to write file contents for file %q: %w", path, err)
	}

	// Attempt to close the file cleanly
	if err := f.Close(); err != nil {
		return fmt.Errorf("unable to close file after write for file %q: %w", path, err)
	}

	return nil
}
