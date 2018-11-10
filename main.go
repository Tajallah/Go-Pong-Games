package main

//TODO: ADD a server to upload high scores to

import (
	//"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"math/rand"
	"os"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
)

//constants
const WINDOW_HEIGHT = 800
const WINDOW_WIDTH = 1000
const SPEED_LIMIT = 25
const BRICK_VALUE = 100

func loadFont(path string, size float64) (font.Face, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	font, err := truetype.Parse(bytes)
	if err != nil {
		return nil, err
	}

	return truetype.NewFace(font, &truetype.Options{
		Size:              size,
		GlyphCacheEntries: 1,
	}), nil
}

type score struct {
	name  []rune
	score int
}

type brickMatrix struct {
	rows [10][20]int
}

/*
func letterInput () rune {
	if win.JustPressed(pixelgl.KeyQ) {
	return 'Q'
	} else if win.JustPressed(pixelgl.KeyQ)
}*/

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func deflection(position, velocity, center pixel.Vec, score_multi int, lives int) (pixel.Vec, pixel.Vec, int, int) {
	random := rand.Float64()
	if random < -10 {
		random = -10
	} else if random > 10 {
		random = 10
	}
	if position.X > WINDOW_WIDTH {
		velocity.X = velocity.X * -1.5
		position.X = WINDOW_WIDTH - 1
	}
	if position.X < 0 {
		score_multi = 0
		velocity.X = velocity.X * -1.5
		position.X = 1
	}
	if position.Y > WINDOW_HEIGHT {
		velocity.Y = velocity.Y * -1.5
		position.Y = WINDOW_HEIGHT - 1
	}
	if position.Y < 0 {
		position = center
		velocity = pixel.V(random, 1)
		lives = lives - 1
	}
	if velocity.X > 7 {
		velocity.X = 7
	}
	if velocity.Y > 7 {
		velocity.Y = 7
	}
	return position, velocity, score_multi, lives

}

func padDeflect(velocity pixel.Vec) pixel.Vec {
	velocity.Y = velocity.Y * -1
	return velocity
}

func ballUpdate(position, velocity, center pixel.Vec, score_multi int, lives int) (pixel.Vec, pixel.Vec, int, int) {
	position = position.Add(velocity)
	position, velocity, score_multi, lives = deflection(position, velocity, center, score_multi, lives)
	return position, velocity, score_multi, lives
}

func intcp(in []int) []int {
	out := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	copy(out, in)
	return out
}

func rcp(in []pixel.Rect) []pixel.Rect {
	ph := pixel.R(0, 0, 0, 0)
	out := []pixel.Rect{ph, ph, ph, ph, ph, ph, ph, ph, ph, ph, ph, ph, ph, ph, ph, ph, ph, ph, ph, ph}
	copy(out, in)
	return out
}

