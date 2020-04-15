package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/appditto/natricon/color"
	"github.com/appditto/natricon/image"
	"github.com/appditto/natricon/nano"
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
)

var seed *string

func getRandomSvg(c *gin.Context) {
	var err error

	address := nano.GenerateAddress()
	sha256 := nano.AddressSha256(address, *seed)

	accessories, err := image.GetAccessoriesForHash(sha256)
	if err != nil {
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}
	bodyHsv := accessories.BodyColor.ToHSV()
	hairHsv := accessories.HairColor.ToHSV()
	deltaHsv := color.HSV{}
	deltaHsv.H = hairHsv.H - bodyHsv.H
	deltaHsv.S = hairHsv.S - bodyHsv.S
	deltaHsv.V = hairHsv.V - bodyHsv.V
	svg, err := image.CombineSVG(accessories)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error occured")
		return
	}
	c.Data(200, "image/svg+xml; charset=utf-8", svg)
}

func getRandom(c *gin.Context) {
	var err error

	address := nano.GenerateAddress()
	sha256 := nano.AddressSha256(address, *seed)

	accessories, err := image.GetAccessoriesForHash(sha256)
	if err != nil {
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}
	bodyHsv := accessories.BodyColor.ToHSV()
	hairHsv := accessories.HairColor.ToHSV()
	deltaHsv := color.HSV{}
	deltaHsv.H = hairHsv.H - bodyHsv.H
	deltaHsv.S = hairHsv.S - bodyHsv.S
	deltaHsv.V = hairHsv.V - bodyHsv.V
	svg, err := image.CombineSVG(accessories)
	var svgStr string
	if err != nil {
		svgStr = "Error"
	} else {
		svgStr = string(svg)
	}
	c.JSON(200, gin.H{
		"bodyColor": accessories.BodyColor.ToHTML(),
		"hairColor": accessories.HairColor.ToHTML(),
		"hash":      sha256,
		"bodyH":     int16(bodyHsv.H),
		"bodyS":     int16(bodyHsv.S * 100.0),
		"bodyV":     int16(bodyHsv.V * 100.0),
		"hairH":     int16(hairHsv.H),
		"hairS":     int16(hairHsv.S * 100.0),
		"hairV":     int16(hairHsv.V * 100.0),
		"deltaH":    int16(deltaHsv.H),
		"deltaS":    int16(deltaHsv.S * 100.0),
		"deltaV":    int16(deltaHsv.V * 100.0),
		"address":   address,
		"svg":       svgStr,
	})
	/*newHTML := strings.Replace(testhtml, "#000", "#"+accessories.HairColor.ToHTML(), -1)
	newHTML = strings.Replace(newHTML, "#FFF", "#"+accessories.BodyColor.ToHTML(), -1)
	newHTML = strings.Replace(newHTML, "address_1", address, -1)
	c.Data(200, "text/html; charset=utf-8", []byte(newHTML))*/
}

func getNatricon(c *gin.Context) {
	var err error

	address := c.Query("address")
	// valid := nano.ValidateAddress(address)
	// if !valid {
	// c.String(http.StatusBadRequest, "Invalid address")
	// return
	// }
	sha256 := nano.AddressSha256(address, *seed)

	accessories, err := image.GetAccessoriesForHash(sha256)
	if err != nil {
		c.String(http.StatusInternalServerError, "%s", err.Error())
		return
	}

	bodyHsv := accessories.BodyColor.ToHSV()
	hairHsv := accessories.HairColor.ToHSV()
	deltaHsv := color.HSV{}
	deltaHsv.H = hairHsv.H - bodyHsv.H
	deltaHsv.S = hairHsv.S - bodyHsv.S
	deltaHsv.V = hairHsv.V - bodyHsv.V
	c.JSON(200, gin.H{
		"bodyColor": accessories.BodyColor.ToHTML(),
		"hairColor": accessories.HairColor.ToHTML(),
		"hash":      sha256,
		"bodyH":     int16(bodyHsv.H),
		"bodyS":     int16(bodyHsv.S * 100.0),
		"bodyV":     int16(bodyHsv.V * 100.0),
		"hairH":     int16(hairHsv.H),
		"hairS":     int16(hairHsv.S * 100.0),
		"hairV":     int16(hairHsv.V * 100.0),
		"deltaH":    int16(deltaHsv.H),
		"deltaS":    int16(deltaHsv.S * 100.0),
		"deltaV":    int16(deltaHsv.V * 100.0),
		"address":   address,
	})
}

func main() {
	// Parse server options
	loadFiles := flag.Bool("load-files", false, "Print assets as GO arrays")
	serverHost := flag.String("host", "127.0.0.1", "Host to listen on")
	serverPort := flag.Int("port", 8080, "Port to listen on")
	seed = flag.String("seed", "1234567890", "Seed to use for icon generation")
	flag.Parse()

	if *loadFiles {
		LoadAssetsToArray()
		return
	}

	// Setup router
	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/natricon", getNatricon)
	router.GET("/random", getRandom)
	router.GET("/randomsvg", getRandomSvg)

	// Run on 8080
	router.Run(fmt.Sprintf("%s:%d", *serverHost, *serverPort))
}
