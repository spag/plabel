/*
 * plabel -- Brother p-touch label printer driver
 * Copyright (c) 2021-2022 
 */

package main

import (
  "syscall"
  "fmt"
  "os"
	"os/signal"
	"flag"
  "image/color"
	"image"
  _ "image/gif"
	_ "image/png"
	_ "image/jpeg"
	"plabel"
)

const (
  PROGRAM_NAME = "plabel"
  PROGRAM_VERSION = "0.0.1"
)

type Settings struct {
	pid_file string
  printer_device string
	image_file string
	black_threshold uint
  verbose uint
  simulate bool
  show_info bool
  batch_mode bool
  no_front_cut bool
  mirror bool
}

func CopyrightMessage() {
  fmt.Fprintf(os.Stderr, "%s - Version %s\nCopyright (c) 2021-2022, sipomat ltd.\n\n", PROGRAM_NAME, PROGRAM_VERSION)
}

func ReadCommandLineParameters(settings *Settings) {
  flag.Usage = func() {
    CopyrightMessage()
    fmt.Fprintf(os.Stderr, "Usage: %s <parameters>\n\n", os.Args[0])
    //flag.PrintDefaults()
		fmt.Println(`  -h, --help                  Print usage information
      --pid-file <file>       Save Process-ID to file
  -p, --printer <device>      Printer device
  -f, --file <file>           Print from file (png)
  -t, --threshold <0-255>     Threshold at which a pixel is determined black
  -v, --verbose <0-4>         Verbosity level
  -b, --batch-mode            Chain printing without feeding and end cut
  -m, --mirror                Mirror output
  -n, --no-cut                No front cut
  -s, --simulate              Just simulate, do not print.
  -i, --info                  Get printer information
		`)
    os.Exit(1)
  }

	flag.StringVar(&settings.printer_device, "p", "/dev/usb/lp0", "Printer")
  flag.StringVar(&settings.printer_device, "printer", "/dev/usb/lp0", "Printer")
	flag.StringVar(&settings.pid_file, "pid-file", "", "Save Process-ID to file")
	flag.StringVar(&settings.image_file, "f", "", "Prit from file")
	flag.StringVar(&settings.image_file, "file", "", "Prit from file")
	flag.UintVar(&settings.black_threshold, "t", 182, "Threshold at which a pixel is determined black")
	flag.UintVar(&settings.black_threshold, "threshold", 182, "Threshold at which a pixel is determined black")
  flag.BoolVar(&settings.simulate, "s", false, "simulate")
  flag.BoolVar(&settings.simulate, "simulate", false, "simulate")
  flag.UintVar(&settings.verbose, "v", 2, "Verbosity level")
	flag.UintVar(&settings.verbose, "verbose", 2, "Verbosity level")
  flag.BoolVar(&settings.show_info, "i", false, "printer info")
  flag.BoolVar(&settings.show_info, "info", false, "printer info")
  flag.BoolVar(&settings.batch_mode, "b", false, "batch_mode printing")
  flag.BoolVar(&settings.batch_mode, "batch-mode", false, "batch_mode printing")
  flag.BoolVar(&settings.no_front_cut, "n", false, "no-cut printing")
  flag.BoolVar(&settings.no_front_cut, "no-cut", false, "no-cut printing")
  flag.BoolVar(&settings.mirror, "m", false, "mirror printing")
  flag.BoolVar(&settings.mirror, "mirror", false, "mirror printing")
  flag.Parse()
}

func CreatePIDFile(pid_file_name string) error {
  pid_file, err := os.Create(pid_file_name)
  if err != nil {
    return fmt.Errorf("error creating PID file: %s\n", err)
  }

  defer pid_file.Close()

  if _, err = fmt.Fprintf(pid_file, "%d", os.Getpid()) ;  err != nil {
    return fmt.Errorf("error creating PID file: %s\n", err)
  }

  return nil
}

