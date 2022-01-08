/*
 * plabel -- Brother p-touch label printer driver
 * Copyright (c) 2021-2022 
 */

package plabel

const (
	STATUS_PRINTING_COMPLETED = 0x01
	STATUS_ERROR 							= 0x02
	STATUS_PHASE_CHANGE 			= 0x06

	PHASE_EDITING		= 0x00
	PHASE_PRINTING 	= 0x01
)

type PrinterStatus struct {
	PrintHeadMark byte
	Size byte
	ManufacturerCode byte
	SeriesCode byte
	ModelCode byte
	CountryCode byte
	Reserved1 uint16
	ErrorCode uint16
	MediaWidth byte
	MediaType byte
	Ncol byte
	Fonts byte
	JpFonts byte
	Mode byte
	Density byte
	MediaLength byte
  StatusCode byte
	PhaseType byte
	PhaseNumber uint16
	NotificationCode byte
	ExpansionArea byte
	TapeColor byte
	TextColor byte
	HwSetting uint32
	Reserved2 uint16
}

func (self *PrinterStatus) IsValid() (bool) {
	return self.PrintHeadMark == 0x80 && self.Size == 0x20 && self.ManufacturerCode == 0x42
}

func (self *PrinterStatus) ErrorDescription() (ed string) {
	error_map := map[uint16]string{
		0x0001: "No media",
		0x0004: "Cutter jam",
		0x0008: "Weak batteries",
		0x0040: "High-voltage adapter",
		0x0100: "Wrong media",
		0x1000: "Cover open",
		0x2000: "Overheating",
	}

	for bitmask, description := range error_map {
		if self.ErrorCode & bitmask > 0 {
			ed += description + " "
		}
	}

	return
}

func (self *PrinterStatus) StatusDescription() (ed string) {
	status_map := map[byte]string{
		0x00: "Reply to status request",
		0x01: "Printing completed",
		0x02: "Error occurred",
		0x03: "Exit IF mode",
		0x04: "Turned off",
		0x05: "Notification",
		0x06: "Phase change",
	}

	if description, ok := status_map[self.StatusCode]; ok {
    return description
	}

	return "Unknown"
}

func (self *PrinterStatus) PhaseTypeDescription() (ed string) {
	phase_map := map[byte]string{
		0x00: "Editing",
		0x01: "Printing",
	}

	if description, ok := phase_map[self.PhaseType]; ok {
    return description
	}

	return "Unknown"
}

func (self *PrinterStatus) PhaseDescription() (ed string) {
	phase_map := map[uint16]string{
		0x0000: "Editing",
		0x0001: "Printing",
		0x000a: "Not used",
		0x0014: "Cover open while receiving",
		0x0019: "Not used",
	}

	if description, ok := phase_map[self.PhaseNumber]; ok {
    return description
	}

	return "Unknown"
}

func (self *PrinterStatus) NotificationDescription() (ed string) {
	note_map := map[byte]string{
		0x00: "Not available",
		0x01: "Cover open",
		0x02: "Cover closed",
	}

	if description, ok := note_map[self.NotificationCode]; ok {
    return description
	}

	return "Unknown"
}

func (self *PrinterStatus) MediaTypeDescription() (ed string) {
	media_map := map[byte]string{
		0x00: "No media",
		0x01: "Laminated tape",
		0x03: "Non-laminated tape",
		0x11: "Heat-Shrink Tube",
		0xff: "Incompatible tape",
	}

	if description, ok := media_map[self.MediaType]; ok {
    return description
	}

	return "Unknown"
}

func (self *PrinterStatus) TapeColorDescription() (ed string) {
	color_map := map[byte]string{
		0x01: "White",
		0x02: "Other",
		0x03: "Clear",
		0x04: "Red",
		0x05: "Blue",
		0x06: "Yellow",
		0x07: "Green",
		0x08: "Black",
		0x09: "Clear(White text)",
		0x20: "Matte White",
		0x21: "Matte Clear",
		0x22: "Matte Silver",
		0x23: "Satin Gold",
		0x24: "Satin Silver",
		0x30: "Blue(D)",
		0x31: "Red(D)",
		0x40: "Fluorescent Orange",
		0x41: "Fluorescent Yellow",
		0x50: "Berry Pink(S)",
		0x51: "Light Gray(S)",
		0x52: "Lime Green(S)",
		0x60: "Yellow(F)",
		0x61: "Pink(F)",
		0x62: "Blue(F)",
		0x70: "White(Heat-shrink Tube)",
		0x90: "White(Flex. ID)",
		0x91: "Yellow(Flex. ID)",
		0xf0: "Clearning",
		0xf1: "Stencil",
		0xff: "Incompatible",
	}

	if description, ok := color_map[self.TapeColor]; ok {
    return description
	}

	return "Unknown"
}

func (self *PrinterStatus) TextColorDescription() (ed string) {
	color_map := map[byte]string{
		0x01: "White",
		0x02: "Other",  
		0x03: "Clear",   //not specified
		0x04: "Red",
		0x05: "Blue",
		0x06: "Yellow",  //not specified
		0x07: "Green",   //not specified
		0x08: "Black",
		0x09: "Clear(White text)", //not specified
		0x20: "Matte White",  //not specified
		0x21: "Matte Clear",  //not specified
		0x22: "Matte Silver", //not specified
		0x23: "Satin Gold",   //not specified
		0x24: "Satin Silver",  //not specified
		0x30: "Blue(D)",  //not specified
		0x31: "Red(D)",  //not specified
		0x40: "Fluorescent Orange",  //not specified
		0x41: "Fluorescent Yellow",  //not specified
		0x50: "Berry Pink(S)",  //not specified
		0x51: "Light Gray(S)",  //not specified
		0x52: "Lime Green(S)",  //not specified
		0x60: "Yellow(F)",  //not specified
		0x61: "Pink(F)",  //not specified
		0x62: "Blue(F)",
		0x70: "White(Heat-shrink Tube)", //not specified
		0x90: "White(Flex. ID)",  //not specified
		0x91: "Yellow(Flex. ID)", //not specified
		0xf0: "Clearning", 
		0xf1: "Stencil",
		0xff: "Incompatible",
	}

	if description, ok := color_map[self.TextColor]; ok {
    return description
	}

	return "Unknown"
}

