/*
 * plabel -- Brother p-touch label printer driver
 * Copyright (c) 2021-2022 
 */

package plabel

import (
	"fmt"
	"os"
	"time"
	"encoding/binary"
)

const (
	VERBOSE_TRACE = 4
	VERBOSE_DEBUG = 3
	VERBOSE_INFO  = 2
	VERBOSE_WARN  = 1

	LOOP_DELAY = 100 //ms

	COMMAND_MODE_ESCP 		= 0x00 //ESC/P mode (default)
	COMMAND_MODE_RASTER 	= 0x01 //Raster mode
	COMMAND_MODE_TEMPLATE = 0x03 //P-touch Template mode 

	PI_KIND 		= 0x02  // Media type
	PI_WIDTH 		= 0x04  // Media width
	PI_LENGTH 	= 0x08  // Media length
	PI_QUALITY 	= 0x40  // Priority given to print quality(Not used)
	PI_RECOVER 	= 0x80  // Printer recovery always on

	MEDIA_TYPE_NO_TAPE 				= 0x00
	MEDIA_TYPE_LAMINATED 			= 0x01
	MEDIA_TYPE_NON_LAMINATED 	= 0x03
	MEDIA_TYPE_HEAT_SHRINK 		= 0x11
	MEDIA_TYPE_INCOMPATIBLE		= 0xff

	DATA_LINE_PIXEL_WIDTH 	= 128
	DATA_LINE_BUFFER_LENGTH = 16
)

type Plabel struct {
	device *os.File
	active bool
	status_updated bool
	is_printing bool

	Verbose byte
	Simulate bool

	PrinterStatus PrinterStatus
	ModelInformation ModelInformation
	StatusCode byte
	InitalSettings bool

	MaxPrintingWidth uint16
}

func New() (self *Plabel) {
	self = new(Plabel)
	self.active = true
	self.Verbose = VERBOSE_DEBUG
	self.MaxPrintingWidth = 128
	return
}

func (self *Plabel) Open(device_path string) bool {
	var err error

	fmt.Printf("Using printer device: %s\n", device_path)
	self.device, err = os.OpenFile(device_path, os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR opening printer: %s\n", err)
		return false
	}
	return true
}

func (self *Plabel) Close() {
	self.device.Close()
}

func (self *Plabel) ProcessStatus() {
	fmt.Printf("ProcessStatus waiting for printer to send status: %t\n", self.active)
	for self.active {
		err := binary.Read(self.device, binary.LittleEndian, &self.PrinterStatus)

		if err != nil {
			//fmt.Printf(".")
			time.Sleep(LOOP_DELAY * time.Millisecond)
			continue
		}

		if !self.PrinterStatus.IsValid() {
			fmt.Fprintln(os.Stderr, "ERROR ProcessStatus invalid response from printer")
			time.Sleep(10 * time.Second)
		}

		if !self.InitalSettings {
			self.InitalSettings = true
			self.ModelInformation = *GetModelInformation(self.PrinterStatus.ModelCode)

			media_max_pixel := MediaWidthToMaxPixel(self.PrinterStatus.MediaWidth, self.ModelInformation.Resolution)
			self.MaxPrintingWidth = self.ModelInformation.PixelWidth
			if media_max_pixel < self.ModelInformation.PixelWidth {
				self.MaxPrintingWidth = media_max_pixel
			}
		}

		self.StatusCode = self.PrinterStatus.StatusCode
		self.status_updated = true
		
		if self.PrinterStatus.StatusCode == STATUS_PHASE_CHANGE {
			self.is_printing = self.PrinterStatus.PhaseType == PHASE_PRINTING
		}

		if self.Verbose >= VERBOSE_DEBUG {
			self.DisplayStatusVerbose()
		} else if self.Verbose >= VERBOSE_INFO {
			self.DisplayStatus()
		}
	}
}

func (self *Plabel) DisplayStatus() {
	fmt.Printf("Printer Status - status: %s, phase: %s\n", self.PrinterStatus.StatusDescription(), self.PrinterStatus.PhaseTypeDescription())
	if self.PrinterStatus.ErrorCode > 0 {
		fmt.Printf("Printer Status - ERROR: (%02x) %s\n", self.PrinterStatus.ErrorCode, self.PrinterStatus.ErrorDescription())
	}
}