func SendFile(printer *plabel.Plabel, file_name string, threshold byte) bool {
  var index byte
  var height int
  var length int
  var padding byte
  var margin int

  fd, err := os.Open(file_name)

	if err != nil {
    fmt.Fprintf(os.Stderr, "ERROR opeing file: %s\n", err)
		return false
	}

	defer fd.Close()

	img, format, err := image.Decode(fd)
	if err != nil {
    fmt.Fprintf(os.Stderr, "ERROR decoding file: %s\n", err)
		return false
	}

  length = img.Bounds().Max.X - img.Bounds().Min.X

  height = img.Bounds().Max.Y - img.Bounds().Min.Y
  min_y := img.Bounds().Min.Y
  max_y := img.Bounds().Max.Y

  if height > int(printer.MaxPrintingWidth) {
    margin = (height - int(printer.MaxPrintingWidth)) / 2
    min_y = img.Bounds().Min.Y + margin
    max_y = img.Bounds().Max.Y - margin
    height = max_y - min_y
  }

  padding = byte((plabel.DATA_LINE_PIXEL_WIDTH-height)/2)

  fmt.Printf("SendFile - type: %s, height: %d px, length: %d px, margins: %d px, padding: %d px, printing width: %d\n", format, img.Bounds().Max.Y - img.Bounds().Min.Y, length, margin, padding, printer.MaxPrintingWidth)

	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
    index = 0
    line := make([]byte, plabel.DATA_LINE_BUFFER_LENGTH)
		for y := min_y; y < max_y; y++ {
			c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			if c.Y < threshold {
        octet := (index + padding)/8
        line[octet] |= (1 << (7-((index + padding) % 8)))

			}     
      index += 1
		}

    printer.SendRasterGraphics(line)
	}

  return true
}

func main() {
  var settings Settings
	var run_process bool = true

	ReadCommandLineParameters(&settings)

  if len(settings.pid_file) > 0 {
    if err := CreatePIDFile(settings.pid_file) ; err != nil {
      fmt.Fprintln(os.Stderr, PROGRAM_NAME, "ERROR creating PID file: ", err)
      os.Exit(1)
    } else {
      fmt.Println(PROGRAM_NAME, " created PID file: ", settings.pid_file)
    }
	}

  printer := 	plabel.New()
  printer.Simulate = settings.simulate
  printer.Verbose = byte(settings.verbose)

  if !printer.Open(settings.printer_device) {
    fmt.Fprintln(os.Stderr, PROGRAM_NAME, "ERROR opening printer: ", settings.printer_device)
    if (!settings.simulate) {
      os.Exit(1)
    }
  }

  defer printer.Close()
  go printer.ProcessStatus()

  printer.WaitForPrinterStatus(1000)

  printer.Invalidate()
  printer.Initialize()
  printer.RequestStatus()

  printer.WaitForPrinterStatus(1000)

  if (settings.show_info) {
    printer.ShowInfo()
  }
  
  if len(settings.image_file) > 0 {
    printer.SwitchRasterMode()
    printer.SetCutMirror(!settings.no_front_cut, settings.mirror)
    printer.SetCompression()
    //printer.SetFeedMargins(16)
    //printer.SetPrintInformation(0, 0, false, 0)
    if SendFile(printer, settings.image_file, byte(settings.black_threshold)) {
      if settings.batch_mode {
        printer.Print()
      } else {
        printer.PrintAndFeed()
      }
    }

    printer.WaitForPrintingCompleted(10000)
  }

  run_process = false

	signal_channel := make(chan os.Signal)
  signal.Notify(signal_channel, syscall.SIGINT)
  signal.Notify(signal_channel, syscall.SIGTERM)
  signal.Notify(signal_channel, syscall.SIGHUP)
	
  for run_process {    
    select {
    case signal_rec := <-signal_channel:
      if signal_rec == syscall.SIGHUP {
        fmt.Println(PROGRAM_NAME, " SIGNAL - received hangup signal: ", signal_rec)
      } else {
        fmt.Println(PROGRAM_NAME, " SIGNAL - received signal: ", signal_rec)
        run_process = false
      }
    }
  }

  fmt.Println(PROGRAM_NAME, " ended")
}
