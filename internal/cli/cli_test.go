package cli

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TamasGorgics/gopiq"
	"golang.org/x/image/font/gofont/goregular"
)

func runCLI(args []string, stdin []byte) (int, []byte, []byte) {
	var stdout, stderr bytes.Buffer
	code := Main(args, bytes.NewReader(stdin), &stdout, &stderr)
	return code, stdout.Bytes(), stderr.Bytes()
}

func testImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8((x*17 + y*3) % 256),
				G: uint8((x*5 + y*29) % 256),
				B: uint8((x*11 + y*7) % 256),
				A: 255,
			})
		}
	}
	return img
}

func pngBytes(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
	return buf.Bytes()
}

func jpegBytes(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}
	return buf.Bytes()
}

func writeTestFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

func decodeImage(t *testing.T, data []byte) (image.Image, string) {
	t.Helper()
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode image: %v", err)
	}
	return img, format
}

func changedPixels(t *testing.T, first, second image.Image) int {
	t.Helper()
	if first.Bounds() != second.Bounds() {
		t.Fatalf("bounds differ: %v vs %v", first.Bounds(), second.Bounds())
	}
	changed := 0
	for y := first.Bounds().Min.Y; y < first.Bounds().Max.Y; y++ {
		for x := first.Bounds().Min.X; x < first.Bounds().Max.X; x++ {
			r1, g1, b1, a1 := first.At(x, y).RGBA()
			r2, g2, b2, a2 := second.At(x, y).RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				changed++
			}
		}
	}
	return changed
}

func TestMainHelpAndVersion(t *testing.T) {
	code, _, stderr := runCLI(nil, nil)
	if code != exitUsage {
		t.Fatalf("no args exit code = %d, want %d", code, exitUsage)
	}
	if !bytes.Contains(stderr, []byte("Usage:")) {
		t.Fatalf("no-args stderr = %q, want usage", stderr)
	}

	for _, args := range [][]string{{"help"}, {"--help"}} {
		code, stdout, _ := runCLI(args, nil)
		if code != exitSuccess {
			t.Fatalf("%v exit code = %d, want %d", args, code, exitSuccess)
		}
		if !bytes.Contains(stdout, []byte("Commands:")) {
			t.Fatalf("%v stdout = %q, want command list", args, stdout)
		}
	}

	code, stdout, _ := runCLI([]string{"help", "resize"}, nil)
	if code != exitSuccess || !bytes.Contains(stdout, []byte("Usage: gopiq resize")) {
		t.Fatalf("help resize = code %d, output %q", code, stdout)
	}

	code, stdout, _ = runCLI([]string{"resize", "--help"}, nil)
	if code != exitSuccess || !bytes.Contains(stdout, []byte("Usage: gopiq resize")) {
		t.Fatalf("resize --help = code %d, output %q", code, stdout)
	}

	for _, args := range [][]string{{"version"}, {"--version"}} {
		code, stdout, _ := runCLI(args, nil)
		if code != exitSuccess || !bytes.Contains(stdout, []byte("gopiq ")) {
			t.Fatalf("%v = code %d, output %q", args, code, stdout)
		}
	}

	code, _, stderr = runCLI([]string{"unknown"}, nil)
	if code != exitUsage || !bytes.Contains(stderr, []byte("unknown command")) {
		t.Fatalf("unknown command = code %d, stderr %q", code, stderr)
	}
}

func TestResizeCommand(t *testing.T) {
	dir := t.TempDir()
	input := writeTestFile(t, dir, "input.png", pngBytes(t, testImage(20, 10)))
	output := filepath.Join(dir, "output.png")

	code, _, stderr := runCLI([]string{
		"resize", "--width", "8", "--height", "6", input, output,
	}, nil)
	if code != exitSuccess {
		t.Fatalf("resize exit code = %d, stderr = %s", code, stderr)
	}
	img, _ := decodeImage(t, mustReadFile(t, output))
	if got, want := img.Bounds().Size(), image.Pt(8, 6); got != want {
		t.Fatalf("resize bounds = %v, want %v", got, want)
	}

	code, _, stderr = runCLI([]string{
		"resize", "--width", "0", "--height", "6", input, filepath.Join(dir, "bad.png"),
	}, nil)
	if code != exitUsage || !bytes.Contains(stderr, []byte("positive")) {
		t.Fatalf("invalid resize = code %d, stderr %q", code, stderr)
	}

	code, _, stderr = runCLI([]string{"resize", "--width", "8", "--height", "6"}, nil)
	if code != exitUsage || !bytes.Contains(stderr, []byte("positional")) {
		t.Fatalf("missing resize args = code %d, stderr %q", code, stderr)
	}
}

