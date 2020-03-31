package ui

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

type View struct {
}

func (view *View) InitView() {
	windowApp := app.New()
	w := windowApp.NewWindow("Streaming Client")
	w.SetIcon(resolveIcon("cast"))
	w.SetContent(
		fyne.NewContainerWithLayout(layout.NewCenterLayout(),
			widget.NewVBox(
				prepareImage(),
				prepareButtonsContainer(),
				prepareStatisticsBox(),
			),
		),
	)
	w.ShowAndRun()
}

func prepareImage() *canvas.Image {
	image := canvas.NewImageFromFile("test.jpg")
	image.FillMode = canvas.ImageFillOriginal
	return image
}

func prepareButtonsContainer() *fyne.Container {
	buttons := []*widget.Button{
		widget.NewButtonWithIcon("Setup", resolveIcon("setup"), func() {}),
		widget.NewButtonWithIcon("Play", resolveIcon("play"), func() {}),
		widget.NewButtonWithIcon("Pause", resolveIcon("pause"), func() {}),
		widget.NewButtonWithIcon("Describe", resolveIcon("describe"), func() {}),
	}

	result := make([]fyne.CanvasObject, 0)
	for _, button := range buttons {
		elem := fyne.CanvasObject(button)
		result = append(result, elem)
	}

	return fyne.NewContainerWithLayout(layout.NewCenterLayout(), widget.NewHBox(result...))
}

func prepareStatisticsBox() *widget.Box {
	result := widget.NewVBox(
		widget.NewLabel("Total Bytes Received:"),
		widget.NewLabel("Package Lost:"),
		widget.NewLabel("Data Rate (bytes/sec):"),
	)
	return result
}

func resolveIcon(name string) fyne.Resource {
	icon, _ := fyne.LoadResourceFromPath(fmt.Sprintf("client/ui/resources/%v-icon.png", name))
	return icon
}
