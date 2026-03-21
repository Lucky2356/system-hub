package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Универсальная карточка
func NewStatCard(title string, subtitle string, content fyne.CanvasObject) *widget.Card {
	return widget.NewCard(
		title,
		subtitle,
		container.NewPadded(content),
	)
}