func TestCropCommand(t *testing.T) {
	dir := t.TempDir()
	source := testImage(20, 15)
	input := writeTestFile(t, dir, "input.png", pngBytes(t, source))
	output := filepath.Join(dir, "output.png")

	code, _, stderr := runCLI([]string{
		"crop", "--x", "2", "--y", "3", "--width", "4", "--height", "3", input, output,
	}, nil)
	if code != exitSuccess {
		t.Fatalf("crop exit code = %d, stderr = %s", code, stderr)
	}
	cropped, _ := decodeImage(t, mustReadFile(t, output))
	if got, want := cropped.Bounds().Size(), image.Pt(4, 3); got != want {
		t.Fatalf("crop bounds = %v, want %v", got, want)
	}
	if !sameColor(source.At(2, 3), cropped.At(0, 0)) {
		t.Fatal("crop did not preserve the source pixel at the crop origin")
	}

	code, _, stderr = runCLI([]string{
		"crop", "--x", "18", "--y", "0", "--width", "4", "--height", "3", input, filepath.Join(dir, "bad.png"),
	}, nil)
	if code != exitFailure || !bytes.Contains(stderr, []byte("out of image bounds")) {
		t.Fatalf("out-of-bounds crop = code %d, stderr %q", code, stderr)
	}

	code, _, stderr = runCLI([]string{
		"crop", "--x", "0", "--y", "0", "--width", "0", "--height", "3", input, filepath.Join(dir, "bad2.png"),
	}, nil)
	if code != exitUsage || !bytes.Contains(stderr, []byte("positive")) {
		t.Fatalf("invalid crop = code %d, stderr %q", code, stderr)
	}
}

func TestGrayscaleCommand(t *testing.T) {
	dir := t.TempDir()
	input := writeTestFile(t, dir, "input.png", pngBytes(t, testImage(20, 15)))

	for _, fast := range []bool{false, true} {
		name := "regular"
		args := []string{"grayscale"}
		if fast {
			name = "fast"
			args = append(args, "--fast")
		}
		output := filepath.Join(dir, name+".png")
		code, _, stderr := runCLI(append(args, input, output), nil)
		if code != exitSuccess {
			t.Fatalf("%s grayscale exit code = %d, stderr = %s", name, code, stderr)
		}
		img, _ := decodeImage(t, mustReadFile(t, output))
		r, g, b, _ := img.At(5, 4).RGBA()
		if r != g || g != b {
			t.Fatalf("%s grayscale pixel is not neutral: %d,%d,%d", name, r, g, b)
		}
	}
}

