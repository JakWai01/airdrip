package components

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/JakWai01/airdrip/pkg/signaling"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type MyComponent struct {
	app.Compo

	tag string
}

func (c *MyComponent) Render() app.UI {
	return app.Div().
		Body(
			app.Meta().
				Charset("utf-8"),
			app.Script().
				Src("wasm_exec.js"),
			app.H1().
				Body(
					app.Text("New Channel"),
				),
			app.Label().
				For("myfile").
				Body(
					app.Text("Select a file:"),
				),
			app.Input().
				Type("File").
				ID("myFile").
				Name("myfile").OnChange(func(ctx app.Context, e app.Event) {
				e.PreventDefault()

				reader := app.Window().JSValue().Get("FileReader").New()
				input := app.Window().GetElementByID("myFile")

				reader.Set("onload", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
					go func() {
						rawFileContent := app.Window().Get("Uint8Array").New(args[0].Get("target").Get("result"))

						fileContent := make([]byte, rawFileContent.Get("length").Int())
						app.CopyBytesToGo(fileContent, rawFileContent)

						// Generate community name
						communityName := generateCommunityName()
						fmt.Println(communityName)

						ctx.Dispatch(func(_ app.Context) {
							c.tag = communityName
						})

						// get filename
						filename := app.Window().GetElementByID("myFile").Get("value").String()
						fmt.Println(strings.TrimPrefix(filename, `C:\fakepath\`))

						// call send function here
						fmt.Println(string(fileContent))

						client := signaling.NewSignalingClient()
						go client.HandleConn("airdrip.herokuapp.com", communityName, strings.TrimPrefix(filename, `C:\fakepath\`), fileContent)

					}()

					return nil
				}))

				if file := input.Get("files").Get("0"); !file.IsUndefined() {
					reader.Call("readAsArrayBuffer", file)
				} else {
					c.clear()
				}

			}),
			app.P().
				ID("tag").
				Body(
					app.Text("Channel tag:"+c.tag),
				),
			app.H1().
				Body(
					app.Text("Join Channel"),
				),
			app.P().
				Body(
					app.Text("Enter channel tag:"),
				),
			app.Input().
				Type("text").
				ID("fname").
				Name("fname"),
			app.Button().
				Body(
					app.Text("Join"),
				).OnClick(func(ctx app.Context, e app.Event) {
				communityName := app.Window().GetElementByID("fname").Get("value").String()
				fmt.Println(communityName)

				// call send function
				client := signaling.NewSignalingClient()
				go client.HandleConn("airdrip.herokuapp.com", communityName, "", []byte(""))
			}),
		)
}

func getRandomInt(max int) int {

	num, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		log.Fatal(err)
	}
	return int(num.Int64())
}

func generateCommunityName() string {
	words := []string{"Tux", "Linux", "OSS", "Arch", "Debian", "Mint", "Fedora", "Open", "HdM", "Selfnet", "SSL", "Telnet", "GPL"}
	return words[getRandomInt(len(words))] + "-" + words[getRandomInt(len(words))] + "-" + words[getRandomInt(len(words))]
}

func (c *MyComponent) clear() {
	// Clear input value
	app.Window().GetElementByID("myFile").Set("value", app.Null())
}
