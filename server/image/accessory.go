package image

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pawr-project/Pawnimals/server/color"
	"github.com/pawr-project/Pawnimals/server/rand"
	"github.com/pawr-project/Pawnimals/server/spc"
)

// Accessories - represents accessories for natricon
type Accessories struct {
	BodyColor         color.RGB
	HairColor         color.RGB
	FaceAsset         Asset
	HairAsset         *Asset
	MouthAsset        *Asset
	EyeAsset          *Asset
	BackHairAsset     *Asset
	BodyOutlineAsset  *Asset
	HairOutlineAsset  *Asset
	MouthOutlineAsset *Asset
	BadgeAsset        *Asset
	OutlineColor      color.RGB
}

// Hex string regex
const hexRegexStr = "^[0-9a-fA-F]+$"

var hexRegex = regexp.MustCompile(hexRegexStr)

// GetSpecificNatricon - Return Accessories object with specific parameters
func GetSpecificNatricon(badgeType spc.BadgeType, outline bool, outlineColor *color.RGB, bodyColor *color.RGB, hairColor *color.RGB, faceAsset int, hairAsset int, mouthAsset int, eyeAsset int) Accessories {
	var accessories = Accessories{}

	// Set colors
	accessories.BodyColor = *bodyColor
	accessories.HairColor = *hairColor

	// Assets
	accessories.FaceAsset = GetFaceAssetWithID(faceAsset)
	//accessories.HairAsset = GetHairAssetWithID(hairAsset)
	//accessories.BackHairAsset = GetBackHairAsset(accessories.HairAsset)

	// Get badge
	if badgeType != "" && badgeType != spc.BTNone {
		accessories.BadgeAsset = GetBadgeAsset(accessories.FaceAsset, badgeType)
	}

	// Eyes and mouth
	//accessories.MouthAsset = GetMouthAssetWithID(mouthAsset)
	//accessories.EyeAsset = GetEyeAssetWithID(eyeAsset)

	// Get outlines
	/*if outline {
		accessories.BodyOutlineAsset = GetBodyOutlineAsset(accessories.FaceAsset)
		accessories.HairOutlineAsset = GetHairOutlineAsset(accessories.HairAsset)
		accessories.MouthOutlineAsset = GetMouthOutlineAsset(accessories.MouthAsset)
		if outlineColor != nil {
			accessories.OutlineColor = *outlineColor
		} else {
			accessories.OutlineColor = color.RGB{R: 0, G: 0, B: 0}
		}
	}*/

	return accessories
}

// GetAccessoriesForHash - Return Accessories object based on 64-character hex string
func GetAccessoriesForHash(hash string, badgeType spc.BadgeType, outline bool, outlineColor *color.RGB) (Accessories, error) {
	var err error
	if len(hash) != 64 {
		return Accessories{}, errors.New("Invalid hash")
	}
	// Validate is a hex string
	if !hexRegex.MatchString(hash) {
		return Accessories{}, errors.New("Invalid hash")
	}

	// Create empty Accessories object
	var accessories = Accessories{}
	// Body color uses first 12 digits of hash as seed
	accessories.BodyColor, err = GetBodyColor(hash[0:16])
	if err != nil {
		return Accessories{}, err
	}

	//// Get hair color
	accessories.HairColor, err = GetHairColor(accessories.BodyColor, hash[16:26], hash[26:30], hash[30:34])

	// Get body and hair illustrations
	accessories.FaceAsset, err = GetFaceAsset(hash[34:40])
	//accessories.HairAsset, err = GetHairAsset(hash[40:46], &accessories.FaceAsset)
	//accessories.BackHairAsset = GetBackHairAsset(accessories.HairAsset)

	/*
	// Get badge
	if badgeType != "" && badgeType != spc.BTNone {
		accessories.BadgeAsset = GetBadgeAsset(accessories.FaceAsset, badgeType)
	}*/

	/*
	// Get mouth and eyes
	targetSex := Neutral
	if accessories.FaceAsset.Sex != Neutral {
		targetSex = accessories.FaceAsset.Sex
	} else if accessories.HairAsset.Sex != Neutral {
		targetSex = accessories.HairAsset.Sex
	}
	accessories.MouthAsset, err = GetMouthAsset(hash[46:55], targetSex, accessories.BodyColor.PerceivedBrightness())
	if targetSex == Neutral && accessories.MouthAsset.Sex != Neutral {
		targetSex = accessories.MouthAsset.Sex
	}
	accessories.EyeAsset, err = GetEyeAsset(hash[55:64], targetSex, accessories.BodyColor.PerceivedBrightness())

	// Get outlines
	if outline {
		accessories.BodyOutlineAsset = GetBodyOutlineAsset(accessories.FaceAsset)
		accessories.HairOutlineAsset = GetHairOutlineAsset(accessories.HairAsset)
		accessories.MouthOutlineAsset = GetMouthOutlineAsset(accessories.MouthAsset)
		if outlineColor != nil {
			accessories.OutlineColor = *outlineColor
		} else {
			accessories.OutlineColor = color.RGB{R: 0, G: 0, B: 0}
		}
	}
	*/
	return accessories, nil
}

// GetFaceAsset - return body illustration to use with given entropy
func GetFaceAsset(entropy string) (Asset, error) {
	// Get detemrinistic RNG
	randSeed, err := strconv.ParseInt(entropy, 16, 64)
	if err != nil {
		return Asset{}, err
	}

	r := rand.Init()
	r.Seed(uint32(randSeed))
	faceIndex := r.Int31n(int32(GetAssets().GetNFaceAssets()))

	return GetAssets().GetFaceAssets()[faceIndex], nil
}

