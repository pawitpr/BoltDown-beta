package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type download struct {
	url      string
	filename string
	size     int64
	progress *widget.ProgressBar
}

type threadOption struct {
	threads int
}

func main() {
	a := app.New()
	w := a.NewWindow("BoltDown")

	threadsEntry := widget.NewEntry()
	threadsEntry.SetPlaceHolder("Enter number of threads")

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Enter download link")

	startBtn := widget.NewButton("Start", func() {
		if urlEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("Download link is required"), w)
			return
		}

		var threads int
		var err error
		if threadsEntry.Text == "" {
			threads = 1 // Default to 1 thread if none provided
		} else {
			threads, err = strconv.Atoi(threadsEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Invalid thread count: %v", err), w)
				return
			}
		}

		d := download{
			url:      urlEntry.Text,
			filename: filepath.Base(urlEntry.Text),
			progress: widget.NewProgressBar(),
		}

		go d.startDownload(w, threads)
	})

	progress := widget.NewProgressBar()

	w.SetContent(container.NewVBox(
		container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 40)),
			widget.NewLabel("Download Link:"),
			urlEntry,
			widget.NewLabel("Thread Count:"),
			threadsEntry,
		),
		progress,
		startBtn,
	))

	w.ShowAndRun()
}

func (d *download) startDownload(w fyne.Window, threads int) {
	res, err := http.Head(d.url)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Error getting content length: %v", err), w)
		return
	}
	d.size = res.ContentLength

	if res.StatusCode >= 400 {
		dialog.ShowError(fmt.Errorf("Error: %s", res.Status), w)
		return
	}

	if _, err := os.Stat(d.filename); err == nil {
		overwrite := dialog.NewConfirm("Overwrite File?", "File already exists. Do you want to overwrite it?", func(ok bool) {
			if ok {
				d.doDownload(w, threads)
			}
		}, w)
		overwrite.SetDismissText("Cancel")
		overwrite.Show()
	} else {
		d.doDownload(w, threads)
	}
}

func (d *download) doDownload(w fyne.Window, threads int) {
	out, err := os.Create(d.filename)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Error creating file: %v", err), w)
		return
	}
	defer out.Close()

	partSize := d.size / int64(threads)
	var wg sync.WaitGroup
	wg.Add(threads)

	for i := 0; i < threads; i++ {
		start := partSize * int64(i)
		var end int64
		if i == threads-1 {
			end = d.size
		} else {
			end = start + partSize - 1
		}

		go d.downloadPart(w, start, end, out, &wg)
	}

	wg.Wait()
	d.progress.SetValue(1)
	dialog.ShowInformation("Download Complete", fmt.Sprintf("File %s downloaded successfully", d.filename), w)
}

func (d *download) downloadPart(w fyne.Window, start, end int64, out *os.File, wg *sync.WaitGroup) {
	defer wg.Done()

	client := http.DefaultClient
	req, err := http.NewRequest("GET", d.url, nil)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Error creating request: %v", err), w)
		return
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	res, err := client.Do(req)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Error downloading file: %v", err), w)
		return
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		dialog.ShowError(fmt.Errorf("Error: %s", res.Status), w)
		return
	}

	buf := make([]byte, 1024*1024)
	var downloaded int64
	for {
		n, err := res.Body.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			dialog.ShowError(fmt.Errorf("Error downloading file: %v", err), w)
			return
		}

		n, err = out.WriteAt(buf[:n], start+downloaded)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Error writing to file: %v", err), w)
			return
		}

		downloaded += int64(n)
		d.progress.SetValue(float64(start+downloaded) / float64(d.size))
	}
}
