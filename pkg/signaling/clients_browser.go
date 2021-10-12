//go:build js && wasm

package signaling

import "github.com/maxence-charriere/go-app/v9/pkg/app"

func Save(file File) {
	blob := app.Window().JSValue().Get("Blob").New([]interface{}{string(file.Payload)}, map[string]interface{}{
		"type": "application/octet-stream",
	})

	link := app.Window().Get("document").Call("createElement", "a")
	link.Set("href", app.Window().Get("URL").Call("createObjectURL", blob))
	link.Set("download", file.Name)
	link.Call("click")
}
