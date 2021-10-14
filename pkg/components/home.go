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

	tag         string
	fileName    string
	fileContent []byte
}

func (c *MyComponent) Render() app.UI {
	return app.Div().
		Class("pf-u-h-100").
		Body(
			app.Div().
				Class("pf-c-background-image").
				Body(
					app.Raw(`<svg
  xmlns="http://www.w3.org/2000/svg"
  class="pf-c-background-image__filter"
  width="0"
  height="0"
>
  <filter id="image_overlay">
    <feColorMatrix
      type="matrix"
      values="1 0 0 0 0 1 0 0 0 0 1 0 0 0 0 0 0 0 1 0"
    ></feColorMatrix>
    <feComponentTransfer
      color-interpolation-filters="sRGB"
      result="duotone"
    >
      <feFuncR
        type="table"
        tableValues="0.086274509803922 0.43921568627451"
      ></feFuncR>
      <feFuncG
        type="table"
        tableValues="0.086274509803922 0.43921568627451"
      ></feFuncG>
      <feFuncB
        type="table"
        tableValues="0.086274509803922 0.43921568627451"
      ></feFuncB>
      <feFuncA type="table" tableValues="0 1"></feFuncA>
    </feComponentTransfer>
  </filter>
</svg>`),
				),
			app.Div().
				Class("pf-c-login").
				Body(
					app.Div().
						Class("pf-c-login__container").
						Body(
							app.Main().
								Class("pf-c-login__main").
								Body(
									app.Header().
										Class("pf-c-login__main-header").
										Body(
											app.H1().
												Class("pf-c-title pf-m-3xl").
												Body(
													app.Text("Airdrip"),
												),
											app.P().
												Class("pf-c-login__main-header-desc").
												Body(
													app.Text("Peer-to-peer File Sharing"),
												),
										),
									app.Div().
										Class("pf-c-login__main-body").
										Body(
											app.Form().
												Class("pf-c-form").
												Body(
													app.P().
														Class("pf-c-form__helper-text pf-m-error pf-m-hidden").
														Body(
															app.Span().
																Class("pf-c-form__helper-text-icon").
																Body(
																	app.I().
																		Class("fas fa-exclamation-circle").
																		Aria("hidden", true),
																),
															app.Text("Invalid login credentials."),
														),
													app.Div().
														Class("pf-c-form__group").
														Body(
															app.Label().
																Class("pf-c-form__label").
																For("login-demo-form-username").
																Body(
																	app.Span().
																		Class("pf-c-form__label-text").
																		Body(
																			app.Text("Choose file you want to send"),
																		),
																	app.Span().
																		Class("pf-c-form__label-required").
																		Aria("hidden", true).
																		Body(
																			app.Text("*"),
																		),
																),
															app.Div().
																Class("pf-c-file-upload").
																Body(
																	app.Div().
																		Class("pf-c-file-upload__file-select").
																		Body(
																			app.Div().
																				Class("pf-c-input-group").
																				Body(
																					app.Input().
																						Class("pf-c-form-control").
																						ID("myFile").
																						Type("File").
																						ReadOnly(true),
																					app.Button().
																						Class("pf-c-button pf-m-control").
																						Type("button").
																						Text("Create").
																						OnClick(func(ctx app.Context, e app.Event) {
																							e.PreventDefault()

																							reader := app.Window().JSValue().Get("FileReader").New()
																							input := app.Window().GetElementByID("myFile")

																							reader.Set("onload", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
																								go func() {
																									rawFileContent := app.Window().Get("Uint8Array").New(args[0].Get("target").Get("result"))

																									c.fileContent = make([]byte, rawFileContent.Get("length").Int())
																									app.CopyBytesToGo(c.fileContent, rawFileContent)

																									communityName := generateCommunityName()
																									fmt.Println(communityName)

																									ctx.Dispatch(func(_ app.Context) {
																										c.tag = communityName
																									})

																									filename := app.Window().GetElementByID("myFile").Get("value").String()
																									fmt.Println(strings.TrimPrefix(filename, `C:\fakepath\`))

																									fmt.Println(string(c.fileContent))

																									client := signaling.NewSignalingClient()
																									//go client.HandleConn("100.64.84.75:23432", communityName, strings.TrimPrefix(filename, `C:\fakepath\`), c.fileContent)
																									go client.HandleConn("airdrip.herokuapp.com", communityName, strings.TrimPrefix(filename, `C:\fakepath\`), c.fileContent)
																								}()

																								return nil
																							}))

																							if file := input.Get("files").Get("0"); !file.IsUndefined() {
																								reader.Call("readAsArrayBuffer", file)
																							} else {
																								c.clear()
																							}
																						}),
																				),
																		),
																),
														),
													app.H1().
														ID("tag").
														Body(
															app.Text("Channel tag: "+c.tag),
														),
													app.Div().
														Class("pf-c-form__group").
														Body(
															app.Label().
																Class("pf-c-form__label").
																For("login-demo-form-password").
																Body(
																	app.Span().
																		Class("pf-c-form__label-text").
																		Body(
																			app.Text("Enter Tag below"),
																		),
																	app.Span().
																		Class("pf-c-form__label-required").
																		Aria("hidden", true).
																		Body(
																			app.Text("*"),
																		),
																),
															app.Div().
																Class("pf-c-search-input").
																Body(
																	app.Div().
																		Class("pf-c-input-group").
																		Body(
																			app.Div().
																				Class("pf-c-search-input__bar").
																				Body(
																					app.Span().
																						Class("pf-c-search-input__text").
																						Body(
																							app.Input().
																								Type("text").
																								ID("fname").
																								Placeholder("Linux-OSS-HdM").
																								Aria("label", "Find by name"),
																						),
																				),
																			app.Button().
																				Class("pf-c-button pf-m-control").
																				Type("submit").
																				Aria("label", "Search").
																				Body(
																					app.Text("Join"),
																				).
																				OnClick(func(ctx app.Context, e app.Event) {
																					e.PreventDefault()

																					communityName := app.Window().GetElementByID("fname").Get("value").String()
																					fmt.Println(communityName)

																					// call send function
																					client := signaling.NewSignalingClient()
																					// go client.HandleConn("100.64.84.75:23432", communityName, "", []byte(""))
																					go client.HandleConn("airdrip.herokuapp.com", communityName, "", []byte(""))
																				}),
																		),
																),
														),
												),
										),
									app.Footer().
										Class("pf-c-login__main-footer").
										Body(
											app.Div().
												Class("pf-c-login__main-footer-band").
												Body(
													app.P().
														Class("pf-c-login__main-footer-band-item").
														Body(
															app.Text("Want to know how it works? "), app.A().
																Href("https://github.com/JakWai01/airdrip").
																Body(
																	app.Text("Find out!"),
																),
														),
												),
										),
								),
							app.Footer().
								Class("pf-c-login__footer").
								Body(
									app.P().
										Body(
											app.Text("Airdrip is a peer-to-peer filesharing service. It uses WebRTC for the peer-to-peer connections, go-app in the frontend and a signaling-server hosted on Heroku."),
										),
									app.Ul().
										Class("pf-c-list pf-m-inline").
										Body(
											app.Li().
												Body(
													app.A().
														Href("https://github.com/JakWai01/airdrip#About").
														Body(
															app.Text("Documentation"),
														),
												),
											app.Li().
												Body(
													app.A().
														Href("https://github.com/JakWai01/airdrip/blob/main/LICENSE").
														Body(
															app.Text("License"),
														),
												),
											app.Li().
												Body(
													app.A().
														Href("https://github.com/JakWai01/airdrip").
														Body(
															app.Text("Source Code"),
														),
												),
										),
								),
						),
				),
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
