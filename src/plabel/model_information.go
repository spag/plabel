/*
 * plabel -- Brother p-touch label printer driver
 * Copyright (c) 2021-2022 
 */

package plabel

const (
	PRINTER_P1230PC = 0x59
	PRINTER_H500		= 0x64
	PRINTER_E500		= 0x65
	PRINTER_P700		= 0x67
)

type ModelInformation struct {
	ModelCode byte
	IsValid bool
	ModelName string
	PixelWidth uint16
	Resolution uint16
	UseCompression bool
	MinTapeWidth byte
	MaxTapeWidth byte
}

func GetModelInformation(model_code byte) (*ModelInformation) {
	return new(ModelInformation).GetModelInformation(model_code)
}

func (self *ModelInformation) GetModelInformation(model_code byte) *ModelInformation {
	model_map := map[byte]ModelInformation{
		PRINTER_P1230PC: 	ModelInformation{PRINTER_P1230PC, true, "PT-P1230PC", 64, 180, false, 4, 12},
		PRINTER_H500: 		ModelInformation{PRINTER_H500, 		true, "PT-H500", 		128, 180, true, 4, 24},
		PRINTER_E500: 		ModelInformation{PRINTER_E500, 		true, "PT-E500", 		128, 180, true, 4, 24},
		PRINTER_P700: 		ModelInformation{PRINTER_P700, 		true, "PT-P700", 		128, 180, true, 4, 24},
	}

	if model_information, ok := model_map[model_code]; ok {
    return &model_information
	}

	return &ModelInformation{ModelCode: model_code, ModelName: "Unknown"}
}
