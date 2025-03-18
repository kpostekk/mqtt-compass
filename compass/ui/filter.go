package ui

import (
	"strings"

	"github.com/mappu/miqt/qt6/mainthread"
)

func (app *AppUi) UpdateFiltersResults() {
	if app.CurrentFilterText == "" {
		for _, item := range app.MapTopicQTreeEntry {
			item.SetHidden(false)
		}

		return
	}

	for _, item := range app.MapTopicQTreeEntry {
		item.SetHidden(true)
	}

	for topic, item := range app.MapTopicQTreeEntry {
		matches := strings.Contains(topic, app.CurrentFilterText)

		if matches {
			showWithParents(item)
			expandWithParents(item)
		}
	}
}

func (app *AppUi) SetupFiltering() {
	app.CurrentFilterInput.OnTextChanged(func(text string) {
		app.CurrentFilterText = text

		go func() {
			app.TopicLock.RLock()
			mainthread.Wait(func() {
				app.UpdateFiltersResults()
			})
			app.TopicLock.RUnlock()
		}()
	})
}
