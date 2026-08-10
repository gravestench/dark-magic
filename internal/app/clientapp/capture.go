package clientapp

import "errors"

// startCapture adds a post-frame observer only for screenshot runs. Waiting
// until after the frame matters: that is when the finished picture exists.
func (app *application) startCapture() error {
	if app.options.CaptureDirectory == "" {
		return nil
	}
	if app.options.NewCapture == nil {
		return wrap("start scene capture", errors.New("capture factory is not configured"))
	}
	var err error
	app.capture, err = app.options.NewCapture(app.options.CaptureDirectory, app.options.CaptureScenes, app.options.CaptureSettle, app.renderer)
	if err != nil {
		return wrap("start scene capture", err)
	}
	app.stopCapture = app.renderer.SubscribePostFrame(func() {
		app.capture.Observe(app.navigator.Stack(), app.composer.Diagnostics().StructuralRevision)
		if app.capture.Complete() {
			app.stop()
		}
	})
	return nil
}