// GetFaceAssetWithID - return body illustration with given ID
func GetFaceAssetWithID(id int) Asset {
	for _, ba := range GetAssets().GetFaceAssets() {
		baid, err := strconv.Atoi(strings.Split(ba.FileName, "_")[0])
		if err != nil {
			baid, err = strconv.Atoi(strings.Split(ba.FileName, ".")[0])
			if err != nil {
				continue
			}
		}
		if baid == id {
			return ba
		}
	}
	return GetAssets().GetFaceAssets()[0]
}

// GetBodyOutlineAsset - return body outline illustration for a given body asset
func GetBodyOutlineAsset(faceAsset Asset) *Asset {
	for _, ba := range GetAssets().GetBodyOutlineAssets() {
		if ba.FileName == faceAsset.FileName {
			return &ba
		}
	}
	return nil
}

// GetBadgeAsset - return badge asset for a particular body
func GetBadgeAsset(bodyAsset Asset, btype spc.BadgeType) *Asset {
	identifier, _ := strconv.Atoi(strings.Split(bodyAsset.FileName, "_")[0])
	searchStr := fmt.Sprintf("b%d", identifier)
	for _, v := range GetAssets().GetBadgeAssets(btype) {
		if strings.Contains(v.FileName, searchStr) {
			return &v
		}
	}
	return nil
}

// GetHairAsset - return hair illustration to use with given entropy
func GetHairAsset(entropy string, bodyAsset *Asset) *Asset {
	// Get detemrinistic RNG
	randSeed, err := strconv.ParseInt(entropy, 16, 64)
	if err != nil {
		return nil
	}

	hairAssetOptions := GetAssets().GetHairAssets(bodyAsset.Sex)

	r := rand.Init()
	r.Seed(uint32(randSeed))
	hairIndex := r.Int31n(int32(len(hairAssetOptions)))

	if(len(hairAssetOptions) > 0) {
		return &hairAssetOptions[hairIndex]
	}
	return nil
}

// GetHairAssetWithID - return body illustration with given ID
func GetHairAssetWithID(id int) *Asset {
	for _, ha := range GetAssets().GetHairAssets(Neutral) {
		haid, err := strconv.Atoi(strings.Split(ha.FileName, "_")[0])
		if err != nil {
			haid, err = strconv.Atoi(strings.Split(ha.FileName, ".")[0])
			if err != nil {
				continue
			}
		}
		if haid == id {
			return &ha
		}
	}
	return &GetAssets().GetHairAssets(Neutral)[0]
}

// GetBackHairAsset - return back hair illustration for a given hair asset
func GetBackHairAsset(hairAsset Asset) *Asset {
	for _, ba := range GetAssets().GetBackHairAssets() {
		if ba.FileName == hairAsset.FileName {
			return &ba
		}
	}
	return nil
}

// GetHairOutlineAsset - return hair outline illustration for a given hair asset
func GetHairOutlineAsset(hairAsset Asset) *Asset {
	for _, ba := range GetAssets().GetHairOutlineAssets() {
		if ba.FileName == hairAsset.FileName {
			return &ba
		}
	}
	return nil
}

// GetEyeAsset - return hair illustration to use with given entropy
func GetEyeAsset(entropy string, sex Sex, luminosity float64) *Asset {
	// Get detemrinistic RNG
	randSeed, err := strconv.ParseInt(entropy, 16, 64)
	if err != nil {
		return nil
	}

	eyeAssetOptions := GetAssets().GetEyeAssets(sex, luminosity)

	r := rand.Init()
	r.Seed(uint32(randSeed))
	eyeIndex := r.Int31n(int32(len(eyeAssetOptions)))

	if(len(eyeAssetOptions) > 0) {
		return &eyeAssetOptions[eyeIndex]
	}
	return nil
}

// GetEyeAssetWithID - return eye illustration with given ID
func GetEyeAssetWithID(id int) *Asset {
	for _, ba := range GetAssets().GetEyeAssets(Neutral, 100) {
		baid, err := strconv.Atoi(strings.Split(ba.FileName, "_")[0])
		if err != nil {
			baid, err = strconv.Atoi(strings.Split(ba.FileName, ".")[0])
			if err != nil {
				continue
			}
		}
		if baid == id {
			return &ba
		}
	}
	return &GetAssets().GetEyeAssets(Neutral, 100)[0]
}

// GetEyeAsset - return hair illustration to use with given entropy
func GetMouthAsset(entropy string, sex Sex, luminosity float64) *Asset {
	// Get detemrinistic RNG
	randSeed, err := strconv.ParseInt(entropy, 16, 64)
	if err != nil {
		return nil
	}

	mouthAssetOptions := GetAssets().GetMouthAssets(sex, luminosity)

	r := rand.Init()
	r.Seed(uint32(randSeed))
	mouthIndex := r.Int31n(int32(len(mouthAssetOptions)))

	if(len(mouthAssetOptions) > 0) {
		return &mouthAssetOptions[mouthIndex]
	}
	return nil
}

// GetMouthAssetWithID - return mouth illustration with given ID
func GetMouthAssetWithID(id int) *Asset {
	for _, ba := range GetAssets().GetMouthAssets(Neutral, 100) {
		baid, err := strconv.Atoi(strings.Split(ba.FileName, "_")[0])
		if err != nil {
			baid, err = strconv.Atoi(strings.Split(ba.FileName, ".")[0])
			if err != nil {
				continue
			}
		}
		if baid == id {
			return &ba
		}
	}
	return &GetAssets().GetMouthAssets(Neutral, 100)[0]
}

// GetMouthOutlineAsset - return mouth outline illustration for a given mouth asset
func GetMouthOutlineAsset(mouthAsset Asset) *Asset {
	for _, ba := range GetAssets().GetMouthOutlineAssets() {
		if ba.FileName == mouthAsset.FileName {
			return &ba
		}
	}
	return nil
}
