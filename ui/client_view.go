package ui

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"log"
	"streming_server/ui/resources"
	"streming_server/video"
)

type View struct {
	FrameSync        *video.FrameSync
	Window           fyne.Window
	Image            *canvas.Image
	ButtonsContainer *fyne.Container
	StatisticsBox    *widget.Box

	// methods to call on specific button click
	onSetup    func()
	onRecord   func()
	onPlay     func()
	onPause    func()
	onDescribe func()
	onTeardown func()
}

func NewView(frameSync *video.FrameSync,
	OnSetup func(), OnRecord func(), OnPlay func(), OnPause func(), OnDescribe func(), OnTeardown func()) *View {
	view := &View{
		FrameSync:  frameSync,
		onSetup:    OnSetup,
		onRecord:   OnRecord,
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

	view.Window.SetOnClosed(func() {
		view.onTeardown()
	})
}

func (view *View) StartGUI() {
	view.Window.ShowAndRun()
}

func (view *View) UpdateImage() {
	if !view.FrameSync.Empty() {
		view.Image.Resource = fyne.NewStaticResource("livestream", view.FrameSync.NextFrame())
		canvas.Refresh(view.Image)
	}
}

func (view *View) UpdateStatistics(totalBytesReceived int, packageLost int, dataRate float64) {
	view.StatisticsBox.Children[0].(*widget.Label).SetText(
		fmt.Sprint(resources.TotalBytesReceivedText, totalBytesReceived),
	)
	view.StatisticsBox.Children[1].(*widget.Label).SetText(
		fmt.Sprint(resources.PackageLostText, packageLost),
	)
	view.StatisticsBox.Children[2].(*widget.Label).SetText(
		fmt.Sprintf("%v%.2f", resources.DataRateText, dataRate),
	)
	view.StatisticsBox.Refresh()
}

func resolveIcon(name string) fyne.Resource {
	icon, err := fyne.LoadResourceFromPath(fmt.Sprintf("ui/resources/icons/%v-icon.png", name))
	if err != nil {
		log.Println("[ERROR] cannot retrieve resource:", err)
	}
	return icon
}

func prepareImage() *canvas.Image {
	imageFrame := canvas.NewImageFromFile("")
	imageFrame.FillMode = canvas.ImageFillContain
	imageFrame.SetMinSize(fyne.NewSize(1280, 720))
	return imageFrame
}

func (view *View) prepareButtonsContainer() *fyne.Container {
	buttons := []*widget.Button{
		widget.NewButtonWithIcon("Setup", resolveIcon("setup"), view.onSetup),
		widget.NewButtonWithIcon("Record", resolveIcon("cast"), view.onRecord),
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