func (self *Plabel) DisplayStatusVerbose() {
	fmt.Printf("Model.........: (%02X) %s\n", self.PrinterStatus.ModelCode, self.ModelInformation.ModelName);
	fmt.Printf("Error.........: (%04X) %s\n", self.PrinterStatus.ErrorCode, self.PrinterStatus.ErrorDescription());
	fmt.Printf("Status........: (%02X) %s\n", self.PrinterStatus.StatusCode, self.PrinterStatus.StatusDescription());
	fmt.Printf("Phase type....: %s, phase: %s\n", self.PrinterStatus.PhaseTypeDescription(), self.PrinterStatus.PhaseDescription());
	fmt.Printf("Notification..: %s\n", self.PrinterStatus.NotificationDescription());
	fmt.Printf("Media type....: (%d) %s\n", self.PrinterStatus.MediaType, self.PrinterStatus.MediaTypeDescription());
	fmt.Printf("Media width...: %d mm\n", self.PrinterStatus.MediaWidth);
	fmt.Printf("Media length..: %d mm\n", self.PrinterStatus.MediaLength);
	fmt.Printf("Media Color...: %s, text: %s\n\n", self.PrinterStatus.TapeColorDescription(), self.PrinterStatus.TextColorDescription());
}

func (self *Plabel) ShowInfo() {
	if self.PrinterStatus.ErrorCode > 0 {
		fmt.Printf("Printer Status - ERROR: (%02x) %s\n", self.PrinterStatus.ErrorCode, self.PrinterStatus.ErrorDescription())
	}

	fmt.Printf("\nPrinter:\n");
	fmt.Printf("Model.........: (%02X) %s\n", self.PrinterStatus.ModelCode, self.ModelInformation.ModelName);
	fmt.Printf("Pixel width...: %d\n", self.ModelInformation.PixelWidth);
	fmt.Printf("Resolution....: %d dpi\n", self.ModelInformation.Resolution);
	fmt.Printf("\nMedia:\n");
	fmt.Printf("Type..........: (%d) %s\n", self.PrinterStatus.MediaType, self.PrinterStatus.MediaTypeDescription());
	fmt.Printf("Width.........: %d mm\n", self.PrinterStatus.MediaWidth);
	fmt.Printf("Length........: %d mm\n", self.PrinterStatus.MediaLength);
	fmt.Printf("Color.........: %s, text: %s\n", self.PrinterStatus.TapeColorDescription(), self.PrinterStatus.TextColorDescription());
	fmt.Printf("Pixel width...: %d\n", MediaWidthToMaxPixel(self.PrinterStatus.MediaWidth, self.ModelInformation.Resolution));
	fmt.Printf("\nPrinting:\n");
	fmt.Printf("Max. width....: %d px\n", self.MaxPrintingWidth);
	fmt.Println()
}

func (self *Plabel) SendCommand(command []byte) {
	if self.Simulate {
		return
	}
	self.device.Write(command)
}

func (self *Plabel) Invalidate() {
	self.SendCommand(make([]byte, 100))
}

func (self *Plabel) Initialize() {
	self.SendCommand([]byte{0x1b, 0x40})
}

func (self *Plabel) RequestStatus() {
	self.ResetStatus()
	self.SendCommand([]byte{0x1b, 0x69, 0x53})
}

func (self *Plabel) SwitchDynamicCommandMode(mode byte) {
	self.SendCommand([]byte{0x1b, 0x69, 0x61, mode})
}

func (self *Plabel) SwitchRasterMode() {
	self.SwitchDynamicCommandMode(COMMAND_MODE_RASTER)
}

func (self *Plabel) SwitchEscpMode() {
	self.SwitchDynamicCommandMode(COMMAND_MODE_ESCP)
}

func (self *Plabel) SetPrintInformation(media_type byte, media_width byte, is_starting_page bool, raster_number uint32) {
	var valid_flag byte
	//var media_type byte
	//var media_width byte
	var media_length byte
	var starting_page byte
	var raster_number_0 byte
	var raster_number_1 byte
	var raster_number_2 byte
	var raster_number_3 byte

	valid_flag = PI_RECOVER
	if media_type > MEDIA_TYPE_NO_TAPE {
		valid_flag += PI_KIND
	}
	if media_width > 0 {
		valid_flag += PI_WIDTH
	}
	
	if is_starting_page {
		starting_page = 1
	}

	raster_number_0 = byte(raster_number & 0x000000ff)
	raster_number_1 = byte((raster_number >> 8) & 0x000000ff)
	raster_number_2 = byte((raster_number >> 16) & 0x000000ff)
	raster_number_3 = byte((raster_number >> 24) & 0x000000ff)

	self.SendCommand([]byte{0x1B, 0x69, 0x7A, valid_flag, media_type, media_width, media_length, raster_number_0, raster_number_1, raster_number_2, raster_number_3, starting_page, 0x00})
}

