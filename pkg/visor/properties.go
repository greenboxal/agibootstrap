package visor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type PropertyInspector struct {
	fyne.CanvasObject
}

func NewPropertyInspector() *PropertyInspector {
	propertiesToolBar := container.NewHBox()
	propertiesList := widget.NewCard("Properties", "Properties", widget.NewLabel("Properties"))

	propertiesPanel := container.NewBorder(
		propertiesToolBar,
		nil,
		nil,
		container.NewMax(container.NewVScroll(propertiesList)),
	)

	return &PropertyInspector{
		CanvasObject: propertiesPanel,
	}
}
