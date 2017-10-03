// This program is a derivation of my previous program saltAndPepper.
// I made it to improve a colored picture of mine that had some scratches.
// It takes the color at the upper left corner and treats that as a key,
// then it scans the entire picture replacing the key with a neighbor color.

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"sort"
)

type brightness uint32    // we sort based on brightness.
type img image.Image      // just an alias.
type colors []color.Color // so that I can easily declare a pointer to an array.

var keyColor color.Color // this is the color to be replaced.

// calcBrightness returns an arbitrary value that roughly represents the brightness.
func calcBrightness(c color.Color) brightness {
	r, g, b, _ := c.RGBA()       // break down into color components.
	return brightness(r + g + b) // simply sum the brightness of each component.
}

// getColor returns the color at (x,y) coordinate respecting the image boundaries.
// ok is false if (x,y) is out of img.
func getColor(img img, x, y int) (c color.Color, ok bool) {
	bounds := img.Bounds() // get the boundaries.
	// if (x,y) is not inside the boundaries
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return color.Black, false // return not ok
	}
	return img.At(x, y), true // return the color and ok = true
}

// isSameColor is an equals method for color values.
// We ignore the alpha component because it doens't matter for us.
func isSameColor(lhs, rhs color.Color) bool {
	lR, lG, lB, _ := lhs.RGBA()             // breaks down the left operand into its componts.
	rR, rG, rB, _ := rhs.RGBA()             // breaks down the right operand into its componts.
	return lR == rR && lG == rG && lB == rB // return true if every component is equal.
}

// collectColor gets the color at (x,y) and puts into the container colors.
// If the color at (x,y) is the keyColor, then ignores it, doing nothing.
func collectColor(img img, x, y int, colors *colors) {
	c, ok := getColor(img, x, y)         // get the color.
	if ok && !isSameColor(c, keyColor) { // if it must collect the color
		*colors = append(*colors, c) // collect it.
	}
}

// collectRing collects the color surrounding a point at (x,y).
// ring is the distance from (x,y) the we will consider this time.
func collectRing(img img, x, y int, colors *colors, ring int) {
	for i := x - ring; i < x+ring; i++ { // scan the upper row.
		collectColor(img, i, y-ring, colors)
	}
	for i := x - ring; i < x+ring; i++ { // scan the lower row.
		collectColor(img, i, y+ring, colors)
	}
	for i := y - ring + 1; i < y+ring-1; i++ { // scan the left column.
		collectColor(img, x-ring, i, colors)
	}
	for i := y - ring + 1; i < y+ring-1; i++ { // scan the right column.
		collectColor(img, x+ring, i, colors)
	}
}

// calculateColor returns the best color to be put in position (x,y).j
func calculateColor(img img, x, y int) color.Color {
	var colors colors                         // to collect the surrounding colors.
	for ring := 1; len(colors) == 0; ring++ { // will open the ring as much as needed.
		collectRing(img, x, y, &colors, ring) // collect this ring.
	}
	// sort the collected colors by brightness.
	sort.Slice(colors, func(i, j int) bool { return calcBrightness(colors[i]) < calcBrightness(colors[j]) })
	return colors[len(colors)/2] // pick up the median from the collected colors.
}

// corrected returns an automatically cleaned up version of img.
func corrected(img img) img {
	bounds := img.Bounds()                         // get the boundaries. We will scan the entire image.
	dst := image.NewRGBA(bounds)                   // a new image, to receive the cleaned up version.
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ { // will scan every row.
		for x := bounds.Min.X; x < bounds.Max.X; x++ { // will scan every column.
			c := img.At(x, y)             // get the color at point (x,y)
			if isSameColor(c, keyColor) { // if this pixel should be replaced
				c = calculateColor(img, x, y) // calculate a better match for this point.
			}
			dst.Set(x, y, c) // put the resulting color at the destination image.j
		}
	}
	return dst // returns the resulting image.j
}

// main is the program entry point.
func main() {
	if len(os.Args) < 3 { // if we don't have enought command line parameters
		fmt.Fprintln(os.Stderr, "Input and output filenames are missing.")
		os.Exit(1)
	}

	in, err := os.Open(os.Args[1]) // open up the source file.
	if err != nil {
		fmt.Fprintln(os.Stderr, "Coudn't open the input file for reading.")
		os.Exit(1)
	}
	defer in.Close() // close later.

	img, err := png.Decode(in) // reads the stored image.
	if err != nil {
		fmt.Fprintln(os.Stderr, "Coudn't decode image from input file.")
		os.Exit(1)
	}
	if img.Bounds().Empty() { // ensure it is not empty.
		fmt.Fprintln(os.Stderr, "Input image is empty")
		os.Exit(1)
	}

	keyColor = img.At(0, 0) // we assume the color at the upper left corner is the key.
	img = corrected(img)    // produces a fixed image.

	out, err := os.Create(os.Args[2]) // open up the destination file.
	if err != nil {
		fmt.Fprintln(os.Stderr, "Coudn't open the output file for writing.")
		os.Exit(1)
	}
	defer out.Close() // close later.

	err = png.Encode(out, img) // encode the resulting image in PNG format.h
	if err != nil {
		panic(err)
	}
}