func TestConvertFormatsAndStreams(t *testing.T) {
	dir := t.TempDir()
	source := testImage(20, 10)
	pngInput := writeTestFile(t, dir, "input.png", pngBytes(t, source))
	jpegInput := writeTestFile(t, dir, "input.jpg", jpegBytes(t, source))

	pngToJPEG := filepath.Join(dir, "png-to-jpeg.jpg")
	code, _, stderr := runCLI([]string{"convert", pngInput, pngToJPEG}, nil)
	if code != exitSuccess {
		t.Fatalf("PNG to JPEG exit code = %d, stderr = %s", code, stderr)
	}
	if _, format := decodeImage(t, mustReadFile(t, pngToJPEG)); format != "jpeg" {
		t.Fatalf("PNG to JPEG output format = %q, want jpeg", format)
	}

	jpegToPNG := filepath.Join(dir, "jpeg-to-png.png")
	code, _, stderr = runCLI([]string{"convert", jpegInput, jpegToPNG}, nil)
	if code != exitSuccess {
		t.Fatalf("JPEG to PNG exit code = %d, stderr = %s", code, stderr)
	}
	if _, format := decodeImage(t, mustReadFile(t, jpegToPNG)); format != "png" {
		t.Fatalf("JPEG to PNG output format = %q, want png", format)
	}

	override := filepath.Join(dir, "override.bin")
	code, _, stderr = runCLI([]string{"convert", "--format", "png", pngInput, override}, nil)
	if code != exitSuccess {
		t.Fatalf("format override exit code = %d, stderr = %s", code, stderr)
	}
	if _, format := decodeImage(t, mustReadFile(t, override)); format != "png" {
		t.Fatalf("format override output = %q, want png", format)
	}

	code, stdout, stderr := runCLI([]string{"convert", "--format", "png", pngInput, "-"}, nil)
	if code != exitSuccess {
		t.Fatalf("stdout conversion exit code = %d, stderr = %s", code, stderr)
	}
	if _, format := decodeImage(t, stdout); format != "png" {
		t.Fatalf("stdout conversion format = %q, want png", format)
	}

	stdinOutput := filepath.Join(dir, "stdin-output.png")
	code, _, stderr = runCLI([]string{"convert", "--format", "png", "-", stdinOutput}, pngBytes(t, source))
	if code != exitSuccess {
		t.Fatalf("stdin conversion exit code = %d, stderr = %s", code, stderr)
	}
	if _, format := decodeImage(t, mustReadFile(t, stdinOutput)); format != "png" {
		t.Fatalf("stdin conversion output = %q, want png", format)
	}
}

func TestFormatAndGIFHandling(t *testing.T) {
	dir := t.TempDir()
	input := writeTestFile(t, dir, "input.png", pngBytes(t, testImage(4, 3)))

	for _, name := range []string{"out.jpg", "out.jpeg", "out.png"} {
		output := filepath.Join(dir, name)
		code, _, stderr := runCLI([]string{"convert", input, output}, nil)
		if code != exitSuccess {
			t.Fatalf("%s exit code = %d, stderr = %s", name, code, stderr)
		}
	}

	code, _, stderr := runCLI([]string{"convert", input, filepath.Join(dir, "out.bin")}, nil)
	if code != exitUsage || !bytes.Contains(stderr, []byte("cannot infer")) {
		t.Fatalf("unknown extension = code %d, stderr %q", code, stderr)
	}

	code, _, stderr = runCLI([]string{"convert", input, "-"}, nil)
	if code != exitUsage || !bytes.Contains(stderr, []byte("--format is required")) {
		t.Fatalf("stdout without format = code %d, stderr %q", code, stderr)
	}

	code, _, stderr = runCLI([]string{"convert", "--format", "gif", input, filepath.Join(dir, "out.gif")}, nil)
	if code != exitFailure || !bytes.Contains(stderr, []byte("GIF output")) {
		t.Fatalf("gif format = code %d, stderr %q", code, stderr)
	}

	const onePixelGIF = "R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
	gifData, err := base64.StdEncoding.DecodeString(onePixelGIF)
	if err != nil {
		t.Fatalf("decode GIF fixture: %v", err)
	}
	gifInput := writeTestFile(t, dir, "input.gif", gifData)
	gifOutput := filepath.Join(dir, "from-gif.png")
	code, _, stderr = runCLI([]string{"convert", gifInput, gifOutput}, nil)
	if code != exitSuccess {
		t.Fatalf("GIF input exit code = %d, stderr = %s", code, stderr)
	}
	img, _ := decodeImage(t, mustReadFile(t, gifOutput))
	if got, want := img.Bounds().Size(), image.Pt(1, 1); got != want {
		t.Fatalf("GIF output size = %v, want %v", got, want)
	}
}

