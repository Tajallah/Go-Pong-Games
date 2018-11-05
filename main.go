package main

import (
	"fmt"
	"image"
	"math/rand"
	"os"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

//constants
const WINDOW_HEIGHT = 800
const WINDOW_WIDTH = 1000
const SPEED_LIMIT = 25

type brickMatrix struct {
	rows [10][20]int
}

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

func deflection(position, velocity, center pixel.Vec) (pixel.Vec, pixel.Vec) {
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
	}
	if velocity.X > 7 {
		velocity.X = 7
	}
	if velocity.Y > 7 {
		velocity.Y = 7
	}
	return position, velocity

}

func padDeflect(velocity pixel.Vec) pixel.Vec {
	velocity.Y = velocity.Y * -1
	return velocity
}

func ballUpdate(position, velocity, center pixel.Vec) (pixel.Vec, pixel.Vec) {
	position = position.Add(velocity)
	position, velocity = deflection(position, velocity, center)
	return position, velocity
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
	//brickshit
	brkln := []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	brk := [10][]int{intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln), intcp(brkln)}
	br := pixel.R(0, 0, 0, 0)
	hbln := []pixel.Rect{br, br, br, br, br, br, br, br, br, br, br, br, br, br, br, br, br, br, br, br}
	brickHitboxes := [10][]pixel.Rect{rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln), rcp(hbln)}

	cfg := pixelgl.WindowConfig{
		Title:  "TEST",
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
		//get the paddle's position
		win.Clear(colornames.Skyblue)
		X := win.MousePosition().X
		Y := win.Bounds().Center().Y - 175
		paddleVector := pixel.V(X, Y)
		pad_sprite.Draw(win, pixel.IM.Moved(paddleVector))

		//get the ball's position
		ballPosition, ballVector = ballUpdate(ballPosition, ballVector, win.Bounds().Center())
		ball_sprite.Draw(win, pixel.IM.Moved(ballPosition))


		//build the wall
		locationVector := pixel.V(0, 0)
		for y, r := range brk {
			for x, b := range r {
				if b > 0 {
					//TODO: simplify these...
					locationVector = pixel.V(float64(((x+1)*50)-25), WINDOW_HEIGHT-float64(((((WINDOW_HEIGHT/25)-y)*25)-12)))
					brick_sprite.Draw(win, pixel.IM.Moved(pixel.V(locationVector.X, WINDOW_HEIGHT-locationVector.Y)))
					brickHitboxes[y][x] = pixel.R(locationVector.X - 25, WINDOW_HEIGHT-locationVector.Y - 12, locationVector.X + 25, WINDOW_HEIGHT-locationVector.Y + 12)
				} else {
					brickHitboxes[y][x] = pixel.R(-1,-1,-1,-1)
				}
			}
		}
		fmt.Println(locationVector)
		fmt.Println(brickHitboxes)

		//update hitboxes
		padRect = pixel.R(paddleVector.X-50, paddleVector.Y-4, paddleVector.X+50, paddleVector.Y+4)
		ballRect = pixel.R(ballPosition.X-4, ballPosition.Y-4, ballPosition.X+4, ballPosition.Y+4)

		fmt.Println("BALL RECT -- ", ballRect)

		//Check for paddle collision
		if padRect.Intersect(ballRect) != pixel.R(0, 0, 0, 0) {
			ballVector = padDeflect(ballVector)
		}

		//check for brick collision
		for y, q := range brickHitboxes {
			for x, _ := range q {
				intersect := ballRect.Intersect(brickHitboxes[y][x])
				if intersect != pixel.R(0, 0, 0, 0) {
					fmt.Println("BRICK INTERSECT ", intersect)
					if intersect.Min.X == ballRect.Min.X { // hit from the top or bottom
						fmt.Println("top Hit")
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
						fmt.Println("bottom hit")
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

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
