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

// Execute runs the process according to the split.Options provided. This will
// generate the files in the given directory.
func (s *Split) Execute() error {
	// Since we'll be iterating over files that potentially might end up being
	// duplicated files, we need to store them somewhere to, later, save them
	// to files
	allFiles := make(map[string]bytes.Buffer)

	// We can totally create a single decoder then decode using that, however,
	// we want to maintain 1:1 exactly the same declaration as the YAML originally
	// fed by the user, so we split and save copies of these resources locally.
	// If we re-marshal the YAML, it might lose the format originally provided
	// by the user.
	scanner := bufio.NewReader(s.data)

	// Create a local buffer to read files line by line
	local := new(bytes.Buffer)

	// Holder for file count
	fileCount := 1

	// Handle the processing of a single YAML and add it to the list of
	// found files for later handling
	fnProcessFile := func() error {
		s.log.Printf("Found a new YAML file in buffer, number %d", fileCount)

		// Grab local copy of the full file
		currentFile := local.Bytes()
		local = new(bytes.Buffer)

		// If there's no data in the buffer, return without doing anything
		// but count the file
		if len(currentFile) == 0 {
			// If it is the first file, it means the original file started
			// with "---", which is valid YAML, but we don't count it
			// as a file.
			if fileCount == 1 {
				s.log.Println("Got empty file. Skipping.")
				return nil
			}

			fileCount++
			return nil
		}

		// Send it for processing
		name, err := s.processSingleYAML(currentFile, fileCount, s.template)
		if err != nil {
			switch err.(type) {
			case *kindSkipErr:
				s.log.Printf("Skipping file %d: %s", fileCount, err.Error())
				fileCount++
				return nil

			case *strictModeErr:
				s.log.Printf("Skipping file %d: %s", fileCount, err.Error())
				fileCount++
				return nil

			default:
				return err
			}
		}

		// See if we have a file with the custom name
		buf, found := allFiles[name]

		// If not, add it. If so, we append it
		if !found {
			s.log.Printf("Got nonexistent file. Adding it to the list: %s", name)
			allFiles[name] = *bytes.NewBuffer(currentFile)
		} else {
			s.log.Printf("Got existent file. Appending to original buffer: %s", name)
			fmt.Fprintln(&buf, "---")
			fmt.Fprintln(&buf, string(currentFile))
			allFiles[name] = buf
		}

		fileCount++
		return nil
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
				if err := fnProcessFile(); err != nil {
					return err
				}

				break
			}

			// Otherwise handle the unexpected error
			return fmt.Errorf("unable to read YAML file number %d: %w", fileCount, err)
		}

		// Check if we're at the end of the file
		if line == "---\n" {
			s.log.Println("Found the end of a file. Sending buffer to process.")
			if err := fnProcessFile(); err != nil {
				return err
			}

			continue
		}

		fmt.Fprint(local, line)
	}

	s.log.Printf(
		"Finished processing buffer. Generated %d individual files, and processed %d files in the original YAML.",
		len(allFiles), fileCount,
	)

	// Create the output folder if it doesn't exist
	if !s.opts.DryRun {
		s.log.Printf("Creating directory %q if it doesn't exist.", s.opts.OutputDirectory)
		if err := os.MkdirAll(s.opts.OutputDirectory, folderChmod); err != nil {
			return fmt.Errorf("unable to create output directory folder %q: %w", s.opts.OutputDirectory, err)
		}
	}

	// Now save those files to disk (or if dry-run is on, print what it would
	// save). Files will be overwritten.
	fileCount = 0
	for name, contents := range allFiles {
		fileCount++
		fullpath := filepath.Join(s.opts.OutputDirectory, name)
		fileLength := contents.Len()

		s.log.Printf("Handling file %q: %d bytes long.", fullpath, fileLength)

		if !s.opts.DryRun {
			dir := filepath.Dir(fullpath)
			s.log.Printf("Ensuring folder %q exists for current file.", dir)

			// Since a single Go Template File Name might render different folder prefixes,
			// we need to ensure they're all created.
			if err := os.MkdirAll(dir, folderChmod); err != nil {
				return fmt.Errorf("unable to create output folder for file %q: %w", fullpath, err)
			}

			// Open the file as read/write, create the file if it doesn't exist, and if
			// it does, truncate it.
			f, err := os.OpenFile(fullpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, defaultChmod)
			if err != nil {
				return fmt.Errorf("unable to create/open file %q: %w", fullpath, err)
			}

			// Write the contents from the buffer
			if _, err := f.Write(contents.Bytes()); err != nil {
				f.Close()
				return fmt.Errorf("unable to write file contents for file %q: %w", fullpath, err)
			}

			// Attempt to close the file cleanly
			if err := f.Close(); err != nil {
				return fmt.Errorf("unable to close file after write for file %q: %w", fullpath, err)
			}

			fmt.Fprintf(os.Stdout, "Wrote %s -- %d bytes.\n", fullpath, fileLength)

			// Go to the next file
			continue
		}

		fmt.Fprintf(os.Stdout, "Would write %s -- %d bytes.\n", fullpath, fileLength)
	}

	if fileCount == 1 {
		fmt.Fprintf(os.Stdout, "%d file generated.\n", fileCount)
	} else {
		fmt.Fprintf(os.Stdout, "%d files generated.\n", fileCount)
	}

	return nil
}