func TestWatermarkCommand(t *testing.T) {
	dir := t.TempDir()
	source := testImage(100, 80)
	input := writeTestFile(t, dir, "input.png", pngBytes(t, source))
	fontPath := writeTestFile(t, dir, "font.ttf", goregular.TTF)

	outputs := map[string][]string{
		"default": {"watermark", "--text", "M5", input},
		"custom": {
			"watermark", "--text", "M5", "--font-size", "18", "--color", "#ff0000ff",
			"--position", "top-left", "--offset", "2,3", "--font", fontPath, input,
		},
	}
	for name, args := range outputs {
		output := filepath.Join(dir, name+".png")
		code, _, stderr := runCLI(append(args, output), nil)
		if code != exitSuccess {
			t.Fatalf("%s watermark exit code = %d, stderr = %s", name, code, stderr)
		}
		img, _ := decodeImage(t, mustReadFile(t, output))
		if changedPixels(t, source, img) == 0 {
			t.Fatalf("%s watermark produced no changed pixels", name)
		}
	}

	invalid := []struct {
		name string
		args []string
		code int
	}{
		{name: "empty text", args: []string{"watermark", "--text", ""}, code: exitUsage},
		{name: "invalid color", args: []string{"watermark", "--text", "M5", "--color", "zzz"}, code: exitUsage},
		{name: "invalid position", args: []string{"watermark", "--text", "M5", "--position", "middle"}, code: exitUsage},
		{name: "invalid offset", args: []string{"watermark", "--text", "M5", "--offset", "1"}, code: exitUsage},
		{name: "invalid font size", args: []string{"watermark", "--text", "M5", "--font-size", "0"}, code: exitUsage},
		{name: "missing font", args: []string{"watermark", "--text", "M5", "--font", filepath.Join(dir, "missing.ttf")}, code: exitFailure},
	}
	for _, tc := range invalid {
		output := filepath.Join(dir, strings.ReplaceAll(tc.name, " ", "-")+".png")
		code, _, stderr := runCLI(append(tc.args, input, output), nil)
		if code != tc.code {
			t.Fatalf("%s = code %d, want %d; stderr %q", tc.name, code, tc.code, stderr)
		}
	}
}

func TestStdinStdoutPipeline(t *testing.T) {
	source := testImage(12, 8)
	input := pngBytes(t, source)

	code, stdout, stderr := runCLI([]string{"grayscale", "--format", "png", "-", "-"}, input)
	if code != exitSuccess {
		t.Fatalf("stdin/stdout exit code = %d, stderr = %s", code, stderr)
	}
	img, format := decodeImage(t, stdout)
	if format != "png" {
		t.Fatalf("stdout format = %q, want png", format)
	}
	r, g, b, _ := img.At(3, 3).RGBA()
	if r != g || g != b {
		t.Fatal("stdin/stdout grayscale output is not neutral")
	}
}

