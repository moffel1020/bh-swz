package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/moffel1020/bh-swz/swz"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

const keyPrefName string = "key"

func makeDecryptContainer(w *fyne.Window) *fyne.Container {
	path := widget.NewLabel("No file selected")

	pickFile := widget.NewButton("select .swz file", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, *w)
				return
			}
			if reader == nil {
				return
			}

			path.SetText(reader.URI().Path())
		}, *w)
		fd.Resize(fyne.NewSize(750, 550))
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".swz"}))
		fd.Show()
	})

	keyLabel := widget.NewLabel("key:")
	keyEntry := widget.NewEntry()
	keyEntry.SetText(fyne.CurrentApp().Preferences().String(keyPrefName))

	fileSelect := container.NewGridWithColumns(2, path, pickFile)
	keyInput := container.NewGridWithColumns(2, keyLabel, keyEntry)

	decryptButton := widget.NewButton("decrypt", func() {
		k, err := strconv.ParseUint(keyEntry.Text, 10, 32)
		if err != nil {
			dialog.ShowError(err, *w)
			return
		}
		key := uint32(k)

		if filepath.Ext(path.Text) != ".swz" {
			dialog.ShowError(errors.New("invalid .swz file"), *w)
			return
		}

		fyne.CurrentApp().Preferences().SetString(keyPrefName, keyEntry.Text)
		fmt.Println("decrypting: " + path.Text)
		fmt.Println("with key: " + fmt.Sprint(key))
		if err = swz.DecryptFile(path.Text, key); err != nil {
			dialog.ShowError(err, *w)
			return
		}
	})

	decryptContainer := container.NewVBox(fileSelect, keyInput, decryptButton)
	return decryptContainer
}

func makeEncryptContainer(w *fyne.Window) *fyne.Container {
	path := widget.NewLabel("No folder selected")

	pickFolder := widget.NewButton("select folder to encrypt", func() {
		fd := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, *w)
				return
			}
			if list == nil {
				fmt.Println("canceled folder selection")
				return
			}

			path.SetText(list.Path())
		}, *w)
		fd.Resize(fyne.NewSize(750, 550))
		fd.Show()
	})

	keyLabel := widget.NewLabel("key:")
	keyEntry := widget.NewEntry()
	keyEntry.SetText(fyne.CurrentApp().Preferences().String(keyPrefName))

	folderSelect := container.NewGridWithColumns(2, path, pickFolder)
	keyInput := container.NewGridWithColumns(2, keyLabel, keyEntry)

	encryptButton := widget.NewButton("encrypt", func() {
		k, err := strconv.ParseUint(keyEntry.Text, 10, 32)
		if err != nil {
			dialog.ShowError(err, *w)
			return
		}

		key := uint32(k)

		_, err = os.Stat(path.Text)
		if err != nil {
			dialog.ShowError(err, *w)
			return
		}

		fyne.CurrentApp().Preferences().SetString(keyPrefName, keyEntry.Text)
		fmt.Println("encrypting: " + path.Text)
		fmt.Println("with key: " + fmt.Sprint(key))
		if err = swz.EncryptToFile(path.Text, key, 0); err != nil {
			dialog.ShowError(err, *w)
			return
		}
	})

	decryptContainer := container.NewVBox(folderSelect, keyInput, encryptButton)
	return decryptContainer
}

func main() {
	a := app.NewWithID("com.moffel.swz")
	w := a.NewWindow("bh-swz")

	tabs := container.NewAppTabs(
		container.NewTabItem("Decrypt", makeDecryptContainer(&w)),
		container.NewTabItem("Encrypt", makeEncryptContainer(&w)),
	)

	fullLayout := container.NewVBox(tabs)

	w.Resize(fyne.NewSize(800, 600))
	w.SetContent(fullLayout)

	w.Show()
	a.Run()
}
