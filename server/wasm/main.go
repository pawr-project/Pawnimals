package main

import (
	"encoding/base64"
	"fmt"
	"syscall/js"

	"github.com/appditto/natricon/server/color"
	"github.com/appditto/natricon/server/image"
	"github.com/appditto/natricon/server/nano"
)

var seed string = "123456789"

func getNatriconStr(this js.Value, inputs []js.Value) interface{} {
	message := inputs[0].String()
	if !nano.ValidateAddress(message) {
		return js.Null()
	}
	sha256 := nano.AddressSha256(message, seed)

	accessories, err := image.GetAccessoriesForHash(sha256)
	if err != nil {
		return js.Null()
	}
	bodyHsv := accessories.BodyColor.ToHSV()
	hairHsv := accessories.HairColor.ToHSV()
	deltaHsv := color.HSV{}
	deltaHsv.H = hairHsv.H - bodyHsv.H
	deltaHsv.S = hairHsv.S - bodyHsv.S
	deltaHsv.V = hairHsv.V - bodyHsv.V
	svg, err := image.CombineSVG(accessories)
	if err != nil {
		return js.Null()
	}
	return js.ValueOf(fmt.Sprintf("data:image/svg+xml;base64,%s", base64.StdEncoding.EncodeToString(svg)))
}

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("getNatriconStr", js.FuncOf(getNatriconStr))
	<-c
}