func (self *Plabel) SetCutMirror(cut bool, mirror bool) {
	var settings byte

	//Setst the cut at the beginning of the page. End cut is done anyway if no chain printing enabled
	if cut {
		settings |= (1 << 6)
	}
	if mirror {
		settings |= (1 << 7)
	}

	self.SendCommand([]byte{0x1B, 0x69, 0x4d, settings})
}

func (self *Plabel) SetAdvancedModeSettings(no_chain_printing bool, special_tape bool, no_buffer_clearing bool) {
	var settings byte

	//true = No chain printing(Feeding and cutting are performed after the last one is printed.)
	//false = Chain printing(Feeding and cutting are not performed after the last one is printed.)
	if no_chain_printing {
		settings |= (1 << 3)
	}

	//Labels are not cut when special tape is installed.
	//Special tape (no cutting) on/off
	if special_tape {
		settings |= (1 << 4)
	}

	//No buffer clearing when printing on/off
	if no_buffer_clearing {
		settings |= (1 << 7)
	}

	self.SendCommand([]byte{0x1B, 0x69, 0x4B, settings}) 
}

func (self *Plabel) SetFeedMargins(margin_dots uint16) {
	low_octet := byte(margin_dots & 0xff)
	high_octet := byte((margin_dots >> 8) & 0xff)
	self.SendCommand([]byte{0x1B, 0x69, 0x64, low_octet, high_octet})
}

func (self *Plabel) Print() {
	self.ResetStatus()
	self.SendCommand([]byte{0x0c})
	if self.Simulate {
		self.StatusCode = STATUS_PRINTING_COMPLETED
	}
}

func (self *Plabel) PrintAndFeed() {
	self.ResetStatus()
	self.SendCommand([]byte{0x1a})
	if self.Simulate {
		self.StatusCode = STATUS_PRINTING_COMPLETED
	}
}

func (self *Plabel) SetCompression() {
	self.SendCommand([]byte{0x4D, 0x02})
}

func (self *Plabel) SendZeroRasterGraphics() {
	self.SendCommand([]byte{0x5a})
}

func (self *Plabel) SendRasterGraphics(raster_data []byte) {
	if self.ModelInformation.UseCompression {
		self.SendRasterGraphicsCompressed(raster_data)
	} else {
		self.SendRasterGraphicsUncompressed(raster_data)
	}
}

func (self *Plabel) SendRasterGraphicsUncompressed(raster_data []byte) {
	rb := make([]byte, 19)
	rb[0] = 0x47
	rb[1] = 0x10
	rb[2] = 0x00
	copy(rb[3:19], raster_data)

	if self.Verbose >= VERBOSE_TRACE {
		for _, octet := range rb {
			fmt.Printf("%02X ", octet)
		}
	}

	if self.Verbose >= VERBOSE_DEBUG {
		self.DisplayRasterGraphics(raster_data)
	}

	self.SendCommand(rb)
}

func (self *Plabel) SendRasterGraphicsCompressed(raster_data []byte) {
	rb := make([]byte, 20)
	rb[0] = 0x47
	rb[1] = 0x11
	rb[2] = 0x00
	rb[3] = 0x10
	copy(rb[4:20], raster_data)

	if self.Verbose >= VERBOSE_TRACE {
		for _, octet := range rb {
			fmt.Printf("%02X ", octet)
		}
	}

	if self.Verbose >= VERBOSE_DEBUG {
		self.DisplayRasterGraphics(raster_data)
	}

	self.SendCommand(rb)
}

func (self *Plabel) DisplayRasterGraphics(raster_data []byte) {
	for octet := len(raster_data)-1; octet >= 0; octet-- {
		for i := 0; i < 8; i++ {
			if (raster_data[octet] & (1 << i)) != 0 {
				fmt.Print("█")
			} else {
				fmt.Print(" ")
			}
		}
	}

	fmt.Print("\n")
}

func (self *Plabel) TimeMilliseconds() int64 {
	return time.Now().UnixNano() / 1000000  
}

func (self *Plabel) ResetStatus() {
	self.status_updated = false
}

func (self *Plabel) WaitForPrinterStatus(timeout uint16) bool {
	start_time := self.TimeMilliseconds()
	
	for self.active && self.TimeMilliseconds() < start_time + int64(timeout)  {
		if (self.status_updated) {
			return true
		}
		time.Sleep(LOOP_DELAY * time.Millisecond)
	}
	return false
}

func (self *Plabel) WaitForPrintingCompleted(timeout uint16) bool {
	start_time := self.TimeMilliseconds()
	
	for self.active && self.TimeMilliseconds() < start_time + int64(timeout) {
		if self.status_updated && !self.is_printing {
			return true
		}
		time.Sleep(LOOP_DELAY * time.Millisecond)
	}
	return false
}
