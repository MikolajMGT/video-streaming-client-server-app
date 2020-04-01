package ui

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"streming_server/client/ui/resources"
)

type View struct {
	Window           fyne.Window
	Image            *canvas.Image
	ButtonsContainer *fyne.Container
	StatisticsBox    *widget.Box

	// methods to call on specific button click
	onSetup    func()
	onPlay     func()
	onPause    func()
	onDescribe func()
	onTeardown func()
}

func NewView(OnSetup func(), OnPlay func(), OnPause func(), OnDescribe func(), OnTeardown func()) *View {
	view := &View{
		onSetup:    OnSetup,
		onPlay:     OnPlay,
		onPause:    OnPause,
		onDescribe: OnDescribe,
		onTeardown: OnTeardown,
	}
	view.InitView()
	view.onSetup = OnSetup
	return view
}

func (view *View) InitView() {
	windowApp := app.New()

	view.Window = windowApp.NewWindow("Streaming Client")
	view.Window.SetIcon(resolveIcon("cast"))

	view.Image = prepareImage()
	view.ButtonsContainer = view.prepareButtonsContainer()
	view.StatisticsBox = prepareStatisticsBox()

	view.Window.SetContent(
		fyne.NewContainerWithLayout(layout.NewCenterLayout(),
			widget.NewVBox(
				view.Image,
				view.ButtonsContainer,
				view.StatisticsBox,
			),
		),
	)
}

func (view *View) StartGUI() {
	view.Window.ShowAndRun()
}

func (view *View) UpdateImage(newImage []byte) {
	view.Image.Resource = fyne.NewStaticResource("img", newImage)
	view.Image.FillMode = canvas.ImageFillOriginal
	canvas.Refresh(view.Image)
}

func (view *View) UpdateStatistics(totalBytesReceived int, packageLost int, dataRate float64) {
	view.StatisticsBox.Children[0].(*widget.Label).SetText(
		fmt.Sprint(resources.TotalBytesReceivedText, totalBytesReceived),
	)
	view.StatisticsBox.Children[1].(*widget.Label).SetText(
		fmt.Sprint(resources.PackageLostText, packageLost),
	)
	view.StatisticsBox.Children[2].(*widget.Label).SetText(
		fmt.Sprint(resources.DataRateText, dataRate),
	)
	view.StatisticsBox.Refresh()
}

func resolveIcon(name string) fyne.Resource {
	icon, _ := fyne.LoadResourceFromPath(fmt.Sprintf("client/ui/resources/icons/%v-icon.png", name))
	return icon
}

func prepareImage() *canvas.Image {
	imageFrame := canvas.NewImageFromFile("")
	imageFrame.FillMode = canvas.ImageFillContain
	return imageFrame
}

func (view *View) prepareButtonsContainer() *fyne.Container {
	buttons := []*widget.Button{
		widget.NewButtonWithIcon("Setup", resolveIcon("setup"), view.onSetup),
		widget.NewButtonWithIcon("Play", resolveIcon("play"), view.onPlay),
		widget.NewButtonWithIcon("Pause", resolveIcon("pause"), view.onPause),
		widget.NewButtonWithIcon("Describe", resolveIcon("describe"), view.onDescribe),
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
		widget.NewLabel(fmt.Sprint(resources.TotalBytesReceivedText, 0)),
		widget.NewLabel(fmt.Sprint(resources.PackageLostText, 0)),
		widget.NewLabel(fmt.Sprint(resources.DataRateText, 0)),
	)
	return result
}
