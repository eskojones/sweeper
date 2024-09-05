package main

import (
    "os"
	"fmt"
    "strconv"
	"math/rand"
    "golang.org/x/term"
)

type Game struct {
    width    int
    height   int     
    field  []bool  // mine field
    state  []uint8 // 's' show, 'h' hidden, 'f' flagged
    curX     int
    curY     int
    moves    int
    mines    int // mines left unflagged
    gameover bool  
}

type Point struct {
    x int
    y int
}

const C_CURSOR   = 82
const C_MINE     = 198
const C_HIDDEN   = 45
const C_FLAGGED  = 207

const C_MINES_LOW  = 43
const C_MINES_MID  = 214
const C_MINES_HIGH = 196

const CH_CUR_A   = '['
const CH_CUR_B   = ']'
const CH_MINE    = '*'
const CH_HIDDEN  = '.'
const CH_FLAGGED = 'x'


func (g *Game) Init(width int, height int) {
    g.width = width
    g.height = height
    g.field = make([]bool, width * height)
    g.state = make([]uint8, width * height)
}

func (g *Game) Reset(minesTotal int) {
    for minesTotal > 1 {
        idx := rand.Int() % (g.width * g.height)
        for g.field[idx] {
            idx = rand.Int() % (g.width * g.height)
        }
        g.field[idx] = true
        minesTotal--
    }
    g.mines = minesTotal
    for i := range g.field {
        g.state[i] = 'h'
    }
}

func (g *Game) CheckCursor() bool {
    return g.field[g.curY * g.width + g.curX]
}

func (g *Game) CheckNeighbours(x int, y int) int {
    mines := 0
    for oy := -1; oy <= 1; oy++ {
        if y+oy < 0 || y+oy >= g.height {
            continue
        }
        for ox := -1; ox <= 1; ox++ {
            if x+ox < 0 || x+ox >= g.width {
                continue
            }
            if g.field[(y+oy) * g.width + (x+ox)] {
                mines++
            }
        }
    }
    return mines
}

func (g *Game) FloodFill(x int, y int, temp []int) {
    if x < 0 || x >= g.width || y < 0 || y >= g.height {
        return
    }
    idx := y * g.width + x
    if g.field[idx] || g.state[idx] == 'f' {        
        return 
    }
    if temp[idx] != 0 || g.CheckNeighbours(x, y) > 3 {
        return
    }
    temp[idx] = 1
    g.FloodFill(x - 1, y, temp)
    g.FloodFill(x + 1, y, temp)
    g.FloodFill(x, y - 1, temp)
    g.FloodFill(x, y + 1, temp)
    g.FloodFill(x - 1, y - 1, temp)
    g.FloodFill(x + 1, y - 1, temp)
    g.FloodFill(x - 1, y + 1, temp)
    g.FloodFill(x + 1, y + 1, temp)
}

func (g *Game) GetReveals(x int, y int) []*Point {
    if g.field[y * g.width + x] {
        return nil
    }
    points := make([]*Point, 1)
    points[0] = &Point{x, y}
    temp := make([]int, g.width * g.height)
    for idx := range g.field {
        if g.field[idx] || g.state[idx] == 'f' {
            temp[idx] = -1
        }
    }
    g.FloodFill(x, y, temp)
    for y := 0; y < g.height; y++ {
        for x := 0; x < g.width; x++ {
            if temp[y * g.width + x] == 1 {
                points = append(points, &Point{ x, y })
            }
        }
    }
    return points
}