func TestOutputOverwriteAndFailures(t *testing.T) {
	dir := t.TempDir()
	input := writeTestFile(t, dir, "input.png", pngBytes(t, testImage(10, 8)))

	output := filepath.Join(dir, "output.png")
	if err := os.WriteFile(output, []byte("old"), 0o600); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}
	code, _, stderr := runCLI([]string{"resize", "--width", "5", "--height", "4", input, output}, nil)
	if code != exitSuccess {
		t.Fatalf("overwrite exit code = %d, stderr = %s", code, stderr)
	}
	overwritten, _ := decodeImage(t, mustReadFile(t, output))
	if got, want := overwritten.Bounds().Size(), image.Pt(5, 4); got != want {
		t.Fatalf("overwritten size = %v, want %v", got, want)
	}

	missingDirOutput := filepath.Join(dir, "missing", "out.png")
	code, _, stderr = runCLI([]string{"convert", input, missingDirOutput}, nil)
	if code != exitFailure {
		t.Fatalf("missing directory = code %d, stderr %q", code, stderr)
	}
	if _, err := os.Stat(missingDirOutput); !os.IsNotExist(err) {
		t.Fatalf("missing-directory output exists: %v", err)
	}

	directoryOutput := filepath.Join(dir, "directory.png")
	if err := os.Mkdir(directoryOutput, 0o755); err != nil {
		t.Fatalf("create output directory: %v", err)
	}
	code, _, stderr = runCLI([]string{"convert", input, directoryOutput}, nil)
	if code != exitFailure {
		t.Fatalf("directory output = code %d, stderr %q", code, stderr)
	}

	missingInput := filepath.Join(dir, "missing.png")
	code, _, stderr = runCLI([]string{"convert", missingInput, filepath.Join(dir, "missing-input-output.png")}, nil)
	if code != exitFailure || !bytes.Contains(stderr, []byte("read")) {
		t.Fatalf("missing input = code %d, stderr %q", code, stderr)
	}

	invalidInput := writeTestFile(t, dir, "invalid.png", []byte("not an image"))
	code, _, stderr = runCLI([]string{"convert", invalidInput, filepath.Join(dir, "invalid-output.png")}, nil)
	if code != exitFailure || !bytes.Contains(stderr, []byte("decode")) {
		t.Fatalf("invalid input = code %d, stderr %q", code, stderr)
	}

	gifOutput := filepath.Join(dir, "existing.gif")
	sentinel := []byte("sentinel")
	if err := os.WriteFile(gifOutput, sentinel, 0o600); err != nil {
		t.Fatalf("write GIF sentinel: %v", err)
	}
	code, _, stderr = runCLI([]string{"convert", input, gifOutput}, nil)
	if code != exitFailure || !bytes.Contains(stderr, []byte("GIF output")) {
		t.Fatalf("GIF output = code %d, stderr %q", code, stderr)
	}
	if got := mustReadFile(t, gifOutput); !bytes.Equal(got, sentinel) {
		t.Fatalf("GIF failure changed destination: %q", got)
	}
}

func TestParsingHelpers(t *testing.T) {
	colorTests := []struct {
		value string
		want  color.NRGBA
	}{
		{value: "#fff", want: color.NRGBA{R: 255, G: 255, B: 255, A: 255}},
		{value: "#f00f", want: color.NRGBA{R: 255, A: 255}},
		{value: "#123456", want: color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 255}},
		{value: "#12345678", want: color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0x78}},
	}
	for _, tc := range colorTests {
		got, err := parseColor(tc.value)
		if err != nil {
			t.Fatalf("parseColor(%q): %v", tc.value, err)
		}
		if got != tc.want {
			t.Fatalf("parseColor(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
	for _, value := range []string{"#12", "#xyz", "#1234567"} {
		if _, err := parseColor(value); err == nil {
			t.Fatalf("parseColor(%q) unexpectedly succeeded", value)
		}
	}

	positionTests := map[string]gopiq.WatermarkPosition{
		"top-left":     gopiq.PositionTopLeft,
		"TOP-RIGHT":    gopiq.PositionTopRight,
		"bottom-left":  gopiq.PositionBottomLeft,
		"bottom-right": gopiq.PositionBottomRight,
		"center":       gopiq.PositionCenter,
	}
	for value, want := range positionTests {
		got, err := parsePosition(value)
		if err != nil || got != want {
			t.Fatalf("parsePosition(%q) = %v, %v; want %v", value, got, err, want)
		}
	}
	if _, err := parsePosition("middle"); err == nil {
		t.Fatal("parsePosition accepted an unknown position")
	}

	if x, y, err := parseOffset("-1,2.5"); err != nil || x != -1 || y != 2.5 {
		t.Fatalf("parseOffset = %v, %v, %v", x, y, err)
	}
	for _, value := range []string{"1", "1,", "x,2", "NaN,2"} {
		if _, _, err := parseOffset(value); err == nil {
			t.Fatalf("parseOffset(%q) unexpectedly succeeded", value)
		}
	}
	if size, err := parseFontSize("12.5"); err != nil || size != 12.5 {
		t.Fatalf("parseFontSize = %v, %v", size, err)
	}
	for _, value := range []string{"0", "-1", "NaN", "Inf"} {
		if _, err := parseFontSize(value); err == nil {
			t.Fatalf("parseFontSize(%q) unexpectedly succeeded", value)
		}
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func sameColor(first, second color.Color) bool {
	r1, g1, b1, a1 := first.RGBA()
	r2, g2, b2, a2 := second.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
