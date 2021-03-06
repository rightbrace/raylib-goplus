package rgif

import (
	"image/gif"
	"os"

	r "github.com/lachee/raylib-goplus/raylib"
)

type FrameDisposal int

const (
	FrameDisposalNone FrameDisposal = iota
	FrameDisposalDontDispose
	FrameDisposalRestoreBackground
	FrameDisposalRestorePrevious
)

//GifImage represents a gif texture
type GifImage struct {

	//Texture is the current frame of the gif
	Texture r.Texture2D
	//Width is the width of a single frame
	Width int
	//Height is the height of a single frame
	Height int
	//Frames is the number of frames available
	Frames int
	//Timing is the delay (in 100ths of seconds) a frame has
	Timing []int
	//Disposal is the disposal for each frame
	Disposal []FrameDisposal

	pixels        [][]r.Color //Cache of each frame's pixels
	currentFrame  int         //The current frame
	lastFrameTime float32     //Update since last frame
}

//LoadGifFromFile loads a new gif
func LoadGifFromFile(fileName string) (*GifImage, error) {

	//Read the GIF file
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	/*//Defer any panics
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error while decoding: %s", r)
		}
	}()*/

	//Decode teh gif
	gif, err := gif.DecodeAll(file)
	if err != nil {
		return nil, err
	}

	//Prepare the tilesheet and overpaint image.
	imgWidth, imgHeight := getGifDimensions(gif)
	frames := len(gif.Image)

	disposals := make([]FrameDisposal, frames)

	images := make([][]r.Color, frames, imgWidth*imgHeight)
	clumative := make([]r.Color, imgWidth*imgHeight)

	previousNonDisposed := gif.Image[0]

	for i, img := range gif.Image {
		disposals[i] = FrameDisposal(gif.Disposal[i])

		pixels := make([]r.Color, imgWidth*imgHeight)
		for y := 0; y < imgHeight; y++ {
			for x := 0; x < imgWidth; x++ {

				color := img.At(x, y)
				red, green, blue, alpha := color.RGBA()

				switch disposals[i] {
				case FrameDisposalNone:
					//Use all our pixels always
					pixels[x+y*imgWidth] = r.NewColor(uint8(red), uint8(green), uint8(blue), uint8(alpha))
					clumative[x+y*imgWidth] = pixels[x+y*imgWidth]
					previousNonDisposed = img

				case FrameDisposalDontDispose:
					if alpha > 0 {
						//Use our own pixels (clumative)
						pixels[x+y*imgWidth] = r.NewColor(uint8(red), uint8(green), uint8(blue), uint8(alpha))
						clumative[x+y*imgWidth] = pixels[x+y*imgWidth]
					} else {
						//Use the previous pixels
						pixels[x+y*imgWidth] = clumative[x+y*imgWidth]
					}
					previousNonDisposed = img

				case FrameDisposalRestoreBackground:
					if disposals[0] == FrameDisposalDontDispose && alpha == 0 {
						red, green, blue, alpha = gif.Image[0].At(x, y).RGBA()
					}

					pixels[x+y*imgWidth] = r.NewColor(uint8(red), uint8(green), uint8(blue), uint8(alpha))
					clumative[x+y*imgWidth] = pixels[x+y*imgWidth]

				case FrameDisposalRestorePrevious:
					if alpha == 0 {
						red, green, blue, alpha = previousNonDisposed.At(x, y).RGBA()
					}
					pixels[x+y*imgWidth] = r.NewColor(uint8(red), uint8(green), uint8(blue), uint8(alpha))
					clumative[x+y*imgWidth] = pixels[x+y*imgWidth]
				}
			}
		}

		images[i] = pixels
	}

	//Load the first initial texture
	texture := r.LoadTextureFromGo(gif.Image[0])

	return &GifImage{
		Texture:  texture,
		pixels:   images,
		Width:    imgWidth,
		Height:   imgHeight,
		Frames:   frames,
		Timing:   gif.Delay,
		Disposal: disposals,
	}, nil
}

//Step performs a time step.
func (gif *GifImage) Step(timeSinceLastStep float32) {

	gif.lastFrameTime += (timeSinceLastStep * 100)
	diff := gif.lastFrameTime - float32(gif.Timing[gif.currentFrame])

	if diff >= 0 {
		gif.NextFrame()
	}
}

//NextFrame increments the frame counter and resets the timing buffer
func (gif *GifImage) NextFrame() {
	gif.lastFrameTime -= float32(gif.Timing[gif.currentFrame])
	gif.currentFrame = (gif.currentFrame + 1) % gif.Frames
	if gif.lastFrameTime < 0 {
		gif.lastFrameTime = 0
	}

	gif.Texture.UpdateTexture(gif.pixels[gif.currentFrame])
}

//Reset clears the last frame time and resets the current frame to zero
func (gif *GifImage) Reset() {
	gif.currentFrame = 0
	gif.lastFrameTime = 0
}

//Unload unloads all the textures and images, making this gif unusable.
func (gif *GifImage) Unload() {
	gif.Texture.Unload()
}

//CurrentFrame returns the current frame index
func (gif *GifImage) CurrentFrame() int { return gif.currentFrame }

//CurrentTiming gets the current timing for the current frame
func (gif *GifImage) CurrentTiming() int { return gif.Timing[gif.currentFrame] }

//GetRectangle gets a rectangle crop for a specified frame
func (gif *GifImage) GetRectangle(frame int) r.Rectangle {
	return r.NewRectangle(float32(gif.Width*frame), 0, float32(gif.Width), float32(gif.Height))
}

//DrawGif draws a single frame of a gif
func DrawGif(gif *GifImage, x int, y int, tint r.Color) {
	r.DrawTexture(gif.Texture, x, y, tint)
}

//DrawGifEx draws a gif with rotation and scale
func DrawGifEx(gif *GifImage, position r.Vector2, rotation float32, scale float32, tint r.Color) {
	r.DrawTextureEx(gif.Texture, position, rotation, scale, tint)
}

func getGifDimensions(gif *gif.GIF) (x, y int) {
	var lowestX int
	var lowestY int
	var highestX int
	var highestY int

	for _, img := range gif.Image {
		if img.Rect.Min.X < lowestX {
			lowestX = img.Rect.Min.X
		}
		if img.Rect.Min.Y < lowestY {
			lowestY = img.Rect.Min.Y
		}
		if img.Rect.Max.X > highestX {
			highestX = img.Rect.Max.X
		}
		if img.Rect.Max.Y > highestY {
			highestY = img.Rect.Max.Y
		}
	}

	return highestX - lowestX, highestY - lowestY
}