func (g *Game) Draw() {
    for y := 0; y < g.height; y++ {
        for x := 0; x < g.width; x++ {
            idx := y * g.width + x
            isCur := g.curX == x && g.curY == y
            if isCur {
                fmt.Printf("%s%c", colourFg(C_CURSOR), CH_CUR_A)
            } else {
                fmt.Printf(" ")
            }
            if g.state[idx] == 'h' {
                fmt.Printf("%s%c", colourFg(C_HIDDEN), CH_HIDDEN)
            } else if g.state[idx] == 'f' {
                fmt.Printf("%s%c", colourFg(C_FLAGGED), CH_FLAGGED)
            } else if g.state[idx] == 's' {
                if g.field[idx] {
                    fmt.Printf("%s%c", colourFg(C_MINE), CH_MINE)
                } else {
                    neighbours := g.CheckNeighbours(x, y)
                    if neighbours == 0 {
                        fmt.Printf(" ")
                    } else {
                        c := C_MINES_LOW
                        if neighbours >= 5 {
                            c = C_MINES_HIGH
                        } else if neighbours >= 3 {
                            c = C_MINES_MID
                        }
                        fmt.Printf("%s%d", colourFg(c), neighbours)
                    }
                }
            }
            if isCur {
                fmt.Printf("%s%c", colourFg(C_CURSOR), CH_CUR_B)
            } else {
                fmt.Printf(" ")
            }

        }
        fmt.Printf("\r\n")
    }
    fmt.Printf("%s%s", curUp(g.height), colourFg(7))
}




func clearscreen() string {
    return fmt.Sprintf("%c[2J", 27)
}

func colourFg(fg int) string {
    return fmt.Sprintf("%c[38;5;%dm", 27, fg)
}

func colourBg(bg int) string {
    return fmt.Sprintf("%c[48;5;%dm", 27, bg)   
}

func curUp(count int) string {
    return fmt.Sprintf("%c[%dA", 27, count)
}

func curDown(count int) string {
    return fmt.Sprintf("%c[%dB", 27, count)
}

func curLeft(count int) string {
    return fmt.Sprintf("%c[%dD", 27, count)
}

func curRight(count int) string {
    return fmt.Sprintf("%c[%dC", 27, count)
}

func curHide() string {
    return fmt.Sprintf("%c[?25l", 27)
}

func curShow() string {
    return fmt.Sprintf("%c[?25h", 27)
}


func main() {
    if len(os.Args) != 4 {
        fmt.Printf("usage: ./sweeper <w> <h> <mines>\n")
        return
    }
    
    w, _ := strconv.Atoi(os.Args[1])
    h, _ := strconv.Atoi(os.Args[2])
    mines, _ := strconv.Atoi(os.Args[3])
    
    mine := Game{}
    mine.Init(w, h)
    mine.Reset(mines)

    oldTermState, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        fmt.Println(err)
        return
    }
    
    defer func () {
        fmt.Printf("%s%s%s", 
            curShow(),
            curDown(mine.height),
            colourFg(7),
        )
        term.Restore(int(os.Stdin.Fd()), oldTermState)
    }()

    fmt.Printf("%s\r\n", curHide())

    mine.Draw()

    for !mine.gameover {
        
        b := make([]byte, 1)
        _, err = os.Stdin.Read(b)
        if err != nil {
            break
        }

        switch uint8(b[0]) {
            case 'q':
                return
            case 'w':            
                if mine.curY > 0 {
                    mine.curY--
                    mine.moves++
                }
            case 's':
                if mine.curY < mine.height - 1 {
                    mine.curY++
                    mine.moves++
                }
            case 'a':
                if mine.curX > 0 {
                    mine.curX--
                    mine.moves++
                }
            case 'd':
                if mine.curX < mine.width - 1 {
                    mine.curX++
                    mine.moves++
                }
            case ' ':
                idx := mine.curY * mine.width + mine.curX            
                if mine.state[idx] == 'h' {
                    mine.moves++
                    if mine.field[idx] {
                        fmt.Printf("%s%sKABOOM!\r\n", curUp(1), colourFg(196))
                        mine.state[idx] = 's'
                        mine.gameover = true
                    } else {
                        reveals := mine.GetReveals(mine.curX, mine.curY)
                        for _, p := range reveals {
                            if mine.state[p.y * mine.width + p.x] == 'h' {
                                mine.state[p.y * mine.width + p.x] = 's'
                            }
                        }
                    }
                }
            case 'f':
                idx := mine.curY * mine.width + mine.curX
                mine.moves++
                if mine.state[idx] == 'f' {
                    mine.state[idx] = 'h'
                    mine.mines++
                } else if mine.state[idx] == 'h' {
                    mine.state[idx] = 'f'
                    mine.mines--
                }
        } 

        mine.Draw()
    }
}
