/*
 * plabel -- Brother p-touch label printer driver
 * Copyright (c) 2021-2022 
 */

package plabel

func MediaWidthToMaxPixel(media_width byte, dpi uint16) uint16 {
	media_map := map[byte]uint16{
		4: 0,
		6: 0,
		12: 70,
		18: 112,
		24: 128,
	}

	if dpi == 180 {
		if pixel_width, ok := media_map[media_width]; ok {
			return pixel_width
		}
	}

	return 0
}
