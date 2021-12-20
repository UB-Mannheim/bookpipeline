// Copyright 2021 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// copyStdoutToChan creates a pipe to copy anything written
// to the file also to a rune channel.
func copyStdoutToChan() (chan rune, error) {
	c := make(chan rune)

	origFile := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return c, fmt.Errorf("Error creating pipe for file redirection: %v", err)
	}
	os.Stdout = w

	bufReader := bufio.NewReader(r)

	go func() {
		defer func() {
			close(c)
			w.Close()
			os.Stdout = origFile
		}()
		for {
			r, _, err := bufReader.ReadRune()
			if err != nil && err != io.EOF {
				return
			}
			c <- r
			if err == io.EOF {
				return
			}
		}
	}()

	return c, nil
}

// copyStderrToChan creates a pipe to copy anything written
// to the file also to a rune channel.
// TODO: would be nice to merge this with copyStdoutToChan,
//       but a naive version using *os.File didn't work.
func copyStderrToChan() (chan rune, error) {
	c := make(chan rune)

	origFile := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		return c, fmt.Errorf("Error creating pipe for file redirection: %v", err)
	}
	os.Stderr = w

	bufReader := bufio.NewReader(r)

	go func() {
		defer func() {
			close(c)
			w.Close()
			os.Stderr = origFile
		}()
		for {
			r, _, err := bufReader.ReadRune()
			if err != nil && err != io.EOF {
				return
			}
			c <- r
			if err == io.EOF {
				return
			}
		}
	}()

	return c, nil
}

// startGui starts the gui process
func startGui(log log.Logger, cmd string, training string, tessdir string) error {
	myApp := app.New()
	myWindow := myApp.NewWindow("Rescribe OCR")

	myWindow.Resize(fyne.NewSize(800, 400))

	var gobtn *widget.Button
	var fullContent *fyne.Container

	dir := widget.NewLabel("")

	dirIcon := widget.NewIcon(theme.FolderIcon())

	folderBtn := widget.NewButtonWithIcon("Choose folder", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				dir.SetText(uri.Path())
				dirIcon.SetResource(theme.FolderIcon())
				myWindow.SetContent(fullContent)
				gobtn.Enable()
			}
		}, myWindow)
	})

	pdfBtn := widget.NewButtonWithIcon("Choose PDF", theme.DocumentIcon(), func() {
		dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err == nil && uri != nil {
				uri.Close()
				dir.SetText(uri.URI().Path())
				dirIcon.SetResource(theme.DocumentIcon())
				myWindow.SetContent(fullContent)
				gobtn.Enable()
			}
		}, myWindow)
	})

	gbookBtn := widget.NewButtonWithIcon("Get Google Book", theme.SearchIcon(), func() {
			// TODO
	})

	progressBar := widget.NewProgressBar()

	logarea := widget.NewMultiLineEntry()
	logarea.Disable()

	// TODO: have the button be pressed if enter is pressed
	gobtn = widget.NewButtonWithIcon("Start OCR", theme.UploadIcon(), func() {
		if dir.Text == "" {
			return
		}

		gobtn.Disable()
		gobtn.SetText("Processing...")

		progressBar.SetValue(0.5)

		stdout, err := copyStdoutToChan()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error copying stdout to chan: %v\n", err)
			return
		}

		// update log area with stdout in a concurrent goroutine
		go func() {
			for r := range stdout {
				logarea.SetText(logarea.Text + string(r))
				logarea.CursorRow = strings.Count(logarea.Text, "\n")
				// TODO: set text on progress bar to latest line printed using progressBar.TextFormatter, rather than just using a whole multiline entry like this
				// TODO: parse the stdout and set progressBar based on that
			}
		}()

		stderr, err := copyStderrToChan()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error copying stderr to chan: %v\n", err)
			return
		}

		// update log area with stderr in a concurrent goroutine
		go func() {
			for r := range stderr {
				logarea.SetText(logarea.Text + string(r))
				logarea.CursorRow = strings.Count(logarea.Text, "\n")
				// TODO: set text on progress bar, or a label below it, to latest line printed, rather than just using a whole multiline entry like this
			}
		}()

		err = startProcess(log, cmd, dir.Text, filepath.Base(dir.Text), training, dir.Text, tessdir)
		if err != nil {
			// add a newline before this printing as another message from stdout
			// or stderr may well be half way through printing
			logarea.SetText(logarea.Text + fmt.Sprintf("\nError executing process: %v\n", err))
			logarea.CursorRow = strings.Count(logarea.Text, "\n")
			progressBar.SetValue(0.0)
			gobtn.SetText("Process OCR")
			gobtn.Enable()
			return
		}

		progressBar.SetValue(1.0)
		gobtn.SetText("Process OCR")
		gobtn.Enable()
	})
	gobtn.Disable()

	choices := container.New(layout.NewGridLayout(3), folderBtn, pdfBtn, gbookBtn)

	chosen := container.New(layout.NewBorderLayout(nil, nil, dirIcon, nil), dirIcon, dir)

	fullContent = container.NewVBox(choices, chosen, gobtn, progressBar, logarea)
	startContent := container.NewVBox(choices, gobtn, progressBar, logarea)

	myWindow.SetContent(startContent)

	myWindow.Show()
	myApp.Run()

	return nil
}