func run() {
	//Overall Game State
	gameover := false
	i := 0

	//points stuff
	scorefont, err := loadFont("fonts/ARCADECLASSIC.TTF", 35)
	if err != nil {
		panic(err)
	}
	SCORE := 0
	score_multi := 1
	atlas := text.NewAtlas(scorefont, text.ASCII)
	score_txt := text.New(pixel.V(0, 0), atlas)
	score_txt.Color = colornames.Blue
	fmt.Fprintln(score_txt, SCORE)
	initials := make([]rune, 3)
	initials_txt := text.New(pixel.V(WINDOW_HEIGHT/2, WINDOW_WIDTH/2), atlas)
	initials_txt.Color = colornames.Black

	//lives stuff
	lives := 3
	lives_txt := text.New(pixel.V(WINDOW_WIDTH-55, 0), atlas)
	lives_txt.Color = colornames.Green
	fmt.Fprintln(lives_txt, lives)

	//these can probably be simplified
	brkln := []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	brk := [10][]int{intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln)}
	br := pixel.R(0, 0, 0, 0)
	hbln := []pixel.Rect{br, br, br, br, br, br, br, br, br, br, br, br, br, br, br, br, br, br, br, br}
	brickHitboxes := [10][]pixel.Rect{rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln)}

	cfg := pixelgl.WindowConfig{
		Title:  "Breakout 0.1",
		Bounds: pixel.R(0, 0, WINDOW_WIDTH, WINDOW_HEIGHT),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	//load sprites
	paddle, err := loadPicture("sprites/paddle.png")
	if err != nil {
		panic(err)
	}
	ball, err := loadPicture("sprites/ball.png")
	if err != nil {
		panic(err)
	}
	brick, err := loadPicture("sprites/brick.png")
	if err != nil {
		panic(err)
	}
	pad_sprite := pixel.NewSprite(paddle, paddle.Bounds())
	ball_sprite := pixel.NewSprite(ball, ball.Bounds())
	brick_sprite := pixel.NewSprite(brick, brick.Bounds())
	ballPosition := win.Bounds().Center()
	ballVector := pixel.V(1, 1)
	paddleVector := pixel.V(0, 0)
	//Ok these are kinda confusing to look at. Each one is a rectangle defined by the distance from the center of the two respective sprites
	padRect := pixel.R(paddleVector.X-50, paddleVector.Y-4, paddleVector.X+50, paddleVector.Y+4)
	ballRect := pixel.R(ballPosition.X-4, ballPosition.Y-4, ballPosition.X+4, ballPosition.Y+4)

	//Main loop
	for !win.Closed() {
		if gameover == false {
			//update text
			score_txt.Clear()
			fmt.Fprintln(score_txt, SCORE)
			lives_txt.Clear()
			fmt.Fprintln(lives_txt, lives)

			//get the paddle's position
			win.Clear(colornames.Skyblue)
			X := win.MousePosition().X
			Y := win.Bounds().Center().Y - 175
			paddleVector := pixel.V(X, Y)
			pad_sprite.Draw(win, pixel.IM.Moved(paddleVector))

			//get the ball's position
			ballPosition, ballVector, score_multi, lives = ballUpdate(ballPosition, ballVector, win.Bounds().Center(), score_multi, lives)
			ball_sprite.Draw(win, pixel.IM.Moved(ballPosition))
			if score_multi == 0 {
				SCORE = SCORE / 2
				score_multi = 1
			}

			//build the wall
			locationVector := pixel.V(0, 0)
			for y, r := range brk {
				for x, b := range r {
					if b > 0 {
						//TODO: simplify these...
						locationVector = pixel.V(float64(((x+1)*50)-25), WINDOW_HEIGHT-float64(((((WINDOW_HEIGHT/25)-y)*25)-12)))
						brick_sprite.Draw(win, pixel.IM.Moved(pixel.V(locationVector.X, WINDOW_HEIGHT-locationVector.Y)))
						brickHitboxes[y][x] = pixel.R(locationVector.X-25, WINDOW_HEIGHT-locationVector.Y-12, locationVector.X+25, WINDOW_HEIGHT-locationVector.Y+12)
					} else {
						brickHitboxes[y][x] = pixel.R(-1, -1, -1, -1)
					}
				}
			}

			//update hitboxes
			padRect = pixel.R(paddleVector.X-50, paddleVector.Y-4, paddleVector.X+50, paddleVector.Y+4)
			ballRect = pixel.R(ballPosition.X-4, ballPosition.Y-4, ballPosition.X+4, ballPosition.Y+4)

			//Check for paddle collision
			if padRect.Intersect(ballRect) != pixel.R(0, 0, 0, 0) {
				score_multi = 1
				ballVector = padDeflect(ballVector)
			}

			//check for brick collision
			for y, q := range brickHitboxes {
				for x, _ := range q {
					intersect := ballRect.Intersect(brickHitboxes[y][x])
					if intersect != pixel.R(0, 0, 0, 0) {
						score_multi += 1
						SCORE += (score_multi * BRICK_VALUE)
						if intersect.Min.X == ballRect.Min.X { // hit from the top or bottom
							ballVector.Y = (ballVector.Y * -2)
							if ballVector.Y > 7 {
								ballVector.Y = 7
							}
							if ballVector.Y < -7 {
								ballVector.Y = -7
							}
							brk[y][x] = 0
							continue
						}
						if intersect.Min.Y == ballRect.Min.Y { // hit from the left or right
							ballVector.X = (ballVector.X * -2)
							if ballVector.X > 7 {
								ballVector.X = 7
							}
							if ballVector.X < -7 {
								ballVector.X = -7
							}
							brk[y][x] = 0
							continue
						}
					}
				}
			}
			//draw score text
			score_txt.Draw(win, pixel.IM)
			lives_txt.Draw(win, pixel.IM)
			if lives <= 0 {
				lives = 0
				gameover = true
			}
			win.Update()
		} else {
			//ask for initials and send to server
			initials_txt.Clear()
			initials_txt.Draw(win, pixel.IM)
			score_txt.Draw(win, pixel.IM)
			q := win.Typed()
			if len(q) != 0 {
				initials[i] = rune(q[0])
				i += 1
			}
			win.Update()
			if i == 2 {
				// upload to server and wait for response
				return
			}
		}
	}
}

func main() {
	pixelgl.Run(run)
}
