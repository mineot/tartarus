// import (
// "image/color"
// "strings"

// "charm.land/lipgloss/v2"
// )

// type pixelFont struct {
// height int
// spacing int
// glyphs map[rune][]string
// }

// var logoFont = pixelFont{
// height: 4,
// spacing: 1,
// glyphs: map[rune][]string{
// 'T': {"▀█▀", " █ ", " █ ", " █ "},
// 'A': {"█▀█", "█ █", "█▀█", "█ █"},
// 'R': {"█▀▄", "█▄▀", "█▀▄", "█ █"},
// 'U': {"█ █", "█ █", "█ █", "▀▄▀"},
// 'S': {"█▀▀", "▀█▄", " █", "▄▄█"},
// '?': {"▀▀█", " ▄▀", " ", " ▄ "},
// },
// }

// var versionFont = pixelFont{
// height: 3,
// spacing: 1,
// glyphs: map[rune][]string{
// '0': {"█▀█", "█ █", "█▄█"},
// '1': {"▄█ ", " █ ", "▄█▄"},
// '2': {"▀▀█", "▄▀▀", "█▄▄"},
// '3': {"▀▀█", " ▀█", "▄▄█"},
// '4': {"█ █", "▀▀█", " █"},
// '5': {"█▀▀", "▀▀█", "▄▄█"},
// '6': {"█▀▀", "█▀█", "█▄█"},
// '7': {"▀▀█", " █", " █"},
// '8': {"█▀█", "█▀█", "█▄█"},
// '9': {"█▀█", "▀▀█", "▄▄█"},
// '.': {" ", " ", "▄"},
// '?': {"▀▀█", " ▀ ", " ▄ "},
// },
// }

// const gradientSteps = 32

// var pixelStyles = buildPixelStyles()

// func buildPixelStyles() [2][gradientSteps]lipgloss.Style {
// stops := [...]color.RGBA{
// // OKLCH palette converted to 8-bit sRGB for terminal rendering.
// {R: 241, G: 245, B: 249, A: 255}, // oklch(96.8% 0.007 247.896)
// {R: 226, G: 232, B: 240, A: 255}, // oklch(92.9% 0.013 255.508)
// {R: 202, G: 213, B: 226, A: 255}, // oklch(86.9% 0.022 252.894)
// {R: 144, G: 161, B: 185, A: 255}, // oklch(70.4% 0.04 256.788)
// {R: 98, G: 116, B: 142, A: 255}, // oklch(55.4% 0.046 257.417)
// }

// var styles [2][gradientSteps]lipgloss.Style

// for row := range styles {
// brightness := 1.0

// if row == 1 {
// brightness = 0.78
// }

// for step := range gradientSteps {
// position := float64(step) \* float64(len(stops)-1) / float64(gradientSteps-1)
// index := min(int(position), len(stops)-2)
// fraction := position - float64(index)
// from, to := stops[index], stops[index+1]

// blend := func(a, b uint8) uint8 {
// return uint8((float64(a)*(1-fraction) + float64(b)*fraction) \* brightness)
// }

// styles[row][step] = lipgloss.NewStyle().Foreground(color.RGBA{
// R: blend(from.R, to.R),
// G: blend(from.G, to.G),
// B: blend(from.B, to.B),
// A: 255,
// })
// }
// }

// return styles
// }

// func renderPixelText(text string, font pixelFont) string {
// if text == "" {
// return ""
// }

// characters := []rune(strings.ToUpper(text))
// lines := make([]string, font.height)

// for row := range font.height {
// var line strings.Builder

// for index, character := range characters {
// if index > 0 {
// line.WriteString(strings.Repeat(" ", font.spacing))
// }

// if character == ' ' {
// line.WriteString(" ")
// continue
// }

// glyph, exists := font.glyphs[character]

// if !exists {
// glyph = font.glyphs['?']
// }

// line.WriteString(glyph[row])
// }

// lines[row] = line.String()
// }

// width := lipgloss.Width(strings.Join(lines, "\n"))

// for row, line := range lines {
// var colored strings.Builder

// for column, pixel := range []rune(line) {
// if pixel == ' ' {
// colored.WriteByte(' ')
// continue
// }

// // Coordinate-based noise keeps the distressed gradient stable across redraws.
// noise := ((column*17 + row*31 + column*row*13) % 7) - 3
// step := column \* (gradientSteps - 1) / max(width-1, 1)
// step = max(0, min(gradientSteps-1, step+noise))
// colored.WriteString(pixelStyles[row%2][step].Render(string(pixel)))
// }

// lines[row] = colored.String()
// }

// return strings.Join(lines, "\n")
// }

// func renderLogo() string {
// title := renderPixelText("Tartarus", logoFont)
// version := renderPixelText("2.0.0", versionFont)
// return lipgloss.JoinVertical(lipgloss.Left, title, "", version)
// }
